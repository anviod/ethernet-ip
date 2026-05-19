package ethernet_ip

import (
	"bytes"
	"sync"
	"testing"

	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

func TestTag_SetInt32(t *testing.T) {
	tag := &Tag{
		Lock: &sync.Mutex{},
		Type: INT,
	}

	tag.SetInt32(12345)

	if len(tag.wValue) != 2 {
		t.Errorf("SetInt32(INT) expected 2 bytes, got %d", len(tag.wValue))
	}

	tag.Type = DINT
	tag.SetInt32(123456789)
	if len(tag.wValue) != 4 {
		t.Errorf("SetInt32(DINT) expected 4 bytes, got %d", len(tag.wValue))
	}
}

func TestTag_GetInt16(t *testing.T) {
	tag := &Tag{
		Lock:  &sync.Mutex{},
		Type:  INT,
		value: []byte{0x39, 0x30}, // 12345 in little-endian
	}

	result := tag.Int16()
	expected := int16(12345)
	if result != expected {
		t.Errorf("Int16() expected %d, got %d", expected, result)
	}
}

func TestTag_GetInt32(t *testing.T) {
	tag := &Tag{
		Lock:  &sync.Mutex{},
		Type:  DINT,
		value: []byte{0x15, 0xcd, 0x5b, 0x07}, // 123456789 in little-endian
	}

	result := tag.Int32()
	expected := int32(123456789)
	if result != expected {
		t.Errorf("Int32() expected %d, got %d", expected, result)
	}
}

func TestTag_GetFloat32(t *testing.T) {
	tag := &Tag{
		Lock:  &sync.Mutex{},
		Type:  REAL,
		value: []byte{0x00, 0x00, 0x80, 0x3f}, // 1.0 in IEEE 754 float32
	}

	result := tag.Float32()
	expected := float32(1.0)
	if result != expected {
		t.Errorf("Float32() expected %f, got %f", expected, result)
	}
}

func TestTag_GetString(t *testing.T) {
	testCases := []struct {
		name     string
		value    []byte
		expected string
	}{
		{
			name:     "empty string",
			value:    []byte{0x00, 0x00, 0x00, 0x00},
			expected: "",
		},
		{
			name:     "hello",
			value:    []byte{0x05, 0x00, 0x00, 0x00, 'h', 'e', 'l', 'l', 'o'},
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock:  &sync.Mutex{},
				Type:  STRING,
				value: tc.value,
			}

			result := tag.String()
			if result != tc.expected {
				t.Errorf("String() expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestTag_Bool(t *testing.T) {
	testCases := []struct {
		name     string
		value    []byte
		expected bool
	}{
		{
			name:     "true",
			value:    []byte{0x01},
			expected: true,
		},
		{
			name:     "false",
			value:    []byte{0x00},
			expected: false,
		},
		{
			name:     "non-zero",
			value:    []byte{0xFF},
			expected: true,
		},
		{
			name:     "empty value",
			value:    []byte{},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock:  &sync.Mutex{},
				Type:  BOOL,
				value: tc.value,
			}

			result := tag.Bool()
			if result != tc.expected {
				t.Errorf("Bool() expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestTag_count(t *testing.T) {
	testCases := []struct {
		name     string
		dim1Len  types.UDInt
		dim2Len  types.UDInt
		dim3Len  types.UDInt
		expected types.UInt
	}{
		{
			name:     "no dimensions",
			dim1Len:  0,
			dim2Len:  0,
			dim3Len:  0,
			expected: 1,
		},
		{
			name:     "one dimension",
			dim1Len:  10,
			dim2Len:  0,
			dim3Len:  0,
			expected: 10,
		},
		{
			name:     "two dimensions",
			dim1Len:  10,
			dim2Len:  5,
			dim3Len:  0,
			expected: 50,
		},
		{
			name:     "three dimensions",
			dim1Len:  10,
			dim2Len:  5,
			dim3Len:  2,
			expected: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock:    &sync.Mutex{},
				dim1Len: tc.dim1Len,
				dim2Len: tc.dim2Len,
				dim3Len: tc.dim3Len,
			}

			result := tag.count()
			if result != tc.expected {
				t.Errorf("count() expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestTag_dims(t *testing.T) {
	testCases := []struct {
		name     string
		tagType  types.UInt
		expected types.USInt
	}{
		{
			name:     "no dimension bit set",
			tagType:  INT,
			expected: 0,
		},
		{
			name:     "1 dimension",
			tagType:  types.UInt(0x20C3),
			expected: 1,
		},
		{
			name:     "2 dimensions",
			tagType:  types.UInt(0x40C3),
			expected: 2,
		},
		{
			name:     "3 dimensions",
			tagType:  types.UInt(0x60C3),
			expected: 3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock: &sync.Mutex{},
				Type: tc.tagType,
			}

			result := tag.dims()
			if result != tc.expected {
				t.Errorf("dims() expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestTag_TypeString(t *testing.T) {
	testCases := []struct {
		name     string
		tagType  types.UInt
		contains string
	}{
		{
			name:     "atomic INT",
			tagType:  INT,
			contains: "atomic",
		},
		{
			name:     "struct STRING",
			tagType:  STRING2,
			contains: "struct",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock: &sync.Mutex{},
				Type: tc.tagType,
			}

			result := tag.TypeString()
			if !bytes.Contains([]byte(result), []byte(tc.contains)) {
				t.Errorf("TypeString() expected to contain %q, got %q", tc.contains, result)
			}
		})
	}
}

func TestTag_Name(t *testing.T) {
	tag := &Tag{
		Lock: &sync.Mutex{},
		name: []byte("TestTag"),
	}

	result := tag.Name()
	if result != "TestTag" {
		t.Errorf("Name() expected %q, got %q", "TestTag", result)
	}
}

func TestTag_readParser_AtomicScalar(t *testing.T) {
	tag := &Tag{
		Lock: &sync.Mutex{},
		Type: INT,
	}

	mrres := &packet.MessageRouterResponse{
		GeneralStatus: 0,
		ResponseData:  []byte{0xc3, 0x01, 0x39, 0x30}, // INT type (0xc3), count 1, value 12345
	}

	err := tag.readParser(mrres, nil)
	if err != nil {
		t.Errorf("readParser() returned error: %v", err)
	}

	if len(tag.value) != 2 {
		t.Errorf("readParser() expected 2 bytes value, got %d", len(tag.value))
	}

	result := tag.Int16()
	if result != 12345 {
		t.Errorf("readParser() expected value 12345, got %d", result)
	}
}

func TestTag_readParser_ExtendedType(t *testing.T) {
	tag := &Tag{
		Lock: &sync.Mutex{},
		Type: 0,
	}

	mrres := &packet.MessageRouterResponse{
		GeneralStatus: 0,
		ResponseData:  []byte{0xa0, 0x02, 0xc3, 0x00, 0x01, 0x00, 0x39, 0x30}, // Extended type (0x2a0), INT (0xc3), count 1, value 12345
	}

	err := tag.readParser(mrres, nil)
	if err != nil {
		t.Errorf("readParser() returned error: %v", err)
	}

	if tag.Type != INT {
		t.Errorf("readParser() expected type INT (0xc3), got 0x%04x", tag.Type)
	}

	result := tag.Int16()
	if result != 12345 {
		t.Errorf("readParser() expected value 12345, got %d", result)
	}
}

func TestTag_readParser_NoCount(t *testing.T) {
	tag := &Tag{
		Lock: &sync.Mutex{},
		Type: INT,
	}

	mrres := &packet.MessageRouterResponse{
		GeneralStatus: 0,
		ResponseData:  []byte{0xc3, 0x39, 0x30}, // INT type (0xc3), value 12345 (no count, 3 bytes - triggers error on count read)
	}

	err := tag.readParser(mrres, nil)
	if err != nil {
		t.Errorf("readParser() returned error: %v", err)
	}

	// With 3 bytes [0xc3, 0x39, 0x30]:
	// - Read type: pos=2, remainingBeforeCount=1
	// - Read count fails (need 2 bytes, only 1 left)
	// - offset = 3 - 1 = 2, payload = [0x30]
	// This tests the no-count error path, though the value will be truncated
	if len(tag.value) != 1 {
		t.Errorf("readParser() expected 1 byte value, got %d", len(tag.value))
	}
}

func TestTag_GetValue(t *testing.T) {
	testCases := []struct {
		name     string
		tagType  types.UInt
		value    []byte
		expected interface{}
	}{
		{
			name:     "INT",
			tagType:  INT,
			value:    []byte{0x39, 0x30},
			expected: int16(12345),
		},
		{
			name:     "DINT",
			tagType:  DINT,
			value:    []byte{0x15, 0xcd, 0x5b, 0x07},
			expected: int32(123456789),
		},
		{
			name:     "BOOL",
			tagType:  BOOL,
			value:    []byte{0x01},
			expected: true,
		},
		{
			name:     "STRING",
			tagType:  STRING,
			value:    []byte{0x05, 0x00, 0x00, 0x00, 'h', 'e', 'l', 'l', 'o'},
			expected: "hello",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tag := &Tag{
				Lock:  &sync.Mutex{},
				Type:  tc.tagType,
				value: tc.value,
			}

			result := tag.GetValue()
			if result != tc.expected {
				t.Errorf("GetValue() expected %v, got %v", tc.expected, result)
			}
		})
	}
}
