package ethernet_ip

import (
	"testing"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

func buildResponseBytes(t types.UInt, payload []byte) []byte {
	io := bufferx.New(nil)
	io.WL(t)
	io.WL(payload)
	return io.Bytes()
}

func buildResponseWithCountBytes(t types.UInt, count types.UInt, payload []byte) []byte {
	io := bufferx.New(nil)
	io.WL(t)
	io.WL(count)
	io.WL(payload)
	return io.Bytes()
}

func TestReadParserAtomicScalarNoCount(t *testing.T) {
	response := buildResponseBytes(BOOL, []byte{0x01})
	mr := &packet.MessageRouterResponse{ResponseData: response}
	tag := new(Tag)
	if err := tag.readParser(mr, nil); err != nil {
		t.Fatalf("readParser failed: %v", err)
	}
	if tag.Type != BOOL {
		t.Fatalf("expected type BOOL, got 0x%04x", tag.Type)
	}
	if got := tag.Bool(); got != true {
		t.Fatalf("expected true, got %v", got)
	}
}

func TestReadParserScalarCountAndPayload(t *testing.T) {
	response := buildResponseWithCountBytes(DINT, types.UInt(1), []byte{0x39, 0x30, 0x00, 0x00})
	mr := &packet.MessageRouterResponse{ResponseData: response}
	tag := new(Tag)
	if err := tag.readParser(mr, nil); err != nil {
		t.Fatalf("readParser failed: %v", err)
	}
	if tag.Type != DINT {
		t.Fatalf("expected type DINT, got 0x%04x", tag.Type)
	}
	if got := tag.Int32(); got != int32(12345) {
		t.Fatalf("expected 12345, got %d", got)
	}
}

func BenchmarkReadParserBoolAtomic(b *testing.B) {
	response := buildResponseBytes(BOOL, []byte{0x01})
	mr := &packet.MessageRouterResponse{ResponseData: response}
	tag := new(Tag)
	for i := 0; i < b.N; i++ {
		if err := tag.readParser(mr, nil); err != nil {
			b.Fatalf("readParser failed: %v", err)
		}
	}
}

func BenchmarkReadParserDintPayload(b *testing.B) {
	response := buildResponseBytes(DINT, []byte{0x39, 0x30, 0x00, 0x00})
	mr := &packet.MessageRouterResponse{ResponseData: response}
	tag := new(Tag)
	for i := 0; i < b.N; i++ {
		if err := tag.readParser(mr, nil); err != nil {
			b.Fatalf("readParser failed: %v", err)
		}
	}
}
