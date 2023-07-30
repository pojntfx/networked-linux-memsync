package filetransfer

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/pojntfx/networked-linux-memsync/pkg/utils"
)

// {
// 		wg.Add(1)

// 		go func() {
// 			if err := filetransfer.ReceiveFileLoop(
// 				ctx,
// 				*hashParallel,
// 				d,
// 				*syncerRaddr,
// 				func(src string) bool {
// 					for _, hv := range hypervisors {
// 						// Ignore local, running virtual machines to prevent sync loopback
// 						if !hv.computePaused && strings.HasPrefix(src, filepath.Join("/", vmDirName, hv.id)) {
// 							return false
// 						}
// 					}

// 					return true
// 				},
// 				func(src string) string {
// 					return filepath.Join(*firecrackerWorkdir, src)
// 				},
// 				*blocksize,
// 				func(src string) bool {
// 					if *rootfsWritable {
// 						return false
// 					}

// 					// Only sync rootfs once if its mounted read-only, as calculating hashes can be very slow
// 					return strings.HasSuffix(src, rootfsName)
// 				},
// 				*verbose,
// 			); err != nil {
// 				panic(err)
// 			}
// 		}()
// 	}

type nopReader struct{}

func (r nopReader) Read(p []byte) (n int, err error) {
	n = copy(p, make([]byte, len(p)))

	return n, nil
}

func ReceiveFile(
	parallel int64,
	file *os.File,
	path string,
	blocksize int64,
	conn io.ReadWriter,
	verbose bool,
) error {
	beforeHash := time.Now()

	localHashes, _, err := GetHashesForBlocks(parallel, path, blocksize)
	if err != nil {
		return err
	}

	if verbose {
		log.Printf("Calculated hashes for %v blocks in %v", len(localHashes), time.Since(beforeHash))
	}

	if err := utils.EncodeJSONFixedLength(conn, localHashes); err != nil {
		return err
	}

	blocksToFetch := []int64{}
	if err := utils.DecodeJSONFixedLength(conn, &blocksToFetch); err != nil {
		return err
	}

	cutoff := int64(0)
	if err := utils.DecodeJSONFixedLength(conn, &cutoff); err != nil {
		return err
	}

	if cutoff == -1 {
		if verbose {
			log.Println("Truncating file")
		}

		if err := file.Truncate(0); err != nil {
			return err
		}

		return nil
	}

	if verbose {
		log.Printf("Fetching changed blocks %v with cutoff %v", blocksToFetch, cutoff)
	}

	beforeDownload := time.Now()

	if len(blocksToFetch) > 0 {
		s, err := os.Stat(path)
		if err != nil {
			return err
		}

		newSize := (((blocksToFetch[len(blocksToFetch)-1] + 1) * blocksize) - cutoff)
		diff := s.Size() - newSize

		if diff > 0 {
			// If the file on the server got smaller, truncate the local file accordingly
			if err := file.Truncate(newSize); err != nil {
				return err
			}
		} else {
			// If the file on the server grew, grow the local file accordingly
			if _, err := file.Seek(0, 2); err != nil {
				return err
			}

			if _, err := io.CopyN(file, nopReader{}, -diff); err != nil {
				return err
			}
		}
	}

	n := int64(0)
	for i, blockToFetch := range blocksToFetch {
		backset := int64(0)
		if i == len(blocksToFetch)-1 {
			backset = cutoff
		}

		b := make([]byte, blocksize-backset)
		if _, err := io.ReadFull(conn, b); err != nil {
			return err
		}

		m, err := file.WriteAt(b, blockToFetch*(blocksize))
		if err != nil {
			return err
		}

		n += int64(m)
	}

	if verbose {
		log.Printf("Received %v bytes in %v", n, time.Since(beforeDownload))
	}

	return nil
}

func ReceiveFileLoop(
	ctx context.Context,
	parallel int64,
	d tls.Dialer,
	syncerRaddr string,
	getAccept func(src string) bool,
	getDstPath func(src string) string,
	blocksize int64,
	once func(src string) bool,
	verbose bool,
) error {
	syncerConn, err := d.DialContext(ctx, "tcp", syncerRaddr)
	if err != nil {
		panic(err)
	}
	defer syncerConn.Close()

	go func() {
		<-ctx.Done()

		_ = syncerConn.Close()
	}()

	if verbose {
		log.Println("Sharer destination connected to syncer", syncerConn.RemoteAddr())
	}

	if err := utils.EncodeJSONFixedLength(syncerConn, "dst-control"); err != nil {
		if utils.IsClosedErr(err) {
			return nil
		}

		panic(err)
	}

	errs := make(chan error)
	go func() {
		for {
			file := ""
			if err := utils.DecodeJSONFixedLength(syncerConn, &file); err != nil {
				if utils.IsClosedErr(err) {
					errs <- nil

					return
				}

				errs <- err

				return
			}

			if !getAccept(file) {
				continue
			}

			go func() {
				dataConn, err := d.DialContext(ctx, "tcp", syncerRaddr)
				if err != nil {
					errs <- err

					return
				}
				defer dataConn.Close()

				dst := getDstPath(file)
				if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
					errs <- err

					return
				}

				f, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE, os.ModePerm)
				if err != nil {
					errs <- err

					return
				}
				defer f.Close()

				if err := utils.EncodeJSONFixedLength(dataConn, "dst"); err != nil {
					if utils.IsClosedErr(err) {
						return
					}

					errs <- err

					return
				}

				if err := utils.EncodeJSONFixedLength(dataConn, file); err != nil {
					if utils.IsClosedErr(err) {
						return
					}

					errs <- err

					return
				}

				for {
					if err := ReceiveFile(parallel, f, dst, blocksize, dataConn, verbose); err != nil {
						if utils.IsClosedErr(err) {
							return
						}

						errs <- err

						return
					}

					if once(file) {
						return
					}
				}
			}()
		}
	}()

	return <-errs
}
