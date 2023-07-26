package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
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

	file, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(file.Name())

	if err := file.Truncate(int64(*size)); err != nil {
		panic(err)
	}

	var deviceFile *os.File
	rawLocalBackend := &dummyBackend{
		rtt:     *rtt,
		backend: backend.NewFileBackend(file),
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
			rawLocalBackend,
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
