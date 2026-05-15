//go:build windows

package utils

import (
	"os"
)

func openMMap(path string) (*MMapFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		file.Close()
		return nil, err
	}

	m := &MMapFile{
		data: data,
		cleanup: func() error {
			return file.Close()
		},
	}

	return m, nil
}
