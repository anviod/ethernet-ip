package ethernet_ip

import (
	"reflect"
	"sync"
	"time"
)

type ConnectionStats struct {
	ConnectCount       int
	DisconnectCount    int
	ReconnectCount     int
	LastConnectTime    time.Time
	LastDisconnectTime time.Time
	LastReconnectTime  time.Time
	Lock               *sync.Mutex
}

func newConnectionStats() *ConnectionStats {
	return &ConnectionStats{
		Lock: new(sync.Mutex),
	}
}

func (s *ConnectionStats) recordConnect() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.ConnectCount++
	s.LastConnectTime = time.Now()
}

func (s *ConnectionStats) recordDisconnect() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.DisconnectCount++
	s.LastDisconnectTime = time.Now()
}

func (s *ConnectionStats) recordReconnect() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	s.ReconnectCount++
	s.LastReconnectTime = time.Now()
}

func (s *ConnectionStats) GetStats() ConnectionStats {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	return *s
}

func (s *ConnectionStats) Reset() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	*s = ConnectionStats{
		Lock: new(sync.Mutex),
	}
}

type ConnectionState int

const (
	StateDisconnected ConnectionState = iota
	StateConnecting
	StateConnected
	StateReconnecting
)

func (s ConnectionState) String() string {
	switch s {
	case StateDisconnected:
		return "Disconnected"
	case StateConnecting:
		return "Connecting"
	case StateConnected:
		return "Connected"
	case StateReconnecting:
		return "Reconnecting"
	default:
		return "Unknown"
	}
}

type ConnectionEvent struct {
	Timestamp time.Time
	State     ConnectionState
	Error     error
	Stats     ConnectionStats
}

type ConnectionEventListener func(event ConnectionEvent)

type ConnectionMonitor struct {
	stats        *ConnectionStats
	currentState ConnectionState
	listeners    []ConnectionEventListener
	listenerLock *sync.Mutex
}

func newConnectionMonitor() *ConnectionMonitor {
	return &ConnectionMonitor{
		stats:        newConnectionStats(),
		currentState: StateDisconnected,
		listeners:    make([]ConnectionEventListener, 0),
		listenerLock: new(sync.Mutex),
	}
}

func (m *ConnectionMonitor) setState(state ConnectionState, err error) {
	m.currentState = state

	event := ConnectionEvent{
		Timestamp: time.Now(),
		State:     state,
		Error:     err,
		Stats:     m.stats.GetStats(),
	}

	m.listenerLock.Lock()
	defer m.listenerLock.Unlock()
	for _, listener := range m.listeners {
		go listener(event)
	}
}

func (m *ConnectionMonitor) addListener(listener ConnectionEventListener) {
	m.listenerLock.Lock()
	defer m.listenerLock.Unlock()
	m.listeners = append(m.listeners, listener)
}

func (m *ConnectionMonitor) removeListener(listener ConnectionEventListener) {
	m.listenerLock.Lock()
	defer m.listenerLock.Unlock()
	listenerVal := reflect.ValueOf(listener)
	for i := len(m.listeners) - 1; i >= 0; i-- {
		if reflect.ValueOf(m.listeners[i]).Pointer() == listenerVal.Pointer() {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			break
		}
	}
}

func (m *ConnectionMonitor) GetState() ConnectionState {
	return m.currentState
}

func (m *ConnectionMonitor) GetStats() ConnectionStats {
	return m.stats.GetStats()
}
