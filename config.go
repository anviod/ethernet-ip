package ethernet_ip

import (
	"time"

	"github.com/anviod/ethernet-ip/types"
)

// Config holds configuration parameters for EtherNet/IP connections.
type Config struct {
	// TCPPort is the TCP port for EtherNet/IP communication (default: 0xAF12 = 44818)
	TCPPort uint16
	// UDPPort is the UDP port for EtherNet/IP communication (default: 0xAF12 = 44818)
	UDPPort uint16
	// Slot specifies the controller slot number in the PLC chassis
	Slot uint8
	// TimeTick is the connection time tick value in milliseconds
	TimeTick types.USInt
	// TimeTickOut is the connection timeout in TimeTick units
	TimeTickOut types.USInt
	// ConnectTimeout is the timeout for TCP connection establishment
	ConnectTimeout time.Duration
	// ReadTimeout is the timeout for read operations
	ReadTimeout time.Duration
	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration
	// OperationTimeout is the timeout for general EIP operations
	OperationTimeout time.Duration
}

// DefaultConfig returns a Config with default values suitable for most PLCs.
func DefaultConfig() *Config {
	cfg := &Config{}
	cfg.TCPPort = 0xAF12
	cfg.UDPPort = 0xAF12
	cfg.Slot = 0
	cfg.TimeTick = types.USInt(3)
	cfg.TimeTickOut = types.USInt(250)
	cfg.ConnectTimeout = time.Second * 30
	cfg.ReadTimeout = time.Second * 10
	cfg.WriteTimeout = time.Second * 10
	cfg.OperationTimeout = time.Second * 15
	return cfg
}
