package path

import (
	"testing"

	"github.com/anviod/ethernet-ip/types"
)

func TestPaths(t *testing.T) {
	path1 := []byte{0x01, 0x02}
	path2 := []byte{0x03, 0x04, 0x05}

	result := Paths(path1, path2)

	if len(result) != len(path1)+len(path2) {
		t.Errorf("Expected length %d, got %d", len(path1)+len(path2), len(result))
	}

	expected := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf("Byte %d: expected %02X, got %02X", i, b, result[i])
		}
	}
}

func TestDataBuild(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03}
	result := DataBuild(DataTypeSimple, data, true)

	if len(result) == 0 {
		t.Error("DataBuild should return non-empty byte slice")
	}

	// First byte should contain SegmentTypeData (0x40) and DataTypeSimple (0x00)
	expectedFirstByte := uint8(SegmentTypeData) | uint8(DataTypeSimple)
	if result[0] != expectedFirstByte {
		t.Errorf("First byte: expected %02X, got %02X", expectedFirstByte, result[0])
	}
}

func TestLogicalBuild(t *testing.T) {
	tests := []struct {
		name     string
		tp       LogicalType
		address  types.UDInt
		padded   bool
		expected int // expected minimum length
	}{
		{"small address", LogicalTypeClassID, 10, false, 2},
		{"medium address", LogicalTypeClassID, 1000, false, 3},
		{"large address", LogicalTypeClassID, 100000, false, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := LogicalBuild(tt.tp, tt.address, tt.padded)
			if len(result) < tt.expected {
				t.Errorf("Expected minimum length %d, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestPortBuild(t *testing.T) {
	tests := []struct {
		name     string
		link     []byte
		portID   uint16
		padded   bool
		expected int // expected minimum length
	}{
		{"simple port", []byte{0x01}, 1, false, 2},
		{"extended port", []byte{0x01}, 20, false, 3},
		{"extended link", []byte{0x01, 0x02}, 1, false, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PortBuild(tt.link, tt.portID, tt.padded)
			if len(result) < tt.expected {
				t.Errorf("Expected minimum length %d, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestLogicalBuildFormat(t *testing.T) {
	// Test that small address uses 1-byte format
	result := LogicalBuild(LogicalTypeClassID, 100, false)
	if len(result) != 2 {
		t.Errorf("Small address should produce 2 bytes, got %d", len(result))
	}

	// Test that medium address uses 2-byte format
	result = LogicalBuild(LogicalTypeClassID, 1000, false)
	if len(result) != 3 {
		t.Errorf("Medium address should produce 3 bytes, got %d", len(result))
	}

	// Test that large address uses 4-byte format
	result = LogicalBuild(LogicalTypeClassID, 100000, false)
	if len(result) != 5 {
		t.Errorf("Large address should produce 5 bytes, got %d", len(result))
	}
}

func TestPortBuildExtended(t *testing.T) {
	// Test extended port ID (> 14) - firstByte(1) + portID(2) + link(1) = 4 bytes
	result := PortBuild([]byte{0x01}, 20, false)
	if len(result) != 4 {
		t.Errorf("Extended port should produce 4 bytes, got %d", len(result))
	}

	// Test extended link (len > 1)
	result = PortBuild([]byte{0x01, 0x02}, 1, false)
	if len(result) != 4 {
		t.Errorf("Extended link should produce 4 bytes, got %d", len(result))
	}
}
