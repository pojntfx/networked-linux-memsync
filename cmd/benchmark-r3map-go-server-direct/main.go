package main

import (
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
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")
	socket := flag.String("socket", filepath.Join(os.TempDir(), "r3map.sock"), "Socket to share the file descriptor over")
	rtt := flag.Duration("rtt", 0, "RTT to simulate")

	flag.Parse()

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

			beforeOpen := time.Now()

			if err := mnt.Open(); err != nil {
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
