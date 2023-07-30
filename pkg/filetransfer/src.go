package filetransfer

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"os"
	"time"

	"github.com/pojntfx/networked-linux-memsync/pkg/utils"
)

// if err := filetransfer.AdvertiseFile(
// 	ctx,
// 	*hashParallel,
// 	d,
// 	*syncerRaddr,
// 	filepath.Join(workdir, file),
// 	strings.TrimPrefix(filepath.Join(workdir, file), *firecrackerWorkdir),
// 	*blocksize,
// 	func() {
// 		if !*rootfsWritable && file == rootfsName {
// 			return
// 		}

// 		broadcaster <- struct{}{}
// 	},
// 	*syncerPollingDuration,
// 	*verbose,
// ); err != nil {
// 	panic(err)
// }

func SendFile(
	parallel int64,
	file *os.File,
	path string,
	blocksize int64,
	conn io.ReadWriter,
	verbose bool,
) (int64, error) {
	if _, err := file.Seek(0, 0); err != nil {
		return -1, err
	}

	remoteHashes := []string{}
	if err := utils.DecodeJSONFixedLength(conn, &remoteHashes); err != nil {
		return -1, err
	}

	beforeHash := time.Now()

	localHashes, cutoff, err := GetHashesForBlocks(parallel, path, blocksize)
	if err != nil {
		return -1, err
	}

	if verbose {
		log.Printf("Calculated hashes for %v blocks in %v with cutoff %v", len(localHashes), time.Since(beforeHash), cutoff)
	}

	blocksToSend := []int64{}
	for i, localHash := range localHashes {
		j := int64(i)

		if len(remoteHashes) <= i {
			blocksToSend = append(blocksToSend, j)

			continue
		}

		if localHash != remoteHashes[i] {
			blocksToSend = append(blocksToSend, j)

			continue
		}
	}

	if err := utils.EncodeJSONFixedLength(conn, blocksToSend); err != nil {
		return -1, err
	}

	if len(remoteHashes) > 0 && len(localHashes) <= 0 {
		if err := utils.EncodeJSONFixedLength(conn, -1); err != nil {
			return -1, err
		}

		if verbose {
			log.Println("Sending request to truncate file")
		}

		return 0, nil
	}

	if err := utils.EncodeJSONFixedLength(conn, cutoff); err != nil {
		return -1, err
	}

	if verbose {
		log.Printf("Sending changed blocks %v with cutoff %v", blocksToSend, cutoff)
	}

	before := time.Now()

	n := int64(0)
	for i, blockToSend := range blocksToSend {
		backset := int64(0)
		if i == len(blocksToSend)-1 {
			backset = cutoff
		}

		b := make([]byte, blocksize-backset)
		if _, err := file.ReadAt(b, blockToSend*(blocksize)); err != nil {
			return -1, err
		}

		m, err := conn.Write(b)
		if err != nil {
			return -1, err
		}

		n += int64(m)
	}

	if verbose {
		log.Printf("Sent %v bytes in %v", n, time.Since(before))
	}

	return n, nil
}

func AdvertiseFile(
	ctx context.Context,
	parallel int64,
	d tls.Dialer,
	syncerRaddr string,
	src string,
	advertisedSrc string,
	blocksize int64,
	onDone func(),
	pollDuration time.Duration,
	verbose bool,
) error {
	syncerConn, err := d.DialContext(ctx, "tcp", syncerRaddr)
	if err != nil {
		return err
	}
	defer syncerConn.Close()

	go func() {
		<-ctx.Done()

		_ = syncerConn.Close()
	}()

	if verbose {
		log.Println("Sharer source connected to syncer", syncerConn.RemoteAddr())
	}

	if err := utils.EncodeJSONFixedLength(syncerConn, "src-control"); err != nil {
		if utils.IsClosedErr(err) {
			return nil
		}

		return err
	}

	if err := utils.EncodeJSONFixedLength(syncerConn, advertisedSrc); err != nil {
		if utils.IsClosedErr(err) {
			return nil
		}

		return err
	}

	errs := make(chan error)
	go func() {
		for {
			id := ""
			if err := utils.DecodeJSONFixedLength(syncerConn, &id); err != nil {
				if utils.IsClosedErr(err) {
					errs <- nil

					return
				}

				errs <- err

				return
			}

			go func() {
				dataConn, err := d.DialContext(ctx, "tcp", syncerRaddr)
				if err != nil {
					errs <- err

					return
				}
				defer dataConn.Close()

				syncStopped := false
				go func() {
					<-ctx.Done()

					syncStopped = true
				}()

				if verbose {
					log.Println("Data connected to", syncerConn.RemoteAddr())
				}

				f, err := os.OpenFile(src, os.O_RDONLY, os.ModePerm)
				if err != nil {
					errs <- err

					return
				}
				defer f.Close()

				if err := utils.EncodeJSONFixedLength(dataConn, "src-data"); err != nil {
					if utils.IsClosedErr(err) {
						errs <- nil

						return
					}

					errs <- err

					return
				}

				if err := utils.EncodeJSONFixedLength(dataConn, id); err != nil {
					if utils.IsClosedErr(err) {
						errs <- nil

						return
					}

					errs <- err

					return
				}

				done := false
				for {
					if done {
						_ = dataConn.Close()

						if onDone != nil {
							onDone()
						}

						return
					}

					if syncStopped {
						// Ensure that file has been synced by finishing the last sync
						if _, err := SendFile(parallel, f, src, blocksize, dataConn, verbose); err != nil {
							if utils.IsClosedErr(err) {
								errs <- nil

								return
							}

							errs <- err

							return
						}

						done = true
					}

					if _, err := SendFile(parallel, f, src, blocksize, dataConn, verbose); err != nil {
						if utils.IsClosedErr(err) {
							errs <- nil

							return
						}

						errs <- err

						return
					}

					time.Sleep(pollDuration)
				}
			}()
		}
	}()

	return <-errs
}
