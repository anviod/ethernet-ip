package utils

import (
	"encoding/binary"
)

// CopyUint16LE copies uint16 values in little-endian format
// Note: Uses safe byte-by-byte copy to ensure compatibility with all architectures
func CopyUint16LE(dst []byte, src []uint16) {
	if len(dst) < len(src)*2 {
		panic("destination buffer too small")
	}

	for i, v := range src {
		offset := i * 2
		binary.LittleEndian.PutUint16(dst[offset:offset+2], v)
	}
}

// CopyUint32LE copies uint32 values in little-endian format
// Note: Uses safe byte-by-byte copy to ensure compatibility with all architectures
func CopyUint32LE(dst []byte, src []uint32) {
	if len(dst) < len(src)*4 {
		panic("destination buffer too small")
	}

	for i, v := range src {
		offset := i * 4
		binary.LittleEndian.PutUint32(dst[offset:offset+4], v)
	}
}

// FastMemset fills a byte slice with a value using optimized copying
func FastMemset(dst []byte, value byte) {
	if len(dst) == 0 {
		return
	}

	// Set first byte
	dst[0] = value

	// Use doubling copy strategy for better performance
	for i := 1; i < len(dst); i *= 2 {
		copy(dst[i:], dst[:i])
	}
}

// VectorizedXOR performs XOR operation on byte slices
func VectorizedXOR(dst, src1, src2 []byte) {
	if len(dst) != len(src1) || len(src1) != len(src2) {
		panic("slice lengths must be equal")
	}

	// Process in 8-byte chunks with alignment check
	for i := 0; i < len(dst); i += 8 {
		if i+8 <= len(dst) {
			// Check alignment for 64-bit operations
			// Use byte-by-byte XOR for compatibility with all architectures
			for j := 0; j < 8; j++ {
				dst[i+j] = src1[i+j] ^ src2[i+j]
			}
		} else {
			for j := i; j < len(dst); j++ {
				dst[j] = src1[j] ^ src2[j]
			}
		}
	}
}