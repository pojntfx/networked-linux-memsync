package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"time"

	"github.com/pojntfx/dudirekta/pkg/rpc"
	v1frpc "github.com/pojntfx/r3map/pkg/api/frpc/mount/v1"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
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
	laddr := flag.String("addr", ":1337", "Listen address")
	dudirekta := flag.Bool("dudirekta", true, "Whether to use Dudirekta")
	enableGrpc := flag.Bool("grpc", false, "Whether to use gRPC instead of Dudirekta")
	enableFrpc := flag.Bool("frpc", false, "Whether to use fRPC instead of Dudirekta")
	verbose := flag.Bool("verbose", false, "Whether to enable verbose logging")

	size := flag.Int("size", os.Getpagesize()*1024*10, "Amount of bytes to read")
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	maxChunkSize := flag.Int64("max-chunk-size", services.MaxChunkSize, "Maximum chunk size to support")

	rtt := flag.Duration("rtt", 0, "RTT to simulate")

	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	b := &dummyBackend{
		size: *size,
		rtt:  *rtt,
		p:    make([]byte, *chunkSize),
	}

	var (
		svc  = services.NewBackend(b, *verbose, *maxChunkSize)
		errs = make(chan error)
	)
	if *enableGrpc {
		server := grpc.NewServer()

		v1proto.RegisterBackendServer(server, services.NewBackendGrpc(svc))

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Listening on", lis.Addr())

		go func() {
			if err := server.Serve(lis); err != nil {
				if !utils.IsClosedErr(err) {
					errs <- err
				}

				return
			}
		}()
	} else if *enableFrpc {
		server, err := v1frpc.NewServer(services.NewBackendFrpc(svc), nil, nil)
		if err != nil {
			panic(err)
		}

		log.Println("Listening on", *laddr)

		go func() {
			if err := server.Start(*laddr); err != nil {
				if !utils.IsClosedErr(err) {
					errs <- err
				}

				return
			}
		}()
	} else if *dudirekta {
		clients := 0
		registry := rpc.NewRegistry(
			svc,
			struct{}{},

			time.Second*10,
			ctx,
			&rpc.Options{
				ResponseBufferLen: rpc.DefaultResponseBufferLen,
				OnClientConnect: func(remoteID string) {
					clients++

					log.Printf("%v clients connected", clients)
				},
				OnClientDisconnect: func(remoteID string) {
					clients--

					log.Printf("%v clients connected", clients)
				},
			},
		)

		lis, err := net.Listen("tcp", *laddr)
		if err != nil {
			panic(err)
		}
		defer lis.Close()

		log.Println("Listening on", lis.Addr())

		go func() {
			for {
				conn, err := lis.Accept()
				if err != nil {
					if !utils.IsClosedErr(err) {
						log.Println("could not accept connection, continuing:", err)
					}

					continue
				}

				go func() {
					defer func() {
						_ = conn.Close()

						if err := recover(); err != nil {
							if !utils.IsClosedErr(err.(error)) {
								log.Printf("Client disconnected with error: %v", err)
							}
						}
					}()

					if err := registry.Link(conn); err != nil {
						panic(err)
					}
				}()
			}
		}()
	}

	for err := range errs {
		if err != nil {
			panic(err)
		}
	}
}
