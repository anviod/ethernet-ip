package bufferx

import (
	"testing"
)

func TestNew(t *testing.T) {
	// Test creating buffer with nil data
	buf := New(nil)
	if buf == nil {
		t.Error("New(nil) should return non-nil buffer")
	}
	if buf.Len() != 0 {
		t.Errorf("New(nil) should have length 0, got %d", buf.Len())
	}

	// Test creating buffer with data
	data := []byte{0x01, 0x02, 0x03}
	buf = New(data)
	if buf == nil {
		t.Error("New(data) should return non-nil buffer")
	}
	if buf.Len() != len(data) {
		t.Errorf("New(data) should have length %d, got %d", len(data), buf.Len())
	}
}

func TestWL(t *testing.T) {
	buf := New(nil)

	// Test writing uint16 in little-endian
	value := uint16(0xABCD)
	buf.WL(value)

	if buf.Error() != nil {
		t.Errorf("WL should not produce error, got: %v", buf.Error())
	}

	// Check bytes - little-endian: CD AB
	expected := []byte{0xCD, 0xAB}
	result := buf.Bytes()
	if len(result) != len(expected) {
		t.Errorf("WL should produce %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("WL byte %d: expected %02X, got %02X", i, b, result[i])
		}
	}
}

func TestWB(t *testing.T) {
	buf := New(nil)

	// Test writing uint16 in big-endian
	value := uint16(0xABCD)
	buf.WB(value)

	if buf.Error() != nil {
		t.Errorf("WB should not produce error, got: %v", buf.Error())
	}

	// Check bytes - big-endian: AB CD
	expected := []byte{0xAB, 0xCD}
	result := buf.Bytes()
	if len(result) != len(expected) {
		t.Errorf("WB should produce %d bytes, got %d", len(expected), len(result))
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("WB byte %d: expected %02X, got %02X", i, b, result[i])
		}
	}
}

func TestRL(t *testing.T) {
	// Test reading uint16 in little-endian
	data := []byte{0xCD, 0xAB} // 0xABCD in little-endian
	buf := New(data)

	var value uint16
	buf.RL(&value)

	if buf.Error() != nil {
		t.Errorf("RL should not produce error, got: %v", buf.Error())
	}
	if value != 0xABCD {
		t.Errorf("RL should read 0xABCD, got 0x%X", value)
	}
}

func TestRB(t *testing.T) {
	// Test reading uint16 in big-endian
	data := []byte{0xAB, 0xCD} // 0xABCD in big-endian
	buf := New(data)

	var value uint16
	buf.RB(&value)

	if buf.Error() != nil {
		t.Errorf("RB should not produce error, got: %v", buf.Error())
	}
	if value != 0xABCD {
		t.Errorf("RB should read 0xABCD, got 0x%X", value)
	}
}

func TestError(t *testing.T) {
	// Test that Error returns nil when no error
	buf := New(nil)
	buf.WL(uint16(0))
	if buf.Error() != nil {
		t.Errorf("Error() should return nil, got: %v", buf.Error())
	}
}

func TestWriteReadRoundTrip(t *testing.T) {
	// Test writing and reading various types
	buf := New(nil)

	// Write values
	buf.WL(uint8(0x12))
	buf.WL(uint16(0x3456))
	buf.WL(uint32(0x789ABCDE))

	// Read back
	buf2 := New(buf.Bytes())

	var u8 uint8
	var u16 uint16
	var u32 uint32

	buf2.RL(&u8)
	buf2.RL(&u16)
	buf2.RL(&u32)

	if u8 != 0x12 {
		t.Errorf("Expected uint8 0x12, got 0x%X", u8)
	}
	if u16 != 0x3456 {
		t.Errorf("Expected uint16 0x3456, got 0x%X", u16)
	}
	if u32 != 0x789ABCDE {
		t.Errorf("Expected uint32 0x789ABCDE, got 0x%X", u32)
	}
}
