package ethernet_ip

import (
	"errors"
	"fmt"
	"math"
	"net"
	"sync"
	"time"

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
	minReconnectDelay time.Duration

	monitor *ConnectionMonitor
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
		requestLock:       new(sync.Mutex),
		config:            config,
		tcpAddr:           tcpAddress,
		readBuf:           make([]byte, 1024*64),
		maxReconnect:      3,
		minReconnectDelay: time.Second * 1,
		monitor:           newConnectionMonitor(),
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
		t.monitor.setState(StateDisconnected, errors.New("max reconnect attempts exceeded"))
		return errors.New("max reconnect attempts exceeded")
	}

	t.monitor.setState(StateReconnecting, nil)

	if t.tcpConn != nil {
		t.tcpConn.Close()
		t.tcpConn = nil
	}

	attempts := t.reconnectAttempts
	t.reset()
	t.reconnectAttempts = attempts

	if t.reconnectAttempts > 0 {
		delay := t.ExponentialBackoff(t.reconnectAttempts)
		time.Sleep(delay)
	}

	dialer := &net.Dialer{
		Timeout: t.config.ConnectTimeout,
	}

	tcpConnection, err := dialer.Dial("tcp", t.tcpAddr.String())
	if err != nil {
		t.reconnectAttempts++
		return err
	}

	tcpConn, ok := tcpConnection.(*net.TCPConn)
	if !ok {
		tcpConnection.Close()
		t.reconnectAttempts++
		return errors.New("failed to cast to TCPConn")
	}

	err = tcpConn.SetKeepAlive(true)
	if err != nil {
		t.reconnectAttempts++
		tcpConn.Close()
		return err
	}

	err = tcpConn.SetReadDeadline(time.Now().Add(t.config.ReadTimeout))
	if err != nil {
		t.reconnectAttempts++
		tcpConn.Close()
		return err
	}

	err = tcpConn.SetWriteDeadline(time.Now().Add(t.config.WriteTimeout))
	if err != nil {
		t.reconnectAttempts++
		tcpConn.Close()
		return err
	}

	t.tcpConn = tcpConn

	if err := t.registerSessionLocked(); err != nil {
		t.reconnectAttempts++
		return err
	}

	t.monitor.stats.recordReconnect()
	t.monitor.setState(StateConnected, nil)
	t.reconnectAttempts = 0
	return nil
}

func (t *EIPTCP) ExponentialBackoff(attempt int) time.Duration {
	if attempt <= 0 {
		if attempt == 0 {
			return t.minReconnectDelay
		}
		return time.Duration(0)
	}
	delay := t.minReconnectDelay * time.Duration(math.Pow(2, float64(attempt)))
	maxDelay := time.Second * 30
	if delay > maxDelay {
		delay = maxDelay
	}
	return delay
}

// IsConnected returns true if the TCP connection is established and session is registered.
func (t *EIPTCP) IsConnected() bool {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.tcpConn != nil && t.established
}

// GetReconnectAttempts returns the number of reconnect attempts made.
func (t *EIPTCP) GetReconnectAttempts() int {
	t.requestLock.Lock()
	defer t.requestLock.Unlock()
	return t.reconnectAttempts
}

// Connect establishes a TCP connection to the device and registers a session.
// It must be called before any read/write operations.
func (t *EIPTCP) Connect() error {
	t.reset()
	t.monitor.setState(StateConnecting, nil)

	dialer := &net.Dialer{
		Timeout: t.config.ConnectTimeout,
	}

	tcpConnection, err := dialer.Dial("tcp", t.tcpAddr.String())
	if err != nil {
		t.monitor.setState(StateDisconnected, err)
		return err
	}

	tcpConn, ok := tcpConnection.(*net.TCPConn)
	if !ok {
		tcpConnection.Close()
		t.monitor.setState(StateDisconnected, errors.New("failed to cast to TCPConn"))
		return errors.New("failed to cast to TCPConn")
	}

	err = tcpConn.SetKeepAlive(true)
	if err != nil {
		tcpConn.Close()
		t.monitor.setState(StateDisconnected, err)
		return err
	}

	err = tcpConn.SetReadDeadline(time.Now().Add(t.config.ReadTimeout))
	if err != nil {
		tcpConn.Close()
		t.monitor.setState(StateDisconnected, err)
		return err
	}

	err = tcpConn.SetWriteDeadline(time.Now().Add(t.config.WriteTimeout))
	if err != nil {
		tcpConn.Close()
		t.monitor.setState(StateDisconnected, err)
		return err
	}

	t.tcpConn = tcpConn

	if err := t.RegisterSession(); err != nil {
		t.tcpConn.Close()
		t.tcpConn = nil
		t.monitor.setState(StateDisconnected, err)
		return err
	}

	t.monitor.stats.recordConnect()
	t.monitor.setState(StateConnected, nil)
	return nil
}

func (t *EIPTCP) write(data []byte) error {
	if err := t.tcpConn.SetWriteDeadline(time.Now().Add(t.config.WriteTimeout)); err != nil {
		return err
	}
	_, err := t.tcpConn.Write(data)
	return err
}

func (t *EIPTCP) read() (*packet.Packet, error) {
	if err := t.tcpConn.SetReadDeadline(time.Now().Add(t.config.ReadTimeout)); err != nil {
		return nil, err
	}
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
	t.monitor.stats.recordDisconnect()
	t.monitor.setState(StateDisconnected, nil)
	t.reset()
	return err
}

// AddConnectionListener adds a listener for connection state changes
func (t *EIPTCP) AddConnectionListener(listener ConnectionEventListener) {
	t.monitor.addListener(listener)
}

// RemoveConnectionListener removes a connection state listener
func (t *EIPTCP) RemoveConnectionListener(listener ConnectionEventListener) {
	t.monitor.removeListener(listener)
}

// GetConnectionState returns the current connection state
func (t *EIPTCP) GetConnectionState() ConnectionState {
	return t.monitor.GetState()
}

// GetConnectionStats returns the connection statistics
func (t *EIPTCP) GetConnectionStats() ConnectionStats {
	return t.monitor.GetStats()
}

// ResetConnectionStats resets all connection statistics
func (t *EIPTCP) ResetConnectionStats() {
	t.monitor.stats.Reset()
}
