package command

import (
	"testing"
)

func TestCheckValid(t *testing.T) {
	// Test valid commands
	validCommands := []Command{
		NOP,
		ListServices,
		ListIdentity,
		ListInterfaces,
		RegisterSession,
		UnRegisterSession,
		SendRRData,
		SendUnitData,
		IndicateStatus,
		Cancel,
	}

	for _, cmd := range validCommands {
		if !CheckValid(cmd) {
			t.Errorf("CheckValid(%v) should return true", cmd)
		}
	}

	// Test invalid commands
	invalidCommands := []Command{
		0x01, // Invalid command
		0x02, // Invalid command
		0xFF, // Invalid command
	}

	for _, cmd := range invalidCommands {
		if CheckValid(cmd) {
			t.Errorf("CheckValid(%v) should return false", cmd)
		}
	}
}

func TestCommandConstants(t *testing.T) {
	// Verify command constants have correct values
	tests := []struct {
		name     string
		command  Command
		expected uint16
	}{
		{"NOP", NOP, 0x00},
		{"ListServices", ListServices, 0x04},
		{"ListIdentity", ListIdentity, 0x63},
		{"ListInterfaces", ListInterfaces, 0x64},
		{"RegisterSession", RegisterSession, 0x65},
		{"UnRegisterSession", UnRegisterSession, 0x66},
		{"SendRRData", SendRRData, 0x6F},
		{"SendUnitData", SendUnitData, 0x70},
		{"IndicateStatus", IndicateStatus, 0x72},
		{"Cancel", Cancel, 0x73},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint16(tt.command) != tt.expected {
				t.Errorf("%s should be 0x%02X, got 0x%02X", tt.name, tt.expected, uint16(tt.command))
			}
		})
	}
}
