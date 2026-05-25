package ethernet_ip

import (
	"errors"
	"fmt"
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
	localRandMu  sync.Mutex
	localRandGen = rand.New(rand.NewSource(0))
)

func randIntn(n int) int {
	localRandMu.Lock()
	defer localRandMu.Unlock()
	return localRandGen.Intn(n)
}

func (t *EIPTCP) request(packet *packet.Packet) (*packet.Packet, error) {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.requestLocked(packet)
}

func (t *EIPTCP) requestLocked(packet *packet.Packet) (*packet.Packet, error) {
	if t.tcpConn == nil {
		return nil, errors.New("connect first")
	}

	b, err := packet.Encode()
	if err != nil {
		return nil, err
	}

	if err := t.write(b); err != nil {
		if err := t.reconnectLocked(); err != nil {
			return nil, err
		}
		if err := t.write(b); err != nil {
			return nil, err
		}
	}

	resp, err := t.read()
	if err != nil {
		if err := t.reconnectLocked(); err != nil {
			return nil, err
		}
		if err := t.write(b); err != nil {
			return nil, err
		}
		return t.read()
	}

	return resp, nil
}

// RegisterSession registers a session with the EtherNet/IP device.
// It is called automatically by Connect(), but can be called manually
// after an UnRegisterSession() to re-establish the session.
func (t *EIPTCP) RegisterSession() error {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.registerSessionLocked()
}

func (t *EIPTCP) registerSessionLocked() error {
	ctx := contextGenerator()
	requestPacket, err := registerSession.New(ctx)
	if err != nil {
		return err
	}

	responsePacket, err := t.requestLocked(requestPacket)
	if err != nil {
		return err
	}

	t.session = responsePacket.SessionHandle
	t.established = t.session != 0
	return nil
}

// UnRegisterSession unregisters the session with the EtherNet/IP device.
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

// ListInterface requests the list of network interfaces from the device.
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

// ListServices requests the list of available services from the device.
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

// ListIdentity requests the identity information from the device.
// This is typically used for device discovery on the network.
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

	spd, err := sendRRData.Decode(responsePacket)
	if err == nil && hasCPFDataItem(spd) {
		return spd, nil
	}

	legacyPacket, legacyErr := sendRRData.NewLegacy(t.session, ctx, cpf)
	if legacyErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, legacyErr
	}
	legacyResponse, legacyErr := t.request(legacyPacket)
	if legacyErr != nil {
		if err != nil {
			return nil, err
		}
		return nil, legacyErr
	}
	return sendRRData.Decode(legacyResponse)
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
	if err != nil {
		return nil, err
	}
	if spd != nil && spd.Packet != nil {
		itemIdx := findCommonPacketFormatDataItem(spd.Packet.Items)
		if itemIdx < 0 {
			return spd, errors.New("unexpected specific data packet item count")
		}
		item := &spd.Packet.Items[itemIdx]
		if len(item.Data) >= 2 {
			item.Data = item.Data[2:]
		}
	}
	return spd, nil
}

func findCommonPacketFormatDataItem(items []packet.CommonPacketFormatItem) int {
	if len(items) == 0 {
		return -1
	}
	if len(items) == 1 {
		return 0
	}
	for i := range items {
		if items[i].TypeID == packet.ItemIDUnconnectedMessage || items[i].TypeID == packet.ItemIDConnectedTransportPacket {
			return i
		}
	}
	return 1
}

func hasCPFDataItem(spd *packet.SpecificData) bool {
	if spd == nil || spd.Packet == nil {
		return false
	}
	return findCommonPacketFormatDataItem(spd.Packet.Items) >= 0
}

// Send sends a MessageRouter request and returns the response.
// This is a low-level method used internally by higher-level operations.
func (t *EIPTCP) Send(mr *packet.MessageRouterRequest) (*packet.SpecificData, error) {
	if t.connID != 0 {
		t.seqNum += 1
		return t.SendUnitData(packet.NewCMM(t.connID, t.seqNum, mr))
	}
	return t.SendRRData(packet.NewUCMM(mr), 10)
}

