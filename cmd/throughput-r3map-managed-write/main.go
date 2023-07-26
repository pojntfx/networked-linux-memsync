package main

import (
	"context"
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

	remote := &dummyBackend{
		size: *size,
		rtt:  *rtt,
		p:    make([]byte, *chunkSize),
	}

	local := &dummyBackend{
		size: *size,
		rtt:  0,
		p:    make([]byte, *chunkSize),
	}

	mnt := mount.NewManagedPathMount(
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

	if _, _, err := mnt.Open(); err != nil {
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
