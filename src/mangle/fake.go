package mangle

import (
	"crypto/rand"
	"encoding/binary"
	"math/big"
)

// GenerateFakeSNI creates a fake TLS Client Hello packet
func GenerateFakeSNI(originalPacket []byte, fakeType FakeType) ([]byte, error) {
	// Determine IP version
	if len(originalPacket) == 0 {
		return nil, ErrPacketTooShort
	}

	ipVersion := originalPacket[0] >> 4
	isIPv6 := ipVersion == 6

	var ipHeaderLen int
	var tcpHeaderStart int

	if isIPv6 {
		ipHeaderLen = 40
		tcpHeaderStart = 40

		if len(originalPacket) < 40 {
			return nil, ErrPacketTooShort
		}
	} else {
		if len(originalPacket) < 20 {
			return nil, ErrPacketTooShort
		}

		ihl := int(originalPacket[0]&0x0F) * 4
		ipHeaderLen = ihl
		tcpHeaderStart = ihl
	}

	if len(originalPacket) < tcpHeaderStart+20 {
		return nil, ErrPacketTooShort
	}

	tcpHeader := originalPacket[tcpHeaderStart:]
	tcpHeaderLen := int((tcpHeader[12] >> 4) * 4)

	// Determine payload length
	var payloadLen int
	switch fakeType.Type {
	case FakePayloadRandom:
		if fakeType.FakeLen == 0 {
			// Random length up to 1200 bytes
			n, _ := rand.Int(rand.Reader, big.NewInt(1200))
			payloadLen = int(n.Int64())
		} else {
			payloadLen = int(fakeType.FakeLen)
		}
	case FakePayloadCustom:
		payloadLen = len(fakeType.FakeData)
	default:
		payloadLen = int(fakeType.FakeLen)
	}

	// Create fake packet
	fakeLen := ipHeaderLen + tcpHeaderLen + payloadLen
	fake := make([]byte, fakeLen)

	// Copy IP header
	copy(fake, originalPacket[:ipHeaderLen])

	// Copy TCP header
	copy(fake[ipHeaderLen:], originalPacket[tcpHeaderStart:tcpHeaderStart+tcpHeaderLen])

	// Generate payload
	payloadStart := ipHeaderLen + tcpHeaderLen
	switch fakeType.Type {
	case FakePayloadRandom:
		rand.Read(fake[payloadStart:])
	case FakePayloadCustom:
		copy(fake[payloadStart:], fakeType.FakeData)
	default:
		// Use provided data
		if len(fakeType.FakeData) > 0 {
			copy(fake[payloadStart:], fakeType.FakeData[:min(len(fakeType.FakeData), payloadLen)])
		}
	}

	// Update lengths
	if isIPv6 {
		binary.BigEndian.PutUint16(fake[4:6], uint16(tcpHeaderLen+payloadLen))
	} else {
		binary.BigEndian.PutUint16(fake[2:4], uint16(fakeLen))
		binary.BigEndian.PutUint16(fake[4:6], uint16(randInt(65536))) // Random ID
	}

	// Apply failing strategy
	if err := ApplyFailingStrategy(fake, fakeType.Strategy, ipHeaderLen, tcpHeaderLen); err != nil {
		return nil, err
	}

	return fake, nil
}

// ApplyFailingStrategy applies packet invalidation techniques
func ApplyFailingStrategy(packet []byte, strategy FailingStrategy, ipHeaderLen, tcpHeaderLen int) error {
	if len(packet) < ipHeaderLen+tcpHeaderLen {
		return ErrPacketTooShort
	}

	ipVersion := packet[0] >> 4
	isIPv6 := ipVersion == 6
	tcpStart := ipHeaderLen

	// Apply each strategy in the bitmask
	if strategy.Strategy&FakeStratRandSeq != 0 {
		// Modify sequence number with random offset
		seq := binary.BigEndian.Uint32(packet[tcpStart+4 : tcpStart+8])

		payloadLen := 0
		if isIPv6 {
			payloadLen = int(binary.BigEndian.Uint16(packet[4:6])) - tcpHeaderLen
		} else {
			totalLen := int(binary.BigEndian.Uint16(packet[2:4]))
			payloadLen = totalLen - ipHeaderLen - tcpHeaderLen
		}

		newSeq := seq - strategy.RandseqOffset - uint32(payloadLen)
		binary.BigEndian.PutUint32(packet[tcpStart+4:tcpStart+8], newSeq)
	}

	if strategy.Strategy&FakeStratPastSeq != 0 {
		// Use past sequence number
		seq := binary.BigEndian.Uint32(packet[tcpStart+4 : tcpStart+8])

		payloadLen := 0
		if isIPv6 {
			payloadLen = int(binary.BigEndian.Uint16(packet[4:6])) - tcpHeaderLen
		} else {
			totalLen := int(binary.BigEndian.Uint16(packet[2:4]))
			payloadLen = totalLen - ipHeaderLen - tcpHeaderLen
		}

		newSeq := seq - uint32(payloadLen)
		binary.BigEndian.PutUint32(packet[tcpStart+4:tcpStart+8], newSeq)
	}

	if strategy.Strategy&FakeStratTTL != 0 {
		// Set low TTL
		if isIPv6 {
			packet[7] = strategy.FakingTTL // Hop limit
		} else {
			packet[8] = strategy.FakingTTL // TTL
		}
	}

	if strategy.Strategy&FakeStratTCPMD5 != 0 {
		// Add TCP MD5 option (kind 19, length 18)
		// This requires expanding the TCP header
		// For simplicity, we'll add it if there's room
		optionsSpace := tcpHeaderLen - 20
		if optionsSpace >= 18 {
			optStart := tcpStart + 20
			packet[optStart] = 19   // MD5 option kind
			packet[optStart+1] = 18 // Length
			// Fill with random data
			rand.Read(packet[optStart+2 : optStart+18])
		}
	}

	// Clear fragment offset for IPv4
	if !isIPv6 {
		packet[6] = 0
		packet[7] = 0
	}

	// Recalculate checksums
	if !isIPv6 {
		SetIPv4Checksum(packet[:ipHeaderLen])
	}
	SetTCPChecksum(packet, isIPv6)

	// Invalidate TCP checksum if requested (must be done AFTER calculation)
	if strategy.Strategy&FakeStratTCPCheck != 0 {
		packet[tcpStart+16]++ // Corrupt checksum
	}

	return nil
}

// SendFakeSequence sends multiple fake packets with strategy applied
func SendFakeSequence(originalPacket []byte, fakeType FakeType, sender func([]byte) error) error {
	for i := uint(0); i < fakeType.SequenceLen; i++ {
		fake, err := GenerateFakeSNI(originalPacket, fakeType)
		if err != nil {
			return err
		}

		if err := sender(fake); err != nil {
			return err
		}
	}

	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func randInt(max int) int {
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max)))
	return int(n.Int64())
}
