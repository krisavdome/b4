package mangle

import (
	"encoding/binary"
	"time"
)

// SectionConfig holds configuration for packet processing
type SectionConfig struct {
	// TLS settings
	TLSEnabled     bool
	FakeSNI        bool
	FakeSNISeqLen  uint
	FakeSNIType    int
	FakeCustomPkt  []byte
	FakingStrategy uint32
	FakingTTL      uint8
	FakeseqOffset  uint32

	// Fragmentation settings
	FragStrategy   int
	FragSNIReverse bool
	FragSNIFaked   bool
	FragMiddleSNI  bool
	FragSNIPos     int

	// Delays
	Seg2Delay uint

	// Window size
	FKWinsize uint16

	// Domain matching (would integrate with your trie/suffix set)
	MatchDomain func(string) bool
}

// ProcessTCPPacket processes a TCP packet and returns verdict
func ProcessTCPPacket(packet []byte, config *SectionConfig, sender func([]byte) error) (int, error) {
	if len(packet) < 20 {
		return PktAccept, nil
	}

	// Parse packet structure
	ipVersion := packet[0] >> 4
	isIPv6 := ipVersion == 6

	var ipHeaderLen int
	var tcpStart int

	if isIPv6 {
		ipHeaderLen = 40
		tcpStart = 40
		if len(packet) < 60 {
			return PktAccept, nil
		}
	} else {
		ihl := int(packet[0]&0x0F) * 4
		ipHeaderLen = ihl
		tcpStart = ihl
		if len(packet) < ihl+20 {
			return PktAccept, nil
		}
	}

	// Check destination port (443 for HTTPS)
	dstPort := binary.BigEndian.Uint16(packet[tcpStart+2 : tcpStart+4])
	if dstPort != 443 {
		return PktAccept, nil
	}

	// Get TCP flags
	flags := packet[tcpStart+13]
	isSYN := flags&0x02 != 0

	// Skip SYN packets (no payload yet)
	if isSYN {
		return PktContinue, nil
	}

	if !config.TLSEnabled {
		return PktContinue, nil
	}

	// Get TCP data offset
	dataOffset := int((packet[tcpStart+12] >> 4) * 4)
	payloadStart := tcpStart + dataOffset

	if len(packet) <= payloadStart {
		return PktAccept, nil
	}

	payload := packet[payloadStart:]

	// Parse TLS to find SNI (you would integrate with your sni package here)
	sniOffset, sniLen := findSNIInTLS(payload)
	if sniOffset < 0 {
		return PktContinue, nil
	}

	sni := string(payload[sniOffset : sniOffset+sniLen])

	// Check if domain matches target
	if config.MatchDomain != nil && !config.MatchDomain(sni) {
		return PktContinue, nil
	}

	// Target SNI detected - apply DPI bypass
	return processTLSTarget(packet, config, sender, payloadStart, sniOffset, sniLen)
}

// processTLSTarget applies fragmentation and fake packet strategies
func processTLSTarget(packet []byte, config *SectionConfig, sender func([]byte) error,
	payloadStart, sniOffset, sniLen int) (int, error) {

	// Create a copy to modify
	pkt := make([]byte, len(packet))
	copy(pkt, packet)

	ipVersion := pkt[0] >> 4
	isIPv6 := ipVersion == 6

	var ipHeaderLen int
	var tcpStart int

	if isIPv6 {
		ipHeaderLen = 40
		tcpStart = 40
	} else {
		ipHeaderLen = int(pkt[0]&0x0F) * 4
		tcpStart = ipHeaderLen
	}

	// Modify TCP window size if configured
	if config.FKWinsize > 0 {
		binary.BigEndian.PutUint16(pkt[tcpStart+14:tcpStart+16], config.FKWinsize)
		SetTCPChecksum(pkt, isIPv6)
	}

	// Send fake packets if enabled
	if config.FakeSNI {
		fakeType := FakeType{
			Type:        config.FakeSNIType,
			FakeData:    config.FakeCustomPkt,
			SequenceLen: config.FakeSNISeqLen,
			Seg2Delay:   config.Seg2Delay,
			Strategy: FailingStrategy{
				Strategy:      config.FakingStrategy,
				FakingTTL:     config.FakingTTL,
				RandseqOffset: config.FakeseqOffset,
			},
		}

		if err := SendFakeSequence(pkt, fakeType, sender); err != nil {
			return PktAccept, err
		}
	}

	// Apply fragmentation strategy
	switch config.FragStrategy {
	case FragStratTCP, FragStratIP:
		return fragmentAndSendTLS(pkt, config, sender, payloadStart, sniOffset, sniLen)
	default:
		// No fragmentation, just send
		if err := sender(pkt); err != nil {
			return PktAccept, err
		}
		return PktDrop, nil
	}
}

