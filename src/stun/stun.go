package stun

import (
	"encoding/binary"
)

// IsSTUNMessage checks if the UDP payload is a STUN message
func IsSTUNMessage(data []byte) bool {
	if len(data) < 20 {
		return false
	}

	messageType := binary.BigEndian.Uint16(data[0:2])
	messageLength := binary.BigEndian.Uint16(data[2:4])

	// Check STUN message type bits (first 2 bits must be 0)
	if messageType&0xC000 != 0 {
		return false
	}

	// Check if it's a request (not response or indication)
	if messageType&0x0110 != 0 {
		return false
	}

	// Verify message length matches remaining data
	if len(data) != int(messageLength)+20 {
		return false
	}

	return true
}
