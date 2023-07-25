package main

import (
	"flag"
	"fmt"
	"os"
	"time"

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

	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	rawBackend := &dummyBackend{
		rtt:     *rtt,
		backend: lbackend.NewDirectoryBackend(dir, int64(*size), *chunkSize, 4096, false),
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

	if _, err := devFile.ReadAt(p, 0); err != nil {
		panic(err)
	}

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Println(afterFirstChunk.Nanoseconds())
}
