package fs

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"
)

var ErrLockBusy = errors.New("repository is locked")

// Lock holds an operating-system lock until Close or process termination.
// The file remains on disk: removing it could allow two owners of different inodes.
func Lock(dir, name string, ctx context.Context, wait bool) (*os.File, error) {
	file, err := os.OpenFile(fitPath(dir, name+".lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}

	for {
		if err = tryLock(file); err == nil {
			return file, nil
		}
		if !wait || !lockBusy(err) {
			if lockBusy(err) {
				err = ErrLockBusy
			}
			_ = file.Close()
			return nil, fmt.Errorf("lock %s: %w", name, err)
		}
		select {
		case <-ctx.Done():
			_ = file.Close()
			return nil, ctx.Err()
		case <-time.After(25 * time.Millisecond):
		}
	}
}
