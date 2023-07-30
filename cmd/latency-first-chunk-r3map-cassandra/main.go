package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/gocql/gocql"
	"github.com/pojntfx/go-nbd/pkg/backend"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
)

type dummyBackend struct {
	rtt     time.Duration
	backend backend.Backend
}

func (b *dummyBackend) ReadAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return b.backend.ReadAt(p, off)
}

func (b *dummyBackend) WriteAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return b.backend.WriteAt(p, off)
}

func (b *dummyBackend) Size() (int64, error) {
	return b.backend.Size()
}

func (b *dummyBackend) Sync() error {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return b.backend.Sync()
}

func main() {
	chunkSize := flag.Int64("chunk-size", int64(os.Getpagesize()), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")
	rtt := flag.Duration("rtt", 0, "RTT to simulate")
	raddr := flag.String("raddr", "cassandra://username:password@localhost:9042?keyspace=r3map&table=r3map&prefix=r3map", "Cassandra address")

	flag.Parse()

	devPath, err := utils.FindUnusedNBDDevice()
	if err != nil {
		panic(err)
	}

	devFile, err := os.Open(devPath)
	if err != nil {
		panic(err)
	}
	defer devFile.Close()

	u, err := url.Parse(*raddr)
	if err != nil {
		panic(err)
	}

	user := u.User
	if user == nil {
		panic("missing credentials")
	}

	pw, ok := user.Password()
	if !ok {
		panic("missing password")
	}

	cluster := gocql.NewCluster(u.Host)
	cluster.Consistency = gocql.Quorum
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: user.Username(),
		Password: pw,
	}

	if u.Scheme == "cassandrasecure" {
		cluster.SslOpts = &gocql.SslOptions{
			EnableHostVerification: true,
		}
	}

	keyspaceName := u.Query().Get("keyspace")
	{
		setupSession, err := cluster.CreateSession()
		if err != nil {
			panic(err)
		}

		if err := setupSession.Query(`create keyspace if not exists ` + keyspaceName + ` with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 }`).Exec(); err != nil {
			setupSession.Close()

			panic(err)
		}

		setupSession.Close()
	}
	cluster.Keyspace = keyspaceName

	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()

	tableName := u.Query().Get("table")

	if err := session.Query(`create table if not exists ` + tableName + ` (key blob primary key, data blob)`).Exec(); err != nil {
		panic(err)
	}

	rawBackend := &dummyBackend{
		rtt:     *rtt,
		backend: lbackend.NewCassandraBackend(session, tableName, u.Query().Get("prefix"), int64(*size), false),
	}

	mnt := mount.NewDirectFileMount(
		lbackend.NewReaderAtBackend(
			chunks.NewArbitraryReadWriterAt(
				rawBackend,
				*chunkSize,
			),
			rawBackend.Size,
			rawBackend.Sync,
			false,
		),
		devFile,

		nil,
		nil,
	)

	go func() {
		if err := mnt.Wait(); err != nil {
			panic(err)
		}
	}()

	if _, err := mnt.Open(); err != nil {
		panic(err)
	}
	defer mnt.Close()

	p := make([]byte, *chunkSize)

	beforeFirstChunk := time.Now()

	if _, err := devFile.ReadAt(p, *chunkSize); err != nil {
		panic(err)
	}

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Println(afterFirstChunk.Nanoseconds())
}
