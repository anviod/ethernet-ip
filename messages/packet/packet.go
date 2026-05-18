package packet

import (
	"errors"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/types"
)

type Header struct {
	Command       command.Command
	Length        types.UInt
	SessionHandle types.UDInt
	Status        types.UDInt
	SenderContext types.ULInt
	Options       types.UDInt
}

type Packet struct {
	Header
	SpecificData []byte
}

func (p *Packet) Encode() ([]byte, error) {
	if p.Length > 65511 {
		return nil, errors.New("specific data over length 65511")
	}

	if !command.CheckValid(p.Command) {
		return nil, errors.New("command not supported")
	}

	buffer := bufferx.NewWithCapacity(24 + len(p.SpecificData))
	buffer.WL(p.Command)
	buffer.WL(p.Length)
	buffer.WL(p.SessionHandle)
	buffer.WL(p.Status)
	buffer.WL(p.SenderContext)
	buffer.WL(p.Options)
	buffer.WL(p.SpecificData)
	if buffer.Error() != nil {
		return nil, buffer.Error()
	}

	return buffer.Bytes(), nil
}

// BatchEncode encodes multiple packets into a single byte slice
func BatchEncode(packets []*Packet) ([][]byte, error) {
	results := make([][]byte, len(packets))
	for i, p := range packets {
		data, err := p.Encode()
		if err != nil {
			return nil, err
		}
		results[i] = data
	}
	return results, nil
}

// BatchDecode decodes multiple packets from byte slices
func BatchDecode(data [][]byte) ([]*Packet, error) {
	results := make([]*Packet, len(data))
	for i, d := range data {
		p := &Packet{}
		reader := bufferx.NewReader(d)
		reader.RL(&p.Header)
		if reader.Error() != nil {
			return nil, reader.Error()
		}
		if p.Options != 0 {
			return nil, errors.New("wrong packet with non-zero option")
		}
		if int(p.Length) != reader.Len() {
			if p.Length == 0 && reader.Len() > 0 {
				p.Length = types.UInt(reader.Len())
			} else {
				return nil, errors.New("wrong packet length")
			}
		}
		p.SpecificData = reader.ReadBytes(reader.Len())
		if reader.Error() != nil {
			return nil, reader.Error()
		}
		results[i] = p
	}
	return results, nil
}
