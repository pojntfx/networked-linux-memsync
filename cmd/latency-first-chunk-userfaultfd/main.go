package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/loopholelabs/userfaultfd-go/pkg/mapper"
	"github.com/loopholelabs/userfaultfd-go/pkg/transfer"
)

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")
	socket := flag.String("socket", filepath.Join(os.TempDir(), "userfaultd.sock"), "Socket to share the file descriptor over")

	flag.Parse()

	addr, err := net.ResolveUnixAddr("unix", *socket)
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		panic(err)
	}

	log.Println("Connected to", conn.RemoteAddr())

	b, uffd, start, err := mapper.Register(*size)
	if err != nil {
		panic(err)
	}

	if err := transfer.SendUFFD(conn, uffd, start); err != nil {
		panic(err)
	}

	p := make([]byte, *chunkSize)

	beforeFirstChunk := time.Now()

	copy(p, b[*chunkSize:])

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Println(afterFirstChunk.Nanoseconds())
}
