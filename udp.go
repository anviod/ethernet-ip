package ethernet_ip

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

type DiscoveredDevice struct {
	IPAddress     net.IP
	MacAddress    string
	DeviceName    string
	ProductCode   uint16
	RevisionMajor uint8
	RevisionMinor uint8
	Status        uint16
	SerialNumber  uint32
	VendorID      uint16
}

func DiscoverDevices(timeout time.Duration) ([]*DiscoveredDevice, error) {
	return DiscoverDevicesWithPort(timeout, 44818)
}

func DiscoverDevicesWithPort(timeout time.Duration, port int) ([]*DiscoveredDevice, error) {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{Port: 0})
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	err = conn.SetDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, err
	}

	broadcastAddr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: port,
	}

	reqPacket := buildListIdentityRequest()
	reqData, err := reqPacket.Encode()
	if err != nil {
		return nil, err
	}

	_, err = conn.WriteToUDP(reqData, broadcastAddr)
	if err != nil {
		return nil, err
	}

	var devices []*DiscoveredDevice
	buf := make([]byte, 1024*64)

	for {
		n, addr, err := conn.ReadFromUDP(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				break
			}
			return devices, err
		}

		device, err := parseListIdentityResponse(buf[:n])
		if err != nil {
			continue
		}
		device.IPAddress = addr.IP
		devices = append(devices, device)
	}

	return devices, nil
}

func buildListIdentityRequest() *packet.Packet {
	p := &packet.Packet{}
	p.Command = command.ListIdentity
	p.Length = 0
	p.SessionHandle = 0
	p.Status = 0
	p.SenderContext = 0
	p.Options = 0
	return p
}

func parseListIdentityResponse(data []byte) (*DiscoveredDevice, error) {
	if len(data) < 24 {
		return nil, errors.New("invalid packet length")
	}

	p := &packet.Packet{}
	reader := bufferx.NewReader(data)

	p.Command = command.Command(reader.ReadUint16())
	p.Length = types.UInt(reader.ReadUint16())
	p.SessionHandle = types.UDInt(reader.ReadUint32())
	p.Status = types.UDInt(reader.ReadUint32())
	p.SenderContext = types.ULInt(reader.ReadUint64())
	p.Options = types.UDInt(reader.ReadUint32())

	if p.Command != command.ListIdentity {
		return nil, errors.New("invalid response command")
	}

	p.SpecificData = reader.ReadBytes(int(p.Length))
	if reader.Error() != nil {
		return nil, reader.Error()
	}

	cpf := &packet.CommonPacketFormat{}
	cpfReader := bufferx.NewReader(p.SpecificData)
	err := cpf.Decode(cpfReader)
	if err != nil {
		return nil, err
	}

	for _, item := range cpf.Items {
		if item.TypeID == 0x000C {
			return parseListIdentityItem(item.Data)
		}
	}

	return nil, errors.New("no ListIdentity item found")
}

func parseListIdentityItem(data []byte) (*DiscoveredDevice, error) {
	if len(data) < 24 {
		return nil, errors.New("invalid ListIdentity item length")
	}

	device := &DiscoveredDevice{}
	io := bufferx.New(data)

	var vendorID types.UInt
	io.RL(&vendorID)
	device.VendorID = uint16(vendorID)

	var productCode types.UInt
	io.RL(&productCode)
	device.ProductCode = uint16(productCode)

	var revisionMajor types.USInt
	io.RL(&revisionMajor)
	device.RevisionMajor = uint8(revisionMajor)

	var revisionMinor types.USInt
	io.RL(&revisionMinor)
	device.RevisionMinor = uint8(revisionMinor)

	var status types.UInt
	io.RL(&status)
	device.Status = uint16(status)

	var serialNumber types.UDInt
	io.RL(&serialNumber)
	device.SerialNumber = uint32(serialNumber)

	var macBytes [6]byte
	io.RL(&macBytes)
	macStr := make([]byte, 17)
	for i := 0; i < 6; i++ {
		b := fmt.Sprintf("%02X", macBytes[i])
		copy(macStr[i*3:i*3+2], []byte(b))
		if i < 5 {
			macStr[i*3+2] = ':'
		}
	}
	device.MacAddress = string(macStr)

	nameLen := int(io.Len())
	if nameLen > 0 {
		nameBytes := make([]byte, nameLen)
		io.RL(nameBytes)
		nullIdx := bytes.IndexByte(nameBytes, 0)
		if nullIdx >= 0 {
			nameBytes = nameBytes[:nullIdx]
		}
		device.DeviceName = string(nameBytes)
	}

	return device, nil
}

func (d *DiscoveredDevice) String() string {
	return d.DeviceName
}
