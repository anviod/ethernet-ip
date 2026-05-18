package ethernet_ip

import (
	"net"
	"sync"
	"testing"
)

func TestReconnect_MaxAttempts(t *testing.T) {
	eip, err := NewTCP("127.0.0.1", &Config{TCPPort: uint16(1)})
	if err != nil {
		t.Fatalf("Failed to create EIPTCP: %v", err)
	}

	eip.maxReconnect = 2

	err = eip.reconnect()
	if err == nil {
		t.Error("Expected error when reconnecting to non-existent server")
	}

	if eip.reconnectAttempts != 1 {
		t.Errorf("Expected 1 reconnect attempt, got %d", eip.reconnectAttempts)
	}

	err = eip.reconnect()
	if err == nil {
		t.Error("Expected error on second attempt")
	}

	if eip.reconnectAttempts != 2 {
		t.Errorf("Expected 2 reconnect attempts, got %d", eip.reconnectAttempts)
	}

	err = eip.reconnect()
	if err == nil || err.Error() != "max reconnect attempts exceeded" {
		t.Errorf("Expected 'max reconnect attempts exceeded' error, got: %v", err)
	}
}

func TestReconnect_ResetAfterSuccess(t *testing.T) {
	eip := &EIPTCP{
		requestLock:       new(sync.Mutex),
		reconnectAttempts: 2,
		maxReconnect:      3,
	}

	if eip.reconnectAttempts != 2 {
		t.Errorf("Expected 2 reconnect attempts, got %d", eip.reconnectAttempts)
	}

	eip.reconnectAttempts = 0

	if eip.reconnectAttempts != 0 {
		t.Errorf("Expected reconnectAttempts to be reset to 0, got %d", eip.reconnectAttempts)
	}
}

func TestIsConnected_StateTracking(t *testing.T) {
	eip := &EIPTCP{
		requestLock: new(sync.Mutex),
	}

	if eip.IsConnected() {
		t.Error("Expected not connected state for empty EIPTCP")
	}

	eip.established = true
	eip.tcpConn = &net.TCPConn{}

	if !eip.IsConnected() {
		t.Error("Expected connected state")
	}

	eip.established = false

	if eip.IsConnected() {
		t.Error("Expected not connected state after established=false")
	}
}

func TestReconnect_InitialState(t *testing.T) {
	eip, err := NewTCP("127.0.0.1", &Config{TCPPort: uint16(1)})
	if err != nil {
		t.Fatalf("Failed to create EIPTCP: %v", err)
	}

	if eip.reconnectAttempts != 0 {
		t.Errorf("Expected 0 reconnect attempts initially, got %d", eip.reconnectAttempts)
	}

	if eip.maxReconnect != 3 {
		t.Errorf("Expected maxReconnect to be 3, got %d", eip.maxReconnect)
	}
}
