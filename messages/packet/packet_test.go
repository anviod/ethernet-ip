package packet

import (
	"testing"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/types"
)

func TestPacketEncode(t *testing.T) {
	tests := []struct {
		name    string
		packet  *Packet
		wantErr bool
	}{
		{
			name: "valid packet",
			packet: &Packet{
				Header: Header{
					Command:       command.RegisterSession,
					Length:        0,
					SessionHandle: 0,
					Status:        0,
					SenderContext: 0,
					Options:       0,
				},
				SpecificData: []byte{},
			},
			wantErr: false,
		},
		{
			name: "invalid command",
			packet: &Packet{
				Header: Header{
					Command: 0xFF, // Invalid command
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.packet.Encode()
			if (err != nil) != tt.wantErr {
				t.Errorf("Packet.Encode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCommonPacketFormatItemEncode(t *testing.T) {
	item := &CommonPacketFormatItem{
		TypeID: ItemIDUCMM,
		Data:   []byte{0x01, 0x02, 0x03},
	}

	encoded := item.Encode()
	if len(encoded) == 0 {
		t.Error("CommonPacketFormatItem.Encode() should return non-empty byte slice")
	}

	// Check TypeID (2 bytes) + Length (2 bytes) + Data (3 bytes) = 7 bytes
	if len(encoded) != 7 {
		t.Errorf("Expected length 7, got %d", len(encoded))
	}
}

func TestCommonPacketFormatItemDecode(t *testing.T) {
	// TypeID (0x0000) + Length (0x0003) + Data (0x01 0x02 0x03)
	data := []byte{0x00, 0x00, 0x03, 0x00, 0x01, 0x02, 0x03}

	item := &CommonPacketFormatItem{}
	item.Decode(bufferx.NewReader(data))

	if item.TypeID != ItemIDUCMM {
		t.Errorf("Expected TypeID 0x%04X, got 0x%04X", ItemIDUCMM, item.TypeID)
	}
	if item.Length != 3 {
		t.Errorf("Expected Length 3, got %d", item.Length)
	}
	if len(item.Data) != 3 {
		t.Errorf("Expected Data length 3, got %d", len(item.Data))
	}
}

func TestCommonPacketFormatEncode(t *testing.T) {
	items := []CommonPacketFormatItem{
		{
			TypeID: ItemIDUCMM,
			Data:   []byte{0x01, 0x02},
		},
	}

	cpf := NewCommonPacketFormat(items)
	encoded := cpf.Encode()

	if len(encoded) == 0 {
		t.Error("CommonPacketFormat.Encode() should return non-empty byte slice")
	}
}

func TestMessageRouterRequestEncode(t *testing.T) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{0x02, 0x03})
	encoded := mr.Encode()

	if len(encoded) == 0 {
		t.Error("MessageRouterRequest.Encode() should return non-empty byte slice")
	}

	// Service (1) + PathSize (1) + Path (1) + Data (2) = 5 bytes
	if len(encoded) != 5 {
		t.Errorf("Expected length 5, got %d", len(encoded))
	}
}

func TestMessageRouterResponseDecode(t *testing.T) {
	// ReplyService + Reserved + GeneralStatus + SizeOfAdditionalStatus + AdditionalStatus + ResponseData
	data := []byte{
		0x03,             // ReplyService
		0x00,             // Reserved
		0x00,             // GeneralStatus
		0x00,             // SizeOfAdditionalStatus
		0x01, 0x02, 0x03, // ResponseData
	}

	rmr := &MessageRouterResponse{}
	rmr.Decode(data)

	if rmr.ReplyService != 0x03 {
		t.Errorf("Expected ReplyService 0x03, got 0x%02X", rmr.ReplyService)
	}
	if len(rmr.ResponseData) != 3 {
		t.Errorf("Expected ResponseData length 3, got %d", len(rmr.ResponseData))
	}
}

func TestSpecificDataEncode(t *testing.T) {
	sd := &SpecificData{
		InterfaceHandle: 0x12345678,
		TimeOut:         0xABCD,
		Packet: NewCommonPacketFormat([]CommonPacketFormatItem{
			{TypeID: ItemIDUCMM, Data: nil},
		}),
	}

	encoded := sd.Encode()
	if len(encoded) == 0 {
		t.Error("SpecificData.Encode() should return non-empty byte slice")
	}
}

func TestSpecificDataDecode(t *testing.T) {
	// InterfaceHandle (4 bytes) + TimeOut (2 bytes) + CMM data
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // InterfaceHandle (little-endian)
		0xCD, 0xAB, // TimeOut (little-endian)
		0x01, 0x00, // ItemCount
		0x00, 0x00, // TypeID (UCMM)
		0x00, 0x00, // Length
	}

	sd := &SpecificData{}
	sd.Decode(data)

	if sd.InterfaceHandle != 0x12345678 {
		t.Errorf("Expected InterfaceHandle 0x%08X, got 0x%08X", 0x12345678, sd.InterfaceHandle)
	}
	if sd.TimeOut != 0xABCD {
		t.Errorf("Expected TimeOut 0x%04X, got 0x%04X", 0xABCD, sd.TimeOut)
	}
	if sd.Packet == nil {
		t.Error("Packet should not be nil")
	}
}

func TestNewUCMM(t *testing.T) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{})
	cpf := NewUCMM(mr)

	if cpf == nil {
		t.Error("NewUCMM should return non-nil CommonPacketFormat")
	}
	if cpf.ItemCount != 2 {
		t.Errorf("Expected ItemCount 2, got %d", cpf.ItemCount)
	}
}

func TestNewCMM(t *testing.T) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{})
	cpf := NewCMM(types.UDInt(0x12345678), types.UInt(0x0001), mr)

	if cpf == nil {
		t.Error("NewCMM should return non-nil CommonPacketFormat")
	}
	if cpf.ItemCount != 2 {
		t.Errorf("Expected ItemCount 2, got %d", cpf.ItemCount)
	}
}

// Helper function to create BufferX
func NewBufferX(data []byte) *bufferx.BufferX {
	return bufferx.New(data)
}
