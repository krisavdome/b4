package mangle

import (
	"encoding/binary"
	"errors"
	"math/rand"
)

var (
	ErrPacketTooShort = errors.New("packet too short")
	ErrInvalidOffset  = errors.New("invalid fragment offset")
)

// SplitTCPPayload splits a TCP packet at the specified payload offset
// Returns two complete packets with updated sequence numbers and checksums
func SplitTCPPayload(packet []byte, payloadOffset int) (frag1, frag2 []byte, err error) {
	// Determine IP version
	if len(packet) == 0 {
		return nil, nil, ErrPacketTooShort
	}

	ipVersion := packet[0] >> 4
	isIPv6 := ipVersion == 6

	var ipHeaderLen int
	var tcpHeaderStart int
	var totalLen int

	if isIPv6 {
		// IPv6 has fixed 40-byte header
		ipHeaderLen = 40
		tcpHeaderStart = 40

		if len(packet) < 40 {
			return nil, nil, ErrPacketTooShort
		}

		payloadLen := int(binary.BigEndian.Uint16(packet[4:6]))
		totalLen = 40 + payloadLen
	} else {
		// IPv4
		if len(packet) < 20 {
			return nil, nil, ErrPacketTooShort
		}

		ihl := int(packet[0]&0x0F) * 4
		ipHeaderLen = ihl
		tcpHeaderStart = ihl
		totalLen = int(binary.BigEndian.Uint16(packet[2:4]))
	}

	if len(packet) < totalLen {
		totalLen = len(packet)
	}

	// Parse TCP header
	if len(packet) < tcpHeaderStart+20 {
		return nil, nil, ErrPacketTooShort
	}

	tcpHeader := packet[tcpHeaderStart:]
	dataOffset := int((tcpHeader[12] >> 4) * 4)

	if len(tcpHeader) < dataOffset {
		return nil, nil, ErrPacketTooShort
	}

	payloadStart := tcpHeaderStart + dataOffset
	payload := packet[payloadStart:totalLen]

	if payloadOffset <= 0 || payloadOffset >= len(payload) {
		return nil, nil, ErrInvalidOffset
	}

	// Calculate fragment sizes
	frag1PayloadLen := payloadOffset
	frag2PayloadLen := len(payload) - payloadOffset

	frag1Len := ipHeaderLen + dataOffset + frag1PayloadLen
	frag2Len := ipHeaderLen + dataOffset + frag2PayloadLen

	// Create fragment 1
	frag1 = make([]byte, frag1Len)
	copy(frag1, packet[:ipHeaderLen])                              // IP header
	copy(frag1[ipHeaderLen:], packet[tcpHeaderStart:payloadStart]) // TCP header
	copy(frag1[payloadStart:], payload[:frag1PayloadLen])          // Payload part 1

	// Create fragment 2
	frag2 = make([]byte, frag2Len)
	copy(frag2, packet[:ipHeaderLen])                              // IP header
	copy(frag2[ipHeaderLen:], packet[tcpHeaderStart:payloadStart]) // TCP header
	copy(frag2[payloadStart:], payload[frag1PayloadLen:])          // Payload part 2

	// Update fragment 1
	if isIPv6 {
		// Update IPv6 payload length
		binary.BigEndian.PutUint16(frag1[4:6], uint16(dataOffset+frag1PayloadLen))
	} else {
		// Update IPv4 total length and ID
		binary.BigEndian.PutUint16(frag1[2:4], uint16(frag1Len))
		binary.BigEndian.PutUint16(frag1[4:6], uint16(rand.Intn(65536))) // Random ID
		SetIPv4Checksum(frag1[:ipHeaderLen])
	}

	// Update TCP checksum for fragment 1
	if err := SetTCPChecksum(frag1, isIPv6); err != nil {
		return nil, nil, err
	}

	// Update fragment 2
	// Get original sequence number
	origSeq := binary.BigEndian.Uint32(packet[tcpHeaderStart+4 : tcpHeaderStart+8])
	newSeq := origSeq + uint32(frag1PayloadLen)

	// Update sequence number in fragment 2
	binary.BigEndian.PutUint32(frag2[tcpHeaderStart+4:tcpHeaderStart+8], newSeq)

	if isIPv6 {
		binary.BigEndian.PutUint16(frag2[4:6], uint16(dataOffset+frag2PayloadLen))
	} else {
		binary.BigEndian.PutUint16(frag2[2:4], uint16(frag2Len))
		binary.BigEndian.PutUint16(frag2[4:6], uint16(rand.Intn(65536)))
		SetIPv4Checksum(frag2[:ipHeaderLen])
	}

	if err := SetTCPChecksum(frag2, isIPv6); err != nil {
		return nil, nil, err
	}

	return frag1, frag2, nil
}

