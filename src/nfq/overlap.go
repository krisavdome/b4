package nfq

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendOverlapFragments exploits TCP segment overlap behavior
// First segment contains fake SNI, second segment overlaps and overwrites with real SNI
// Server uses the SECOND segment's data in overlap region (RFC 793 behavior varies, but most use later data)
func (w *Worker) sendOverlapFragments(cfg *config.SetConfig, packet []byte, dst net.IP) {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	if payloadLen < 20 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	payload := packet[payloadStart:]
	seq0 := binary.BigEndian.Uint32(packet[ipHdrLen+4 : ipHdrLen+8])
	id0 := binary.BigEndian.Uint16(packet[4:6])

	// Find SNI position
	sniStart, sniEnd, ok := locateSNI(payload)
	if !ok || sniEnd <= sniStart {
		// Fallback to regular fragmentation
		w.sendTCPFragments(cfg, packet, dst)
		return
	}

	// Segment 1: From start to beyond SNI, but with FAKE SNI in the overlap region
	seg1End := sniEnd + 2
	if seg1End > payloadLen {
		seg1End = payloadLen
	}

	seg1Len := payloadStart + seg1End
	seg1 := make([]byte, seg1Len)
	copy(seg1[:payloadStart], packet[:payloadStart])
	copy(seg1[payloadStart:], payload[:seg1End])

	// Replace SNI with fake in seg1
	fakeSNI := []byte("yandex.ru") // Use a definitely-unblocked domain
	if len(fakeSNI) <= sniEnd-sniStart {
		// Pad fake SNI to same length
		paddedFake := make([]byte, sniEnd-sniStart)
		copy(paddedFake, fakeSNI)
		for i := len(fakeSNI); i < len(paddedFake); i++ {
			paddedFake[i] = '.'
		}
		copy(seg1[payloadStart+sniStart:payloadStart+sniEnd], paddedFake)
	}

	binary.BigEndian.PutUint16(seg1[2:4], uint16(seg1Len))
	seg1[ipHdrLen+13] &^= 0x08 // Clear PSH
	sock.FixIPv4Checksum(seg1[:ipHdrLen])
	sock.FixTCPChecksum(seg1)

	// Segment 2: Starts BEFORE seg1 ends (overlap), contains real SNI, goes to end
	overlapStart := sniStart - 2
	if overlapStart < 0 {
		overlapStart = 0
	}

	seg2Len := payloadStart + (payloadLen - overlapStart)
	seg2 := make([]byte, seg2Len)
	copy(seg2[:payloadStart], packet[:payloadStart])
	copy(seg2[payloadStart:], payload[overlapStart:]) // Contains REAL SNI

	binary.BigEndian.PutUint32(seg2[ipHdrLen+4:ipHdrLen+8], seq0+uint32(overlapStart))
	binary.BigEndian.PutUint16(seg2[4:6], id0+1)
	binary.BigEndian.PutUint16(seg2[2:4], uint16(seg2Len))
	sock.FixIPv4Checksum(seg2[:ipHdrLen])
	sock.FixTCPChecksum(seg2)

	delay := cfg.TCP.Seg2Delay

	// Send seg1 first (with fake SNI), then seg2 (with real SNI that overlaps)
	_ = w.sock.SendIPv4(seg1, dst)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	_ = w.sock.SendIPv4(seg2, dst)
}
