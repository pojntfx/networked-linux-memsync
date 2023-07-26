package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
	"golang.org/x/sys/unix"
)

type dummyBackend struct {
	size int
	rtt  time.Duration
	p    []byte
}

func (b *dummyBackend) ReadAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	copy(b.p, p)

	return len(p), nil
}

func (b *dummyBackend) WriteAt(p []byte, off int64) (int, error) {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	copy(p, b.p)

	return len(p), nil
}

func (b *dummyBackend) Size() (int64, error) {
	return int64(b.size), nil
}

func (b *dummyBackend) Sync() error {
	if b.rtt > 0 {
		time.Sleep(b.rtt)
	}

	return nil
}

func main() {
	chunkSize := flag.Int64("chunk-size", int64(os.Getpagesize()), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024, "Amount of bytes to write")
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

	mnt := mount.NewDirectPathMount(
		&dummyBackend{
			size: *size,
			rtt:  *rtt,
			p:    make([]byte, *chunkSize),
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

	if err := mnt.Open(); err != nil {
		panic(err)
	}
	defer mnt.Close()

	f, err := os.OpenFile(devPath, os.O_RDWR|unix.O_DIRECT, 0)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	source := &dummyBackend{
		size: *size,
		rtt:  0,
		p:    make([]byte, *chunkSize),
	}

	p := make([]byte, *chunkSize)

	beforeWrite := time.Now()

	for i := int64(0); i < int64(*size) / *chunkSize; i++ {
		if _, err := source.ReadAt(p, i**chunkSize); err != nil {
			panic(err)
		}

		if _, err := f.WriteAt(p, i**chunkSize); err != nil {
			panic(err)
		}
	}

	afterRead := time.Since(beforeWrite)

	fmt.Println(float64(*size) / (1024 * 1024) / afterRead.Seconds())
}
