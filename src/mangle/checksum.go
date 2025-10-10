package mangle

import (
	"encoding/binary"
)

// calculateChecksum computes Internet checksum (RFC 1071)
func calculateChecksum(data []byte) uint16 {
	sum := uint32(0)

	// Add 16-bit words
	for i := 0; i < len(data)-1; i += 2 {
		sum += uint32(binary.BigEndian.Uint16(data[i : i+2]))
	}

	// Add remaining byte if odd length
	if len(data)%2 == 1 {
		sum += uint32(data[len(data)-1]) << 8
	}

	// Fold 32-bit sum to 16 bits
	for sum>>16 != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}

	return ^uint16(sum)
}

// SetIPv4Checksum calculates and sets IPv4 header checksum
func SetIPv4Checksum(ipHeader []byte) {
	if len(ipHeader) < 20 {
		return
	}

	// Clear existing checksum
	ipHeader[10] = 0
	ipHeader[11] = 0

	// Calculate IHL (header length in 32-bit words)
	ihl := int(ipHeader[0]&0x0F) * 4

	// Calculate checksum
	checksum := calculateChecksum(ipHeader[:ihl])
	binary.BigEndian.PutUint16(ipHeader[10:12], checksum)
}

// SetTCPChecksum calculates and sets TCP checksum
func SetTCPChecksum(packet []byte, ipv6 bool) error {
	var ipHeaderLen int
	var srcIP, dstIP []byte
	var tcpLen uint16

	if ipv6 {
		// IPv6 header is fixed 40 bytes
		ipHeaderLen = 40
		if len(packet) < ipHeaderLen {
			return ErrPacketTooShort
		}

		srcIP = packet[8:24]
		dstIP = packet[24:40]
		tcpLen = binary.BigEndian.Uint16(packet[4:6]) // payload length
	} else {
		// IPv4
		if len(packet) < 20 {
			return ErrPacketTooShort
		}

		ihl := int(packet[0]&0x0F) * 4
		ipHeaderLen = ihl

		srcIP = packet[12:16]
		dstIP = packet[16:20]
		totalLen := binary.BigEndian.Uint16(packet[2:4])
		tcpLen = totalLen - uint16(ipHeaderLen)
	}

	if len(packet) < ipHeaderLen+20 {
		return ErrPacketTooShort
	}

	// Ensure we don't read beyond packet bounds
	if len(packet) < ipHeaderLen+int(tcpLen) {
		tcpLen = uint16(len(packet) - ipHeaderLen)
	}

	tcpHeader := packet[ipHeaderLen:]

	// Clear existing checksum
	tcpHeader[16] = 0
	tcpHeader[17] = 0

	// Build pseudo-header
	var pseudoHeader []byte
	if ipv6 {
		pseudoHeader = make([]byte, 40)
		copy(pseudoHeader[0:16], srcIP)
		copy(pseudoHeader[16:32], dstIP)
		binary.BigEndian.PutUint32(pseudoHeader[32:36], uint32(tcpLen))
		pseudoHeader[39] = 6 // TCP protocol number
	} else {
		pseudoHeader = make([]byte, 12)
		copy(pseudoHeader[0:4], srcIP)
		copy(pseudoHeader[4:8], dstIP)
		pseudoHeader[8] = 0
		pseudoHeader[9] = 6 // TCP protocol
		binary.BigEndian.PutUint16(pseudoHeader[10:12], tcpLen)
	}

	// Combine pseudo-header and TCP segment for checksum
	checksumData := append(pseudoHeader, tcpHeader[:tcpLen]...)
	checksum := calculateChecksum(checksumData)

	binary.BigEndian.PutUint16(tcpHeader[16:18], checksum)
	return nil
}

// SetUDPChecksum calculates and sets UDP checksum
func SetUDPChecksum(packet []byte, ipv6 bool) error {
	var ipHeaderLen int
	var srcIP, dstIP []byte
	var udpLen uint16

	if ipv6 {
		ipHeaderLen = 40
		if len(packet) < ipHeaderLen {
			return ErrPacketTooShort
		}

		srcIP = packet[8:24]
		dstIP = packet[24:40]
		udpLen = binary.BigEndian.Uint16(packet[4:6])
	} else {
		if len(packet) < 20 {
			return ErrPacketTooShort
		}

		ihl := int(packet[0]&0x0F) * 4
		ipHeaderLen = ihl
		srcIP = packet[12:16]
		dstIP = packet[16:20]

		totalLen := binary.BigEndian.Uint16(packet[2:4])
		udpLen = totalLen - uint16(ipHeaderLen)
	}

	if len(packet) < ipHeaderLen+8 {
		return ErrPacketTooShort
	}

	udpHeader := packet[ipHeaderLen:]

	// Clear existing checksum
	udpHeader[6] = 0
	udpHeader[7] = 0

	// Build pseudo-header
	var pseudoHeader []byte
	if ipv6 {
		pseudoHeader = make([]byte, 40)
		copy(pseudoHeader[0:16], srcIP)
		copy(pseudoHeader[16:32], dstIP)
		binary.BigEndian.PutUint32(pseudoHeader[32:36], uint32(udpLen))
		pseudoHeader[39] = 17 // UDP protocol
	} else {
		pseudoHeader = make([]byte, 12)
		copy(pseudoHeader[0:4], srcIP)
		copy(pseudoHeader[4:8], dstIP)
		pseudoHeader[8] = 0
		pseudoHeader[9] = 17 // UDP protocol
		binary.BigEndian.PutUint16(pseudoHeader[10:12], udpLen)
	}

	checksumData := append(pseudoHeader, udpHeader[:udpLen]...)
	checksum := calculateChecksum(checksumData)

	// UDP checksum of 0 means no checksum
	if checksum == 0 {
		checksum = 0xFFFF
	}

	binary.BigEndian.PutUint16(udpHeader[6:8], checksum)
	return nil
}
