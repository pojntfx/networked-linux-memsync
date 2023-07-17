package main

import (
	"flag"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/loopholelabs/userfaultfd-go/pkg/mapper"
	"github.com/loopholelabs/userfaultfd-go/pkg/transfer"
)

type dummyReader struct {
	rtt time.Duration
	p   []byte
}

func (r *dummyReader) ReadAt(p []byte, off int64) (int, error) {
	if r.rtt > 0 {
		time.Sleep(r.rtt)
	}

	copy(r.p, p)

	return len(p), nil
}

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	socket := flag.String("socket", filepath.Join(os.TempDir(), "userfaultd.sock"), "Socket to share the file descriptor over")
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

			uffd, start, err := transfer.ReceiveUFFD(conn)
			if err != nil {
				panic(err)
			}

			p := make([]byte, *chunkSize)

			if err := mapper.Handle(uffd, start, &dummyReader{
				rtt: *rtt,
				p:   p,
			}); err != nil {
				panic(err)
			}
		}()
	}
}
