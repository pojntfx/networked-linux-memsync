package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
	"github.com/redis/go-redis/v9"
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
	raddr := flag.String("raddr", "redis://localhost:6379/0", "Redis address")

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

	options, err := redis.ParseURL(*raddr)
	if err != nil {
		panic(err)
	}

	client := redis.NewClient(options)
	defer client.Close()

	rawBackend := &dummyBackend{
		rtt:     *rtt,
		backend: lbackend.NewRedisBackend(ctx, client, int64(*size), false),
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
