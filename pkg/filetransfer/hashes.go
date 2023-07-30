package filetransfer

import (
	"context"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"sync"

	"hash/crc32"

	"golang.org/x/sync/semaphore"
)

func GetHashesForBlocks(
	parallel int64,
	file string,
	blocksize int64,
) ([]string, int64, error) {
	s, err := os.Stat(file)
	if err != nil {
		return []string{}, 0, nil
	}

	var wg sync.WaitGroup
	blocks := int64(math.Ceil(float64(s.Size()) / float64((blocksize))))
	hashes := make([]string, blocks)
	var hashesLock sync.Mutex

	wg.Add(int(blocks))

	lock := semaphore.NewWeighted(parallel)
	calculateHash := func(j int64) {
		_ = lock.Acquire(context.Background(), 1)

		defer func() {
			lock.Release(1)

			wg.Done()
		}()

		var checkFile *os.File
		checkFile, err = os.Open(file)
		if err != nil {
			return
		}

		hash := crc32.NewIEEE()

		if _, err = io.CopyN(hash, io.NewSectionReader(checkFile, j*(blocksize), blocksize), blocksize); err != nil && !errors.Is(err, io.EOF) {
			return
		}

		if err = checkFile.Close(); err != nil {
			return
		}

		hashesLock.Lock()
		hashes[j] = hex.EncodeToString(hash.Sum(nil))
		hashesLock.Unlock()
	}

	for i := int64(0); i < blocks; i++ {
		j := i

		go calculateHash(j)
	}
	wg.Wait()

	return hashes, (blocksize * blocks) - s.Size(), err
}
