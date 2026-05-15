package packet

import (
	"testing"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/types"
)

func BenchmarkPacketEncode(b *testing.B) {
	p := &Packet{
		Header: Header{
			Command:       command.RegisterSession,
			Length:        0,
			SessionHandle: 0,
			Status:        0,
			SenderContext: 0,
			Options:       0,
		},
		SpecificData: []byte{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = p.Encode()
	}
}

func BenchmarkCommonPacketFormatEncode(b *testing.B) {
	items := []CommonPacketFormatItem{
		{TypeID: ItemIDUCMM, Data: []byte{0x01, 0x02}},
		{TypeID: ItemIDUnconnectedMessage, Data: []byte{0x03, 0x04, 0x05}},
	}
	cpf := NewCommonPacketFormat(items)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = cpf.Encode()
	}
}

func BenchmarkCommonPacketFormatDecode(b *testing.B) {
	data := []byte{
		0x02, 0x00, // ItemCount
		0x00, 0x00, // TypeID (UCMM)
		0x02, 0x00, // Length
		0x01, 0x02, // Data
		0xB2, 0x00, // TypeID (UnconnectedMessage)
		0x03, 0x00, // Length
		0x03, 0x04, 0x05, // Data
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cpf := &CommonPacketFormat{}
		cpf.Decode(bufferx.NewReader(data))
	}
}

func BenchmarkMessageRouterRequestEncode(b *testing.B) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{0x02, 0x03})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mr.Encode()
	}
}

func BenchmarkMessageRouterResponseDecode(b *testing.B) {
	data := []byte{
		0x03,             // ReplyService
		0x00,             // Reserved
		0x00,             // GeneralStatus
		0x00,             // SizeOfAdditionalStatus
		0x01, 0x02, 0x03, // ResponseData
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rmr := &MessageRouterResponse{}
		rmr.Decode(data)
	}
}

func BenchmarkSpecificDataEncode(b *testing.B) {
	sd := &SpecificData{
		InterfaceHandle: 0x12345678,
		TimeOut:         0xABCD,
		Packet: NewCommonPacketFormat([]CommonPacketFormatItem{
			{TypeID: ItemIDUCMM, Data: nil},
		}),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = sd.Encode()
	}
}

func BenchmarkSpecificDataDecode(b *testing.B) {
	data := []byte{
		0x78, 0x56, 0x34, 0x12, // InterfaceHandle
		0xCD, 0xAB, // TimeOut
		0x01, 0x00, // ItemCount
		0x00, 0x00, // TypeID (UCMM)
		0x00, 0x00, // Length
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sd := &SpecificData{}
		sd.Decode(data)
	}
}

func BenchmarkNewUCMM(b *testing.B) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewUCMM(mr)
	}
}

func BenchmarkNewCMM(b *testing.B) {
	mr := NewMessageRouter(ServiceGetAttributes, []byte{0x01}, []byte{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewCMM(types.UDInt(0x12345678), types.UInt(0x0001), mr)
	}
}
