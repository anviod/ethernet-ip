package utils

import (
	"os"
)

type MMapFile struct {
	data    []byte
	cleanup func() error
}

func OpenMMap(path string) (*MMapFile, error) {
	return openMMap(path)
}

func (m *MMapFile) Data() []byte {
	if m == nil {
		return nil
	}
	return m.data
}

func (m *MMapFile) Close() error {
	if m == nil || m.cleanup == nil {
		return nil
	}
	return m.cleanup()
}

func WriteFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0644)
}
