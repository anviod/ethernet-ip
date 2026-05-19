package nop

import (
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

func New(data []byte) (*packet.Packet, error) {
	// Ensure NOP packet has minimal valid data per EtherNet/IP specification
	// Empty NOP packets can cause state machine errors in some servers (e.g., cpppo)
	if data == nil || len(data) == 0 {
		data = []byte{0x00}
	}
	return &packet.Packet{
		Header: packet.Header{
			Command:       command.NOP,
			Length:        types.UInt(len(data)),
			SessionHandle: 0,
			Status:        0,
			SenderContext: 0,
			Options:       0,
		},
		SpecificData: data,
	}, nil
}
