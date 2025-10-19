package sock

import (
	"encoding/binary"
)

func udpChecksumIPv4(pkt []byte) {
	ihl := int((pkt[0] & 0x0f) << 2)
	udpo := ihl
	ulen := int(binary.BigEndian.Uint16(pkt[udpo+4 : udpo+6]))
	pseudo := make([]byte, 12)
	copy(pseudo[0:4], pkt[12:16])
	copy(pseudo[4:8], pkt[16:20])
	pseudo[9] = 17
	binary.BigEndian.PutUint16(pseudo[10:12], uint16(ulen))
	pkt[udpo+6], pkt[udpo+7] = 0, 0
	sum := uint32(0)
	for i := 0; i < len(pseudo); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(pseudo[i : i+2]))
	}
	udp := pkt[udpo : udpo+ulen]
	for i := 0; i+1 < len(udp); i += 2 {
		sum += uint32(binary.BigEndian.Uint16(udp[i : i+2]))
	}
	if len(udp)%2 == 1 {
		sum += uint32(udp[len(udp)-1]) << 8
	}
	for sum > 0xffff {
		sum = (sum >> 16) + (sum & 0xffff)
	}
	c := ^uint16(sum)
	if c == 0 {
		c = 0xffff
	}
	binary.BigEndian.PutUint16(pkt[udpo+6:udpo+8], c)
}

func ApplyUDPFailingStrategy(packet []byte, strategy string, ttl uint8) {
	if len(packet) < 20 {
		return
	}

	ihl := int((packet[0] & 0x0f) << 2)
	if len(packet) < ihl+8 {
		return
	}

	switch strategy {
	case "ttl":
		packet[8] = ttl // Set IP TTL
		// Clear fragment flags for UDP
		packet[6] = 0
		packet[7] = 0
		FixIPv4Checksum(packet[:ihl])
	case "checksum":
		// Corrupt UDP checksum
		packet[ihl+6] += 1
		packet[ihl+7] += 1
	case "none":
		// Do nothing
	}
}

func BuildFakeUDPFromOriginal(orig []byte, fakeLen int, ttl uint8, strategy string) ([]byte, bool) {
	if len(orig) < 20 || orig[0]>>4 != 4 {
		return nil, false
	}
	ihl := int((orig[0] & 0x0f) << 2)
	if len(orig) < ihl+8 {
		return nil, false
	}

	out := make([]byte, 20+8+fakeLen)
	copy(out, orig[:20])

	// Apply strategy-specific changes
	if strategy == "ttl" {
		out[8] = ttl
		out[6] = 0 // Clear fragment flags
		out[7] = 0
	} else {
		out[8] = 64 // Default TTL
	}

	id := binary.BigEndian.Uint16(out[4:6])
	binary.BigEndian.PutUint16(out[4:6], id+1)

	binary.BigEndian.PutUint16(out[2:4], uint16(20+8+fakeLen))
	copy(out[20:], orig[ihl:ihl+8])
	binary.BigEndian.PutUint16(out[20+4:20+6], uint16(8+fakeLen))

	// Fill with random data instead of zeros for better DPI evasion
	for i := 0; i < fakeLen; i++ {
		out[28+i] = byte(i & 0xFF)
	}

	FixIPv4Checksum(out[:20])
	udpChecksumIPv4(out)

	if strategy == "checksum" {
		// Corrupt checksum after calculation
		out[20+6] ^= 0xFF
		out[20+7] ^= 0xFF
	}

	return out, true
}

func IPv4FragmentUDP(orig []byte, split int) ([][]byte, bool) {
	if len(orig) < 28 || orig[0]>>4 != 4 {
		return nil, false
	}
	ihl := int((orig[0] & 0x0f) << 2)
	if len(orig) < ihl+8 {
		return nil, false
	}
	total := int(binary.BigEndian.Uint16(orig[2:4]))
	if total > len(orig) {
		total = len(orig)
	}
	udp := orig[ihl:total]
	if len(udp) < 8 {
		return nil, false
	}
	payload := udp[8:]
	if split < 1 || split >= len(payload) {
		split = 8
	}
	firstData := 8 + split
	firstDataAligned := firstData - (firstData % 8)
	if firstDataAligned < 8 {
		firstDataAligned = 8
	}
	if firstDataAligned >= len(udp) {
		return nil, false
	}

	id := binary.BigEndian.Uint16(orig[4:6])

	ip1 := make([]byte, 20+firstDataAligned)
	copy(ip1, orig[:20])
	binary.BigEndian.PutUint16(ip1[4:6], id)

	// Set MF (More Fragments) flag - bit 13 (0x2000)
	ip1[6] = 0x20 // MF flag set
	ip1[7] = 0x00 // Offset 0

	binary.BigEndian.PutUint16(ip1[2:4], uint16(20+firstDataAligned))
	copy(ip1[20:], udp[:firstDataAligned])
	FixIPv4Checksum(ip1[:20])

	// Second fragment
	ip2Data := udp[firstDataAligned:]
	ip2 := make([]byte, 20+len(ip2Data))
	copy(ip2, orig[:20])
	binary.BigEndian.PutUint16(ip2[4:6], id)

	// Calculate fragment offset in 8-byte units
	offsetUnits := uint16(firstDataAligned / 8)

	// Fragment offset is 13 bits (bits 3-15), flags are first 3 bits
	fragField := offsetUnits & 0x1FFF // 13-bit offset
	// No MF flag for last fragment (unless original had MF)
	origFragField := binary.BigEndian.Uint16(orig[6:8])
	if origFragField&0x2000 != 0 { // Preserve original MF if set
		fragField |= 0x2000
	}

	binary.BigEndian.PutUint16(ip2[6:8], fragField)
	binary.BigEndian.PutUint16(ip2[2:4], uint16(20+len(ip2Data)))
	copy(ip2[20:], ip2Data)
	FixIPv4Checksum(ip2[:20])

	return [][]byte{ip2, ip1}, true
}
