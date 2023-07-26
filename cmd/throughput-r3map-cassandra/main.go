package main

import (
	"context"
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

type rttBackend struct {
	size int
	rtt  time.Duration
	p    []byte
}

func (b *rttBackend) ReadAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	copy(b.p, p)

	return len(p), nil
}

func (b *rttBackend) WriteAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	copy(p, b.p)

	return len(p), nil
}

func (b *rttBackend) Size() (int64, error) {
	return int64(b.size), nil
}

func (b *rttBackend) Sync() error {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return nil
}

func main() {
	chunkSize := flag.Int64("chunk-size", int64(os.Getpagesize()), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*100, "Amount of bytes to read")
	rtt := flag.Duration("rtt", 0, "RTT to simulate")
	raddr := flag.String("raddr", "cassandra://username:password@localhost:9042?keyspace=r3map&table=r3map&prefix=r3map", "Cassandra address")

	managed := flag.Bool("managed", false, "Whether to use a managed mount instead of a direct mount")

	pullWorkers := flag.Int64("pull-workers", 4096, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	pullFirst := flag.Bool("pull-first", false, "Whether to completely pull from the remote before opening")

	pushWorkers := flag.Int64("push-workers", 4096, "Push workers to launch in the background; pass in 0 to disable push")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval after which to push changed chunks to the remote")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

	var deviceFile *os.File
	rawBackend := lbackend.NewCassandraBackend(session, tableName, u.Query().Get("prefix"), int64(*size), false)

	rawLocalBackend := &dummyBackend{
		rtt: *rtt,
		backend: lbackend.NewReaderAtBackend(
			chunks.NewArbitraryReadWriterAt(
				rawBackend,
				*chunkSize,
			),
			rawBackend.Size,
			rawBackend.Sync,
			false,
		),
	}

	if *managed {
		remote := &dummyBackend{
			rtt:     *rtt,
			backend: rawLocalBackend,
		}

		local := &rttBackend{
			size: *size,
			rtt:  0,
			p:    make([]byte, *chunkSize),
		}

		mnt := mount.NewManagedFileMount(
			ctx,

			remote,
			local,

			&mount.ManagedMountOptions{
				ChunkSize: *chunkSize,

				PullWorkers: *pullWorkers,
				PullFirst:   *pullFirst,

				PushWorkers:  *pushWorkers,
				PushInterval: *pushInterval,
			},
			nil,

			nil,
			nil,
		)

		go func() {
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		deviceFile, err = mnt.Open()
		if err != nil {
			panic(err)
		}
		defer mnt.Close()
	} else {
		mnt := mount.NewDirectFileMount(
			&dummyBackend{
				rtt:     *rtt,
				backend: rawLocalBackend,
			},
			devFile,

			nil,
			nil,
		)

		go func() {
			if err := mnt.Wait(); err != nil {
				panic(err)
			}
		}()

		deviceFile, err = mnt.Open()
		if err != nil {
			panic(err)
		}
		defer mnt.Close()
	}

	p := make([]byte, *chunkSize)

	beforeRead := time.Now()

	for i := int64(0); i < int64(*size) / *chunkSize; i++ {
		if _, err := deviceFile.ReadAt(p, i**chunkSize); err != nil {
			panic(err)
		}
	}

	afterRead := time.Since(beforeRead)

	fmt.Println(float64(*size) / (1024 * 1024) / afterRead.Seconds())
}
