package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

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
	socket := flag.String("socket", filepath.Join(os.TempDir(), "r3map.sock"), "Socket to share the file descriptor over")
	rtt := flag.Duration("rtt", 0, "RTT to simulate")

	pullWorkers := flag.Int64("pull-workers", 4096, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	pullFirst := flag.Bool("pull-first", false, "Whether to completely pull from the remote before opening")

	pushWorkers := flag.Int64("push-workers", 4096, "Push workers to launch in the background; pass in 0 to disable push")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval after which to push changed chunks to the remote")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_ = os.Remove(*socket)

	addr, err := net.ResolveUnixAddr("unix", *socket)
	if err != nil {
		panic(err)
	}

	lis, err := net.ListenUnix("unix", addr)
	if err != nil {
		panic(err)
	}

	log.Println("Listening on", addr.String())

	for {
		conn, err := lis.AcceptUnix()
		if err != nil {
			panic(err)
		}

		go func() {
			defer func() {
				if err := recover(); err != nil {
					log.Println("Could not handle connection, stopping:", err)
				}

				_ = conn.Close()
			}()

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

			beforeOpen := time.Now()

			if _, _, err := mnt.Open(); err != nil {
				panic(err)
			}
			defer mnt.Close()

			afterOpen := time.Since(beforeOpen)

			fmt.Printf("Open: %v\n", afterOpen)

			if err := json.NewEncoder(conn).Encode(devPath); err != nil {
				panic(err)
			}

			if _, err := conn.Read(make([]byte, 1)); err != nil && !utils.IsClosedErr(err) {
				panic(err)
			}
		}()
	}
}
