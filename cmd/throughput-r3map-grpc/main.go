package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/pojntfx/go-nbd/pkg/backend"
	v1proto "github.com/pojntfx/r3map/pkg/api/proto/mount/v1"
	lbackend "github.com/pojntfx/r3map/pkg/backend"
	"github.com/pojntfx/r3map/pkg/chunks"
	"github.com/pojntfx/r3map/pkg/mount"
	"github.com/pojntfx/r3map/pkg/services"
	"github.com/pojntfx/r3map/pkg/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	size := flag.Int("size", os.Getpagesize()*1024*10, "Amount of bytes to read")
	raddr := flag.String("raddr", "localhost:1337", "Remote address")

	managed := flag.Bool("managed", false, "Whether to use a managed mount instead of a direct mount")

	pullWorkers := flag.Int64("pull-workers", 4096, "Pull workers to launch in the background; pass in 0 to disable preemptive pull")
	pullFirst := flag.Bool("pull-first", false, "Whether to completely pull from the remote before opening")

	pushWorkers := flag.Int64("push-workers", 4096, "Push workers to launch in the background; pass in 0 to disable push")
	pushInterval := flag.Duration("push-interval", 5*time.Minute, "Interval after which to push changed chunks to the remote")

	rtt := flag.Duration("rtt", 0, "RTT to simulate")

	chunking := flag.Bool("chunking", false, "Whether the backend requires to be interfaced with in fixed chunks")
	nestedChunkSize := flag.Int64("nested-chunk-size", int64(os.Getpagesize()/2), "Nested chunk size to test chunking with")

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

	conn, err := grpc.Dial(*raddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	client := v1proto.NewBackendClient(conn)

	peer := &services.BackendRemote{
		ReadAt: func(ctx context.Context, length int, off int64) (r services.ReadAtResponse, err error) {
			res, err := client.ReadAt(ctx, &v1proto.ReadAtArgs{
				Length: int32(length),
				Off:    off,
			})
			if err != nil {
				return services.ReadAtResponse{}, err
			}

			return services.ReadAtResponse{
				N: int(res.GetN()),
				P: res.GetP(),
			}, err
		},
		WriteAt: func(context context.Context, p []byte, off int64) (n int, err error) {
			res, err := client.WriteAt(ctx, &v1proto.WriteAtArgs{
				Off: off,
				P:   p,
			})
			if err != nil {
				return 0, err
			}

			return int(res.GetLength()), nil
		},
		Size: func(context context.Context) (int64, error) {
			res, err := client.Size(ctx, &v1proto.SizeArgs{})
			if err != nil {
				return 0, err
			}

			return res.GetSize(), nil
		},
		Sync: func(context context.Context) error {
			if _, err := client.Sync(ctx, &v1proto.SyncArgs{}); err != nil {
				return err
			}

			return nil
		},
	}

	var b backend.Backend
	b = lbackend.NewRPCBackend(ctx, peer, false)

	if *chunking {
		b = lbackend.NewReaderAtBackend(
			chunks.NewArbitraryReadWriterAt(
				b,
				*nestedChunkSize,
			),
			b.Size,
			b.Sync,
			false,
		)
	}

	rawRemoteBackend := &dummyBackend{
		rtt:     *rtt,
		backend: b,
	}

	var deviceFile *os.File
	if *managed {
		local := &rttBackend{
			size: *size,
			rtt:  0,
			p:    make([]byte, *chunkSize),
		}

		mnt := mount.NewManagedFileMount(
			ctx,

			rawRemoteBackend,
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
			rawRemoteBackend,
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
