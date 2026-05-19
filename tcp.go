package ethernet_ip

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/anviod/ethernet-ip/bufferx"
	"github.com/anviod/ethernet-ip/command"
	"github.com/anviod/ethernet-ip/messages/packet"
	"github.com/anviod/ethernet-ip/types"
)

// EIPTCP represents a TCP connection to an EtherNet/IP device.
// It provides methods for communicating with PLCs and compatible devices.
type EIPTCP struct {
	config  *Config
	tcpAddr *net.TCPAddr
	tcpConn *net.TCPConn
	session types.UDInt

	established bool
	connID      types.UDInt
	seqNum      types.UInt

	requestLock *sync.Mutex
	readBuf     []byte

	reconnectAttempts int
	maxReconnect      int
}

// NewTCP creates a new EIPTCP instance and resolves the target address.
// The address parameter should be an IP address or hostname.
// If config is nil, DefaultConfig() will be used.
func NewTCP(address string, config *Config) (*EIPTCP, error) {
	if config == nil {
		config = DefaultConfig()
	}

	tcpAddress, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", address, config.TCPPort))
	if err != nil {
		return nil, err
	}

	return &EIPTCP{
		requestLock:  new(sync.Mutex),
		config:       config,
		tcpAddr:      tcpAddress,
		readBuf:      make([]byte, 1024*64),
		maxReconnect: 3,
	}, nil
}

func (t *EIPTCP) reset() {
	t.established = false
	t.connID = 0
	t.seqNum = 0
	t.reconnectAttempts = 0
}

func (t *EIPTCP) reconnect() error {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.reconnectLocked()
}

func (t *EIPTCP) reconnectLocked() error {
	if t.reconnectAttempts >= t.maxReconnect {
		return errors.New("max reconnect attempts exceeded")
	}

	if t.tcpConn != nil {
		t.tcpConn.Close()
		t.tcpConn = nil
	}

	attempts := t.reconnectAttempts
	t.reset()
	t.reconnectAttempts = attempts

	tcpConnection, err := net.DialTCP("tcp", nil, t.tcpAddr)
	if err != nil {
		t.reconnectAttempts++
		return err
	}

	err = tcpConnection.SetKeepAlive(true)
	if err != nil {
		t.reconnectAttempts++
		tcpConnection.Close()
		return err
	}

	t.tcpConn = tcpConnection

	if err := t.registerSessionLocked(); err != nil {
		t.reconnectAttempts++
		return err
	}

	t.reconnectAttempts = 0
	return nil
}

// IsConnected returns true if the TCP connection is established and session is registered.
func (t *EIPTCP) IsConnected() bool {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.tcpConn != nil && t.established
}

// Connect establishes a TCP connection to the device and registers a session.
// It must be called before any read/write operations.
func (t *EIPTCP) Connect() error {
	t.reset()

	tcpConnection, err := net.DialTCP("tcp", nil, t.tcpAddr)
	if err != nil {
		return err
	}

	err = tcpConnection.SetKeepAlive(true)
	if err != nil {
		return err
	}

	t.tcpConn = tcpConnection

	if err := t.RegisterSession(); err != nil {
		return err
	}

	return nil
}

func (t *EIPTCP) write(data []byte) error {
	_, err := t.tcpConn.Write(data)
	return err
}

func (t *EIPTCP) read() (*packet.Packet, error) {
	if t.readBuf == nil {
		t.readBuf = make([]byte, 1024*64)
	}
	length, err := t.tcpConn.Read(t.readBuf)
	if err != nil {
		return nil, err
	}
	return t.parse(t.readBuf[:length])
}

// ReadFromFile reads data from a memory-mapped file
func (t *EIPTCP) ReadFromFile(filePath string) (*packet.Packet, error) {
	return nil, errors.New("memory mapping not implemented")
}

// WriteToFile writes packet data to a file using memory mapping
func (t *EIPTCP) WriteToFile(filePath string, packet *packet.Packet) error {
	return errors.New("memory mapping not implemented")
}

func (t *EIPTCP) parse(buf []byte) (*packet.Packet, error) {
	if len(buf) < 24 {
		return nil, errors.New("invalid packet, length < 24")
	}
	_packet := &packet.Packet{}
	reader := bufferx.NewReader(buf)

	_packet.Command = command.Command(reader.ReadUint16())
	_packet.Length = types.UInt(reader.ReadUint16())
	_packet.SessionHandle = types.UDInt(reader.ReadUint32())
	_packet.Status = types.UDInt(reader.ReadUint32())
	_packet.SenderContext = types.ULInt(reader.ReadUint64())
	_packet.Options = types.UDInt(reader.ReadUint32())

	if reader.Error() != nil {
		return nil, reader.Error()
	}
	if _packet.Options != 0 {
		return nil, errors.New("wrong packet with non-zero option")
	}
	if int(_packet.Length) != reader.Len() {
		if _packet.Length == 0 && reader.Len() > 0 {
			_packet.Length = types.UInt(reader.Len())
		} else {
			return nil, errors.New("wrong packet length")
		}
	}

	_packet.SpecificData = reader.ReadBytes(reader.Len())
	if reader.Error() != nil {
		return nil, reader.Error()
	}
	return _packet, nil
}

// BatchRead reads multiple packets from the connection
func (t *EIPTCP) BatchRead(count int) ([]*packet.Packet, error) {
	results := make([]*packet.Packet, 0, count)
	for i := 0; i < count; i++ {
		p, err := t.read()
		if err != nil {
			return results, err
		}
		results = append(results, p)
	}
	return results, nil
}

// BatchWrite writes multiple packets to the connection
func (t *EIPTCP) BatchWrite(packets []*packet.Packet) error {
	for _, p := range packets {
		data, err := p.Encode()
		if err != nil {
			return err
		}
		if err := t.write(data); err != nil {
			return err
		}
	}
	return nil
}

func (t *EIPTCP) Close() error {
	if t.tcpConn == nil {
		return nil
	}

	_ = t.UnRegisterSession()

	if t.tcpConn == nil {
		return nil
	}

	err := t.tcpConn.Close()
	t.tcpConn = nil
	t.reset()
	return err
}
