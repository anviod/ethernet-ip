package packet

import (
	"errors"

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
		if legacyErr := r.decodeLegacy(data); legacyErr == nil {
			return nil
		}
		return err
	}
	if io.Error() != nil {
		if legacyErr := r.decodeLegacy(data); legacyErr == nil {
			return nil
		}
		return io.Error()
	}
	if r.Packet.ItemCount == 0 && len(data) >= 6 {
		if legacyErr := r.decodeLegacy(data); legacyErr == nil {
			return nil
		}
	}
	return nil
}

func (r *SpecificData) decodeLegacy(data []byte) error {
	if len(data) < 6 {
		return errors.New("specific data too short")
	}
	io := bufferx.NewReader(data)
	io.RL(&r.InterfaceHandle)
	r.TimeOut = 0
	r.Packet = new(CommonPacketFormat)
	if err := r.Packet.Decode(io); err != nil {
		return err
	}
	return io.Error()
}
