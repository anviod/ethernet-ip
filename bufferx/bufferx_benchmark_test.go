package bufferx

import (
	"testing"
)

func BenchmarkWL_UInt16(b *testing.B) {
	buf := New(nil)
	value := uint16(0xABCD)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WL(value)
	}
}

func BenchmarkWL_Multiple(b *testing.B) {
	buf := New(nil)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WL(uint8(0x12))
		buf.WL(uint16(0x3456))
		buf.WL(uint32(0x789ABCDE))
		buf.WL(uint64(0x123456789ABCDEF0))
	}
}

func BenchmarkRL_UInt16(b *testing.B) {
	data := []byte{0xCD, 0xAB}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := New(data)
		var value uint16
		buf.RL(&value)
	}
}

func BenchmarkRL_Multiple(b *testing.B) {
	data := []byte{
		0x12,           // uint8
		0x56, 0x34,     // uint16 (little-endian)
		0xDE, 0xBC, 0x9A, 0x78, // uint32 (little-endian)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := New(data)
		var u8 uint8
		var u16 uint16
		var u32 uint32
		buf.RL(&u8)
		buf.RL(&u16)
		buf.RL(&u32)
	}
}

func BenchmarkBufferPool(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := Get()
		buf.WL(uint16(0xABCD))
		Put(buf)
	}
}

func BenchmarkBufferPool_Reuse(b *testing.B) {
	buf := Get()
	defer Put(buf)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		buf.WL(uint16(0xABCD))
	}
}

func BenchmarkNewWithCapacity(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewWithCapacity(128)
		buf.WL(uint16(0xABCD))
	}
}

func BenchmarkNew(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := New(nil)
		buf.WL(uint16(0xABCD))
	}
}

func BenchmarkReaderRL_UInt16(b *testing.B) {
	data := []byte{0xCD, 0xAB}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(data)
		var value uint16
		reader.RL(&value)
	}
}

func BenchmarkReaderRL_Multiple(b *testing.B) {
	data := []byte{
		0x12,           // uint8
		0x56, 0x34,     // uint16 (little-endian)
		0xDE, 0xBC, 0x9A, 0x78, // uint32 (little-endian)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader := NewReader(data)
		var u8 uint8
		var u16 uint16
		var u32 uint32
		reader.RL(&u8)
		reader.RL(&u16)
		reader.RL(&u32)
	}
}

func BenchmarkBatchEncode(b *testing.B) {
	// Simplified benchmark without packet import
	data := make([][]byte, 10)
	for i := range data {
		data[i] = []byte{0x01, 0x02, 0x03, 0x04}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := make([][]byte, len(data))
		copy(result, data)
	}
}

func BenchmarkBatchDecode(b *testing.B) {
	// Simplified benchmark without packet import
	data := make([][]byte, 10)
	for i := range data {
		data[i] = []byte{0x01, 0x02, 0x03, 0x04}
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := make([][]byte, len(data))
		for j, d := range data {
			result[j] = make([]byte, len(d))
			copy(result[j], d)
		}
	}
}