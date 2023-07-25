package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/go-nbd/pkg/client"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/utils"
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
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")
	rtt := flag.Duration("rtt", 0, "RTT to simulate")
	udev := flag.Bool("udev", false, "Whether to check for device availability with udev instead of polling")

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
			p:    make([]byte, *chunkSize),
		},
		devFile,

		nil,
		&client.Options{
			ReadyCheckUdev: *udev,
		},
	)

	go func() {
		if err := mnt.Wait(); err != nil {
			panic(err)
		}
	}()

	beforeOpen := time.Now()

	if _, err := mnt.Open(); err != nil {
		panic(err)
	}
	defer mnt.Close()

	afterOpen := time.Since(beforeOpen)

	fmt.Println(afterOpen.Nanoseconds())
}