// fragmentAndSendTLS fragments the packet around SNI position
func fragmentAndSendTLS(packet []byte, config *SectionConfig, sender func([]byte) error,
	payloadStart, sniOffset, sniLen int) (int, error) {

	// Parse IP header length from the packet
	ipVersion := packet[0] >> 4
	var ipHeaderLen int

	if ipVersion == 6 {
		ipHeaderLen = 40 // IPv6 has fixed 40-byte header
	} else {
		ipHeaderLen = int(packet[0]&0x0F) * 4 // IPv4 IHL field * 4
	}

	// Calculate fragmentation positions relative to TCP payload start
	var positions []int

	// Add configured position
	if config.FragSNIPos > 0 {
		positions = append(positions, config.FragSNIPos)
	}

	// Add middle of SNI position
	if config.FragMiddleSNI {
		midOffset := sniOffset + sniLen/2
		positions = append(positions, midOffset)
	}

	// Sort positions if we have multiple
	if len(positions) > 1 {
		// Simple bubble sort for small arrays
		for i := 0; i < len(positions)-1; i++ {
			for j := i + 1; j < len(positions); j++ {
				if positions[i] > positions[j] {
					positions[i], positions[j] = positions[j], positions[i]
				}
			}
		}
	}

	// For IP fragmentation, adjust positions to be relative to IP payload
	// and round to 8-byte boundaries
	if config.FragStrategy == FragStratIP {
		// TCP header length = payloadStart - ipHeaderLen
		tcpHeaderLen := payloadStart - ipHeaderLen

		for i := range positions {
			// Convert from TCP payload offset to IP payload offset
			// IP payload = TCP header + TCP payload
			positions[i] = tcpHeaderLen + positions[i]

			// Round to 8-byte boundary (required for IP fragmentation)
			positions[i] = (positions[i] + 7) &^ 7
		}
	}

	// Create sender wrapper that handles delays
	sendWithDelay := func(pkt []byte) error {
		return sender(pkt)
	}

	// If we need to send fake packets on fragments
	if config.FragSNIFaked {
		// This would send fake packets between fragments
		// Implementation depends on your specific needs
	}

	// Fragment and send
	err := FragmentAndSend(packet, positions, config.FragStrategy, config.FragSNIReverse, sendWithDelay)
	if err != nil {
		return PktAccept, err
	}

	return PktDrop, nil
}

// findSNIInTLS searches for SNI in TLS ClientHello
// Returns offset and length, or -1 if not found
// This is a simplified version - integrate with your sni package
func findSNIInTLS(data []byte) (offset int, length int) {
	// TLS Record: type(1) + version(2) + length(2) + handshake
	if len(data) < 5 {
		return -1, 0
	}

	// Check for Handshake (0x16)
	if data[0] != 0x16 {
		return -1, 0
	}

	// Check TLS version (0x03 0x01 for TLS 1.0, 0x03 0x03 for TLS 1.2, etc)
	if data[1] != 0x03 {
		return -1, 0
	}

	recordLen := int(binary.BigEndian.Uint16(data[3:5]))
	if len(data) < 5+recordLen {
		recordLen = len(data) - 5
	}

	handshake := data[5 : 5+recordLen]
	if len(handshake) < 1 {
		return -1, 0
	}

	// Check for ClientHello (0x01)
	if handshake[0] != 0x01 {
		return -1, 0
	}

	// Skip handshake header: type(1) + length(3) + version(2) + random(32)
	if len(handshake) < 38 {
		return -1, 0
	}

	pos := 38

	// Session ID length
	if pos >= len(handshake) {
		return -1, 0
	}
	sessionIDLen := int(handshake[pos])
	pos++
	pos += sessionIDLen

	// Cipher suites length
	if pos+2 > len(handshake) {
		return -1, 0
	}
	cipherSuitesLen := int(binary.BigEndian.Uint16(handshake[pos : pos+2]))
	pos += 2 + cipherSuitesLen

	// Compression methods length
	if pos >= len(handshake) {
		return -1, 0
	}
	compMethodsLen := int(handshake[pos])
	pos++
	pos += compMethodsLen

	// Extensions length
	if pos+2 > len(handshake) {
		return -1, 0
	}
	extensionsLen := int(binary.BigEndian.Uint16(handshake[pos : pos+2]))
	pos += 2

	extensionsEnd := pos + extensionsLen
	if extensionsEnd > len(handshake) {
		extensionsEnd = len(handshake)
	}

	// Parse extensions
	for pos+4 <= extensionsEnd {
		extType := binary.BigEndian.Uint16(handshake[pos : pos+2])
		extLen := int(binary.BigEndian.Uint16(handshake[pos+2 : pos+4]))
		pos += 4

		if pos+extLen > extensionsEnd {
			break
		}

		// Server Name extension (type 0)
		if extType == 0 {
			extData := handshake[pos : pos+extLen]
			if len(extData) < 2 {
				break
			}

			listLen := int(binary.BigEndian.Uint16(extData[0:2]))
			if len(extData) < 2+listLen {
				break
			}

			serverNameList := extData[2 : 2+listLen]
			if len(serverNameList) < 3 {
				break
			}

			// Name type (0 for hostname)
			if serverNameList[0] == 0 {
				nameLen := int(binary.BigEndian.Uint16(serverNameList[1:3]))
				if len(serverNameList) >= 3+nameLen {
					// Calculate offset in original data
					sniStart := 5 + pos + 3 // 5 for TLS record header, pos for position in handshake, 3 for list header
					return sniStart, nameLen
				}
			}
		}

		pos += extLen
	}

	return -1, 0
}

// DelayedSender creates a sender that delays packet transmission
func DelayedSender(baseSender func([]byte) error, delay time.Duration) func([]byte) error {
	return func(pkt []byte) error {
		if delay > 0 {
			time.Sleep(delay)
		}
		return baseSender(pkt)
	}
}
