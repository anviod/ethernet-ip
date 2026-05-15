//go:build !windows

package utils

import (
	"os"
	"syscall"
)

func openMMap(path string) (*MMapFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, err
	}

	size := int(info.Size())
	data, err := syscall.Mmap(int(file.Fd()), 0, size, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		file.Close()
		return nil, err
	}

	m := &MMapFile{
		data: data,
		cleanup: func() error {
			err1 := syscall.Munmap(data)
			err2 := file.Close()
			if err1 != nil {
				return err1
			}
			return err2
		},
	}

	return m, nil
}
