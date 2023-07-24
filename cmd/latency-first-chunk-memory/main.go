package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

func main() {
	chunkSize := flag.Int("chunk-size", os.Getpagesize(), "Chunk size to use")
	size := flag.Int("size", os.Getpagesize()*1024, "Amount of bytes to allocate")

	flag.Parse()

	b, err := unix.Mmap(
		-1,
		0,
		*size,
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_POPULATE,
	)
	if err != nil {
		panic(err)
	}
	defer unix.Munmap(b)

	p := make([]byte, *chunkSize)

	beforeFirstChunk := time.Now()

	copy(p, b[*chunkSize:])

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Println(afterFirstChunk.Nanoseconds())
}
