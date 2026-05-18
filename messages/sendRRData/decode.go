package sendRRData

import "github.com/anviod/ethernet-ip/messages/packet"

func Decode(_packet *packet.Packet) (*packet.SpecificData, error) {
	result := new(packet.SpecificData)
	if err := result.Decode(_packet.SpecificData); err != nil {
		return nil, err
	}
	return result, nil
}
