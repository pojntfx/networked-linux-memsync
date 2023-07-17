package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
)

type dummyBackend struct {
	size int
	rtt  time.Duration
}

func (b *dummyBackend) ReadAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return len(p), nil
}

func (b *dummyBackend) WriteAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return len(p), nil
}

func (b *dummyBackend) Size() (int64, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return int64(b.size), nil
}

func (b *dummyBackend) Sync() error {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return nil
}

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
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

	mnt := mount.NewDirectFileMount(
		&dummyBackend{
			size: *size,
			rtt:  *rtt,
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

	beforeOpen := time.Now()

	b, err := mnt.Open()
	if err != nil {
		panic(err)
	}

	afterOpen := time.Since(beforeOpen)

	fmt.Printf("Open: %v\n", afterOpen)

	defer mnt.Close()

	beforeFirstTwoChunks := time.Now()

	p := make([]byte, *chunkSize)

	for i := 0; i < 2; i++ {
		if _, err := b.ReadAt(p, int64(i**chunkSize)); err != nil {
			panic(err)
		}
	}

	afterFirstTwoChunks := time.Since(beforeFirstTwoChunks)

	fmt.Printf("Latency till first two chunks: %v\n", afterFirstTwoChunks)

	beforeRead := time.Now()

	for i := 0; i < *size / *chunkSize; i++ {
		if _, err := b.ReadAt(p, int64(i**chunkSize)); err != nil {
			panic(err)
		}
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(*size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
}
