package fs

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func MakeWatcher(paths ...string) (int, error) {
	if fd, err := syscall.InotifyInit(); err != nil {
		return -1, err
	} else {
		syscall.Close(fd)
		for _, path := range paths {
			var watchmode uint32
			if stat, err := os.Stat(path); err != nil {
				syscall.Close(fd)
				return -1, err
			} else if stat.IsDir() {
				watchmode = unix.IN_CREATE | unix.IN_MODIFY | unix.IN_DELETE
			} else {
				watchmode = unix.IN_CLOSE_WRITE
			}

			if _, err := syscall.InotifyAddWatch(fd, path, watchmode); err != nil {
				syscall.Close(fd)
				return -1, err
			}
		}

		return fd, nil
	}
}

func Wait(fd int) error {
	var dummy = make([]byte, 0, 100)
	_, err := unix.Read(fd, dummy)
	return err
}
