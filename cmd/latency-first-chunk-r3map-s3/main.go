package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go"
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
	raddr := flag.String("raddr", "http://minioadmin:minioadmin@localhost:9000?bucket=r3map&prefix=r3map", "Redis address")

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

	client, err := minio.New(u.Host, user.Username(), pw, u.Scheme == "https")
	if err != nil {
		panic(err)
	}

	bucketName := u.Query().Get("bucket")

	bucketExists, err := client.BucketExists(bucketName)
	if err != nil {
		panic(err)
	}

	if !bucketExists {
		if err := client.MakeBucket(bucketName, ""); err != nil {
			panic(err)
		}
	}

	rawBackend := &dummyBackend{
		rtt:     *rtt,
		backend: lbackend.NewS3Backend(ctx, client, bucketName, u.Query().Get("prefix"), int64(*size), false),
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
