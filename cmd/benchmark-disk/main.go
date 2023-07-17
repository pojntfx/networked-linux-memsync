package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/edsrzf/mmap-go"
)

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")

	flag.Parse()

	inputFile, err := os.CreateTemp("", "")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = inputFile.Close()

		_ = os.Remove(inputFile.Name())
	}()

	if err := inputFile.Truncate(int64(*size)); err != nil {
		panic(err)
	}

	b, err := mmap.MapRegion(inputFile, *size, mmap.RDWR, 0, 0)
	if err != nil {
		panic(err)
	}
	defer b.Unmap()

	p := make([]byte, *chunkSize)

	beforeFirstChunk := time.Now()

	copy(p, b[*chunkSize:])

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Printf("Latency till first chunk: %v\n", afterFirstChunk)

	beforeRead := time.Now()

	for i := 0; i < *size / *chunkSize; i++ {
		copy(p, b[i**chunkSize:])
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(*size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
}