// SplitIPv4Fragment splits an IPv4 packet at the transport payload offset
// offset must be a multiple of 8 bytes for IP fragmentation
func SplitIPv4Fragment(packet []byte, offset int) (frag1, frag2 []byte, err error) {
	if len(packet) < 20 {
		return nil, nil, ErrPacketTooShort
	}

	// Validate offset is multiple of 8
	if offset%8 != 0 {
		return nil, nil, errors.New("offset must be multiple of 8 for IP fragmentation")
	}

	ihl := int(packet[0]&0x0F) * 4
	totalLen := int(binary.BigEndian.Uint16(packet[2:4]))

	if totalLen > len(packet) {
		totalLen = len(packet)
	}

	payload := packet[ihl:totalLen]
	if offset <= 0 || offset >= len(payload) {
		return nil, nil, ErrInvalidOffset
	}

	// Fragment 1: IP header + first part of payload
	frag1Len := ihl + offset
	frag1 = make([]byte, frag1Len)
	copy(frag1, packet[:ihl])
	copy(frag1[ihl:], payload[:offset])

	// Fragment 2: IP header + rest of payload
	frag2Len := ihl + (len(payload) - offset)
	frag2 = make([]byte, frag2Len)
	copy(frag2, packet[:ihl])
	copy(frag2[ihl:], payload[offset:])

	// Update fragment 1 header
	fragOffset1 := binary.BigEndian.Uint16(packet[6:8])
	fragOffset1 &= 0xE000 // Keep flags
	fragOffset1 |= 0x2000 // Set More Fragments flag

	binary.BigEndian.PutUint16(frag1[2:4], uint16(frag1Len))
	binary.BigEndian.PutUint16(frag1[4:6], uint16(rand.Intn(65536)))
	binary.BigEndian.PutUint16(frag1[6:8], fragOffset1)
	SetIPv4Checksum(frag1)

	// Update fragment 2 header
	origFragOffset := binary.BigEndian.Uint16(packet[6:8])
	fragOffset2 := (origFragOffset & 0xE000) // Preserve original flags

	// If original had MF flag, keep it; otherwise clear
	if origFragOffset&0x2000 != 0 {
		fragOffset2 |= 0x2000
	}

	// Add our offset (in 8-byte units)
	fragOffset2 |= uint16(offset / 8)

	binary.BigEndian.PutUint16(frag2[2:4], uint16(frag2Len))
	binary.BigEndian.PutUint16(frag2[4:6], uint16(rand.Intn(65536)))
	binary.BigEndian.PutUint16(frag2[6:8], fragOffset2)
	SetIPv4Checksum(frag2)

	return frag1, frag2, nil
}

// FragmentAndSend splits a packet at multiple positions and sends them
func FragmentAndSend(packet []byte, positions []int, fragStrat int, reverse bool, sender func([]byte) error) error {
	if len(positions) == 0 {
		return sender(packet)
	}

	// For recursive fragmentation, we work from the end
	pos := positions[0]
	remaining := positions[1:]

	var frag1, frag2 []byte
	var err error

	switch fragStrat {
	case FragStratTCP:
		frag1, frag2, err = SplitTCPPayload(packet, pos)
	case FragStratIP:
		// Round to nearest 8-byte boundary
		alignedPos := (pos + 7) &^ 7
		frag1, frag2, err = SplitIPv4Fragment(packet, alignedPos)
	default:
		return sender(packet)
	}

	if err != nil {
		return err
	}

	if reverse {
		// Send fragment 2 first, then fragment 1
		if err := FragmentAndSend(frag2, remaining, fragStrat, reverse, sender); err != nil {
			return err
		}
		return sender(frag1)
	}

	// Normal order: send fragment 1, then fragment 2
	if err := sender(frag1); err != nil {
		return err
	}
	return FragmentAndSend(frag2, remaining, fragStrat, reverse, sender)
}
