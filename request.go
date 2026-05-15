package ethernet_ip

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/messages/listIdentity"
	"github.com/anviod/ethernet-ip/messages/listInterface"
	"github.com/anviod/ethernet-ip/messages/listServices"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/messages/registerSession"
	"github.com/anviod/ethernet-ip/messages/sendRRData"
	"github.com/anviod/ethernet-ip/messages/sendUnitData"
	"github.com/anviod/ethernet-ip/messages/unRegisterSession"
	"github.com/anviod/ethernet-ip/path"
	"github.com/anviod/ethernet-ip/types"
	"github.com/anviod/ethernet-ip/utils"
)

var (
	localRandMu   sync.Mutex
	localRandGen  = rand.New(rand.NewSource(0))
)

func randIntn(n int) int {
	localRandMu.Lock()
	defer localRandMu.Unlock()
	return localRandGen.Intn(n)
}

func (t *EIPTCP) request(packet *packet.Packet) (*packet.Packet, error) {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()

	if t.tcpConn == nil {
		return nil, errors.New("connect first")
	}

	b, err := packet.Encode()
	if err != nil {
		return nil, err
	}

	if err := t.write(b); err != nil {
		return nil, err
	}

	return t.read()
}

func (t *EIPTCP) RegisterSession() error {
	ctx := contextGenerator()
	requestPacket, err := registerSession.New(ctx)
	if err != nil {
		return err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return err
	}

	t.session = responsePacket.SessionHandle
	return nil
}

func (t *EIPTCP) UnRegisterSession() error {
	ctx := contextGenerator()
	requestPacket, err := unRegisterSession.New(t.session, ctx)
	if err != nil {
		return err
	}

	_, _ = t.request(requestPacket)

	_ = t.tcpConn.Close()
	t.tcpConn = nil
	return nil
}

func (t *EIPTCP) ListInterface() (*listInterface.ListInterface, error) {
	ctx := contextGenerator()
	requestPacket, err := listInterface.New(ctx)
	if err != nil {
		return nil, err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return nil, err
	}

	return listInterface.Decode(responsePacket)
}

func (t *EIPTCP) ListServices() (*listServices.ListServices, error) {
	ctx := contextGenerator()
	requestPacket, err := listServices.New(ctx)
	if err != nil {
		return nil, err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return nil, err
	}

	return listServices.Decode(responsePacket)
}

func (t *EIPTCP) ListIdentity() (*listIdentity.ListIdentity, error) {
	ctx := contextGenerator()
	requestPacket, err := listIdentity.New(ctx)
	if err != nil {
		return nil, err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return nil, err
	}

	return listIdentity.Decode(responsePacket)
}

func (t *EIPTCP) SendRRData(cpf *packet.CommonPacketFormat, timeout types.UInt) (*packet.SpecificData, error) {
	ctx := contextGenerator()
	requestPacket, err := sendRRData.New(t.session, ctx, cpf, timeout)
	if err != nil {
		return nil, err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return nil, err
	}

	return sendRRData.Decode(responsePacket)
}

func (t *EIPTCP) SendUnitData(cpf *packet.CommonPacketFormat) (*packet.SpecificData, error) {
	ctx := contextGenerator()
	requestPacket, err := sendUnitData.New(t.session, ctx, cpf)
	if err != nil {
		return nil, err
	}

	responsePacket, err := t.request(requestPacket)
	if err != nil {
		return nil, err
	}

	spd, err := sendUnitData.Decode(responsePacket)
	if spd != nil {
		spd.Packet.Items[1].Data = spd.Packet.Items[1].Data[2:]
	}
	return spd, err
}

func (t *EIPTCP) Send(mr *packet.MessageRouterRequest) (*packet.SpecificData, error) {
	if !t.established {
		mr = packet.UnConnected(t.config.Slot, t.config.TimeTick, t.config.TimeTickOut, mr)
	}
	if t.established {
		t.seqNum += 1
		return t.SendUnitData(packet.NewCMM(t.connID, t.seqNum, mr))
	} else {
		return t.SendRRData(packet.NewUCMM(mr), 10)
	}
}

func (t *EIPTCP) ForwardOpen() error {
	io := bufferx.NewWithCapacity(64)
	io.WL(types.USInt(3))
	io.WL(types.USInt(125))
	io.WL(types.UDInt(0))
	io.WL(types.UDInt(randIntn(2147483647)))
	io.WL(types.UInt(randIntn(32767)))
	io.WL(types.UInt(0x3333))
	io.WL(types.UDInt(0x1337))
	io.WL(types.UDInt(5))
	io.WL(types.UDInt(1000000))
	io.WL(types.UInt(0x43f4))
	io.WL(types.UDInt(1000000))
	io.WL(types.UInt(0x43f4))
	io.WL(types.USInt(0xA3))

	portPath := packet.Paths(
		path.PortBuild([]byte{t.config.Slot}, 1, true),
		path.LogicalBuild(path.LogicalTypeClassID, 0x02, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	)
	io.WL(utils.Len(portPath))
	io.WL(portPath)

	mr := packet.NewMessageRouter(packet.ServiceForwardOpen, packet.Paths(
		path.LogicalBuild(path.LogicalTypeClassID, 0x06, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	), io.Bytes())

	sd, err := t.SendRRData(packet.NewUCMM(mr), 10)
	if err != nil {
		return err
	}

	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(sd.Packet.Items[1].Data)
	io1 := bufferx.New(rmr.ResponseData)
	io1.RL(&t.connID)
	t.established = true

	return nil
}

func (t *EIPTCP) ForwardOpenLarge() error {
	io := bufferx.NewWithCapacity(64)
	io.WL(types.USInt(3))
	io.WL(types.USInt(125))
	io.WL(types.UDInt(0))
	io.WL(types.UDInt(randIntn(2147483647)))
	io.WL(types.UInt(randIntn(32767)))
	io.WL(types.UInt(0x3333))
	io.WL(types.UDInt(0x1337))
	io.WL(types.UDInt(5))
	io.WL(types.UDInt(1000000))
	io.WL(types.UDInt(0x42000FA2))
	io.WL(types.UDInt(1000000))
	io.WL(types.UDInt(0x42000FA2))
	io.WL(types.USInt(0xA3))

	portPath := packet.Paths(
		path.PortBuild([]byte{t.config.Slot}, 1, true),
		path.LogicalBuild(path.LogicalTypeClassID, 0x02, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	)
	io.WL(utils.Len(portPath))
	io.WL(portPath)

	mr := packet.NewMessageRouter(packet.ServiceForwardOpenLarge, packet.Paths(
		path.LogicalBuild(path.LogicalTypeClassID, 0x06, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	), io.Bytes())

	sd, err := t.SendRRData(packet.NewUCMM(mr), 10)
	if err != nil {
		return err
	}

	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(sd.Packet.Items[1].Data)
	io1 := bufferx.New(rmr.ResponseData)
	io1.RL(&t.connID)
	t.established = true

	return nil
}