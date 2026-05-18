package packet

import (
	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/types"
)

type SpecificData struct {
	InterfaceHandle types.UDInt
	TimeOut         types.UInt
	Packet          *CommonPacketFormat
}

func (r *SpecificData) Encode() []byte {
	io := bufferx.New(nil)
	io.WL(r.InterfaceHandle)
	io.WL(r.TimeOut)
	io.WL(r.Packet.Encode())
	return io.Bytes()
}

func (r *SpecificData) Decode(data []byte) error {
	io := bufferx.NewReader(data)
	io.RL(&r.InterfaceHandle)
	io.RL(&r.TimeOut)
	r.Packet = new(CommonPacketFormat)
	if err := r.Packet.Decode(io); err != nil {
		return err
	}
	return io.Error()
}