// ForwardOpen establishes a forward open connection to the device.
// This is used for time-critical communications that require a dedicated
// connection path. Call ForwardClose to close the connection when done.
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

	itemIdx := findCommonPacketFormatDataItem(sd.Packet.Items)
	if itemIdx < 0 {
		return errors.New("unexpected specific data packet item count")
	}
	item := &sd.Packet.Items[itemIdx]

	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(item.Data)
	io1 := bufferx.New(rmr.ResponseData)
	io1.RL(&t.connID)
	t.established = true

	return nil
}

// ReadClass2Attribute 使用 Get Attribute Single 服务读取 Class 2 对象的属性
// attrID: 属性ID (1-12)，对应 cpppo 服务器的标签
func (t *EIPTCP) ReadClass2Attribute(attrID int) ([]byte, error) {
	// CIP Get Attribute Single (0x0E)
	// Path: Class 2, Instance 1, Attribute attrID
	pathData := []byte{
		0x20, 0x02, // Class ID: 2
		0x24, 0x01, // Instance ID: 1
		0x30, byte(attrID), // Attribute ID
	}

	mr := packet.NewMessageRouter(packet.ServiceGetAttributeSingle, pathData, nil)
	response, err := t.Send(mr)
	if err != nil {
		return nil, err
	}

	if response == nil || response.Packet == nil {
		return nil, errors.New("空响应")
	}

	itemIdx := findCommonPacketFormatDataItem(response.Packet.Items)
	if itemIdx < 0 {
		return nil, errors.New("未找到 CIP 响应数据")
	}

	item := &response.Packet.Items[itemIdx]
	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(item.Data)

	if rmr.GeneralStatus != 0 {
		return nil, fmt.Errorf("CIP error: 0x%02X", rmr.GeneralStatus)
	}

	return rmr.ResponseData, nil
}

// ForwardClose closes a forward open connection established by ForwardOpen.
// This releases resources on both the client and the PLC device.
func (t *EIPTCP) ForwardClose() error {
	if t.connID == 0 {
		return errors.New("no forward open connection to close")
	}

	io := bufferx.NewWithCapacity(16)
	io.WL(t.connID)

	portPath := packet.Paths(
		path.PortBuild([]byte{t.config.Slot}, 1, true),
		path.LogicalBuild(path.LogicalTypeClassID, 0x02, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	)
	io.WL(utils.Len(portPath))
	io.WL(portPath)

	mr := packet.NewMessageRouter(packet.ServiceForwardClose, packet.Paths(
		path.LogicalBuild(path.LogicalTypeClassID, 0x06, true),
		path.LogicalBuild(path.LogicalTypeInstanceID, 0x01, true),
	), io.Bytes())

	_, err := t.SendRRData(packet.NewUCMM(mr), 10)

	t.connID = 0
	t.seqNum = 0

	return err
}

// WriteClass2Attribute 使用 Set Attribute Single 服务写入 Class 2 对象的属性
// attrID: 属性ID (1-12)，对应 cpppo 服务器的标签
// value: 要写入的值（字节数组格式）
func (t *EIPTCP) WriteClass2Attribute(attrID int, value []byte) error {
	// CIP Set Attribute Single (0x10)
	// Path: Class 2, Instance 1, Attribute attrID
	pathData := []byte{
		0x20, 0x02, // Class ID: 2
		0x24, 0x01, // Instance ID: 1
		0x30, byte(attrID), // Attribute ID
	}

	mr := packet.NewMessageRouter(packet.ServiceSetAttributeSingle, pathData, value)
	response, err := t.Send(mr)
	if err != nil {
		return err
	}

	if response == nil || response.Packet == nil {
		return errors.New("空响应")
	}

	itemIdx := findCommonPacketFormatDataItem(response.Packet.Items)
	if itemIdx < 0 {
		return errors.New("未找到响应数据")
	}

	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(response.Packet.Items[itemIdx].Data)

	if rmr.GeneralStatus != 0 {
		return fmt.Errorf("写入失败，状态码: 0x%02X", rmr.GeneralStatus)
	}

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

	itemIdx := findCommonPacketFormatDataItem(sd.Packet.Items)
	if itemIdx < 0 {
		return errors.New("unexpected specific data packet item count")
	}
	item := &sd.Packet.Items[itemIdx]

	rmr := &packet.MessageRouterResponse{}
	rmr.Decode(item.Data)
	io1 := bufferx.New(rmr.ResponseData)
	io1.RL(&t.connID)
	t.established = true

	return nil
}
