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

	deviceFile, err := os.OpenFile(devPath, os.O_RDWR, os.ModePerm)
	if err != nil {
		panic(err)
	}
	defer deviceFile.Close()

	p := make([]byte, *chunkSize)

	beforeFirstChunk := time.Now()

	if _, err := deviceFile.ReadAt(p, int64(*chunkSize)); err != nil {
		panic(err)
	}

	afterFirstChunk := time.Since(beforeFirstChunk)

	fmt.Println(afterFirstChunk.Nanoseconds())
}
