package utils

import (
	"unsafe"
)

// SIMD operations for encoding/decoding
// Note: This is a simplified implementation. For production use,
// consider using assembly or external libraries like gonum.

// CopyUint16LE copies uint16 values in little-endian format using SIMD-like operations
func CopyUint16LE(dst []byte, src []uint16) {
	if len(dst) < len(src)*2 {
		panic("destination buffer too small")
	}

	// Use unsafe.Pointer for direct memory access (simulating SIMD)
	dstPtr := unsafe.Pointer(&dst[0])
	srcPtr := unsafe.Pointer(&src[0])

	// Copy in chunks of 8 bytes (4 uint16 values) for better performance
	for i := 0; i < len(src); i += 4 {
		if i+4 <= len(src) {
			// Copy 4 uint16 values at once (64 bits)
			*(*uint64)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i*2))) =
				*(*uint64)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i*2)))
		} else {
			// Handle remaining values
			for j := i; j < len(src); j++ {
				*(*uint16)(unsafe.Pointer(uintptr(dstPtr) + uintptr(j*2))) =
					*(*uint16)(unsafe.Pointer(uintptr(srcPtr) + uintptr(j*2)))
			}
		}
	}
}

// CopyUint32LE copies uint32 values in little-endian format
func CopyUint32LE(dst []byte, src []uint32) {
	if len(dst) < len(src)*4 {
		panic("destination buffer too small")
	}

	dstPtr := unsafe.Pointer(&dst[0])
	srcPtr := unsafe.Pointer(&src[0])

	// Copy in chunks of 16 bytes (4 uint32 values)
	for i := 0; i < len(src); i += 4 {
		if i+4 <= len(src) {
			// Copy 4 uint32 values at once (128 bits)
			*(*uint64)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i*4))) =
				*(*uint64)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i*4)))
			*(*uint64)(unsafe.Pointer(uintptr(dstPtr) + uintptr(i*4+8))) =
				*(*uint64)(unsafe.Pointer(uintptr(srcPtr) + uintptr(i*4+8)))
		} else {
			for j := i; j < len(src); j++ {
				*(*uint32)(unsafe.Pointer(uintptr(dstPtr) + uintptr(j*4))) =
					*(*uint32)(unsafe.Pointer(uintptr(srcPtr) + uintptr(j*4)))
			}
		}
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

	// Process in 8-byte chunks
	for i := 0; i < len(dst); i += 8 {
		if i+8 <= len(dst) {
			*(*uint64)(unsafe.Pointer(&dst[i])) =
				*(*uint64)(unsafe.Pointer(&src1[i])) ^
				*(*uint64)(unsafe.Pointer(&src2[i]))
		} else {
			for j := i; j < len(dst); j++ {
				dst[j] = src1[j] ^ src2[j]
			}
		}
	}
}