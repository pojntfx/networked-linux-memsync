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
}

func (r *dummyReader) ReadAt(p []byte, off int64) (int, error) {
	if r.rtt > 0 {
		time.Sleep(r.rtt)
	}

	return len(p), nil
}

func main() {
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

			if err := mapper.Handle(uffd, start, &dummyReader{*rtt}); err != nil {
				panic(err)
			}
		}()
	}
}
