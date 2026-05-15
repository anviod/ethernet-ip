package listInterface

import (
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

func New(context types.ULInt) (*packet.Packet, error) {
	return &packet.Packet{
		Header: packet.Header{
			Command:       command.ListInterfaces,
			Length:        0,
			SessionHandle: 0,
			Status:        0,
			SenderContext: context,
			Options:       0,
		},
		SpecificData: nil,
	}, nil
}
