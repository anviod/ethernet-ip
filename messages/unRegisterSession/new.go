package unRegisterSession

import (
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

func New(session types.UDInt, context types.ULInt) (*packet.Packet, error) {
	return &packet.Packet{
		Header: packet.Header{
			Command:       command.UnRegisterSession,
			Length:        0,
			SessionHandle: session,
			Status:        0,
			SenderContext: context,
			Options:       0,
		},
		SpecificData: nil,
	}, nil
}
