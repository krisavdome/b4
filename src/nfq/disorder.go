package nfq

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendDisorderFragments - splits and sends in random order without any faking
// DPI expects sequential data; this exploits that assumption
func (w *Worker) sendDisorderFragments(cfg *config.SetConfig, packet []byte, dst net.IP) {
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	if payloadLen < 10 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	payload := packet[payloadStart:]
	seq0 := binary.BigEndian.Uint32(packet[ipHdrLen+4 : ipHdrLen+8])
	id0 := binary.BigEndian.Uint16(packet[4:6])

	// Find optimal split points
	var splits []int

	// Try to find SNI and split around it
	if sniStart, sniEnd, ok := locateSNI(payload); ok && sniEnd > sniStart {
		sniLen := sniEnd - sniStart
		// Split into 3-4 pieces around SNI
		splits = append(splits, sniStart)
		if sniLen > 6 {
			splits = append(splits, sniStart+sniLen/2)
		}
		splits = append(splits, sniEnd)
	} else {
		// Fallback: split at 1, middle, and 3/4
		splits = []int{1, payloadLen / 2, payloadLen * 3 / 4}
	}

	// Filter valid splits
	validSplits := []int{0}
	for _, s := range splits {
		if s > 0 && s < payloadLen {
			validSplits = append(validSplits, s)
		}
	}
	validSplits = append(validSplits, payloadLen)

	// Build segments
	type segment struct {
		data   []byte
		seqOff uint32
		order  int
	}

	segments := make([]segment, 0, len(validSplits)-1)
	for i := 0; i < len(validSplits)-1; i++ {
		start := validSplits[i]
		end := validSplits[i+1]

		segLen := payloadStart + (end - start)
		seg := make([]byte, segLen)
		copy(seg[:payloadStart], packet[:payloadStart])
		copy(seg[payloadStart:], payload[start:end])

		binary.BigEndian.PutUint32(seg[ipHdrLen+4:ipHdrLen+8], seq0+uint32(start))
		binary.BigEndian.PutUint16(seg[4:6], id0+uint16(i))
		binary.BigEndian.PutUint16(seg[2:4], uint16(segLen))

		// Keep PSH only on last segment
		if i < len(validSplits)-2 {
			seg[ipHdrLen+13] &^= 0x08 // Clear PSH
		}

		sock.FixIPv4Checksum(seg[:ipHdrLen])
		sock.FixTCPChecksum(seg)

		segments = append(segments, segment{data: seg, seqOff: uint32(start), order: i})
	}

	// Shuffle order (but keep first segment sometimes first for compatibility)
	if len(segments) > 2 {
		// Fisher-Yates shuffle for middle segments
		for i := len(segments) - 1; i > 1; i-- {
			j := 1 + rand.Intn(i)
			segments[i], segments[j] = segments[j], segments[i]
		}
		// 50% chance to also shuffle first segment
		if rand.Intn(2) == 0 {
			j := rand.Intn(len(segments))
			segments[0], segments[j] = segments[j], segments[0]
		}
	} else if len(segments) == 2 {
		// Always send second first for 2-segment case
		segments[0], segments[1] = segments[1], segments[0]
	}

	// Send with small jitter delays
	for i, seg := range segments {
		_ = w.sock.SendIPv4(seg.data, dst)
		if i < len(segments)-1 {
			// Random delay 0-2ms to look like natural jitter
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Microsecond)
		}
	}
}
