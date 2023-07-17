package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024*1024, "Amount of bytes to read")

	flag.Parse()

	input := make([]byte, *chunkSize)
	output := make([]byte, *chunkSize)

	beforeFirstTwoChunks := time.Now()

	for i := 0; i < 2; i++ {
		copy(output, input)
	}

	afterFirstTwoChunks := time.Since(beforeFirstTwoChunks)

	fmt.Printf("Latency till first two chunks: %v\n", afterFirstTwoChunks)

	beforeRead := time.Now()

	for i := 0; i < *size / *chunkSize; i++ {
		copy(output, input)
	}

	afterRead := time.Since(beforeRead)

	throughputMB := float64(*size) / (1024 * 1024) / afterRead.Seconds()

	fmt.Printf("Read throughput: %.2f MB/s (%.2f Mb/s)\n", throughputMB, throughputMB*8)
}
