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
)

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")
	socket := flag.String("socket", filepath.Join(os.TempDir(), "r3map.sock"), "Socket to share the file descriptor over")

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

	devPath := ""
	if json.NewDecoder(conn).Decode(&devPath); err != nil {
		panic(err)
	}

	b, err := os.OpenFile(devPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer b.Close()

	beforeFirstTwoChunks := time.Now()

	p := make([]byte, *chunkSize)

	for i := 0; i < 2; i++ {
		if _, err := b.ReadAt(p, int64(i**chunkSize)); err != nil {
			panic(err)
		}
	}

	afterFirstTwoChunks := time.Since(beforeFirstTwoChunks)

	fmt.Printf("Latency till first two chunks: %v\n", afterFirstTwoChunks)

	beforeRead := time.Now()

	for i := 0; i < *size / *chunkSize; i++ {
		if _, err := b.ReadAt(p, int64(i**chunkSize)); err != nil {
			panic(err)
		}
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(*size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
}
