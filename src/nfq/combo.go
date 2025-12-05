package nfq

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendComboFragments combines multiple evasion techniques
// Strategy: split at multiple points + send out of order + optional delay
func (w *Worker) sendComboFragments(cfg *config.SetConfig, packet []byte, dst net.IP) {
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

	// Find split points
	splits := []int{1} // Always split after first byte (firstbyte desync)

	// Add pre-SNI extension split if found
	if extSplit := findPreSNIExtensionPoint(payload); extSplit > 1 && extSplit < payloadLen-5 {
		splits = append(splits, extSplit)
	}

	// Add mid-SNI split if found
	if sniStart, sniEnd, ok := locateSNI(payload); ok && sniEnd > sniStart {
		midSNI := sniStart + (sniEnd-sniStart)/2
		if midSNI > splits[len(splits)-1]+2 {
			splits = append(splits, midSNI)
		}
	}

	// Ensure splits are sorted and unique
	splits = uniqueSorted(splits, payloadLen)

	if len(splits) < 2 {
		splits = []int{1, payloadLen / 2}
	}

	// Build segments
	type segment struct {
		data []byte
		seq  uint32
	}

	segments := make([]segment, 0, len(splits)+1)
	prevEnd := 0

	for i, split := range splits {
		if split <= prevEnd || split >= payloadLen {
			continue
		}

		segLen := payloadStart + (split - prevEnd)
		seg := make([]byte, segLen)
		copy(seg[:payloadStart], packet[:payloadStart])
		copy(seg[payloadStart:], payload[prevEnd:split])

		binary.BigEndian.PutUint32(seg[ipHdrLen+4:ipHdrLen+8], seq0+uint32(prevEnd))
		binary.BigEndian.PutUint16(seg[4:6], id0+uint16(i))
		binary.BigEndian.PutUint16(seg[2:4], uint16(segLen))
		seg[ipHdrLen+13] &^= 0x08 // Clear PSH

		sock.FixIPv4Checksum(seg[:ipHdrLen])
		sock.FixTCPChecksum(seg)

		segments = append(segments, segment{data: seg, seq: seq0 + uint32(prevEnd)})
		prevEnd = split
	}

	// Final segment
	if prevEnd < payloadLen {
		segLen := payloadStart + (payloadLen - prevEnd)
		seg := make([]byte, segLen)
		copy(seg[:payloadStart], packet[:payloadStart])
		copy(seg[payloadStart:], payload[prevEnd:])

		binary.BigEndian.PutUint32(seg[ipHdrLen+4:ipHdrLen+8], seq0+uint32(prevEnd))
		binary.BigEndian.PutUint16(seg[4:6], id0+uint16(len(segments)))
		binary.BigEndian.PutUint16(seg[2:4], uint16(segLen))
		// Keep PSH on last segment

		sock.FixIPv4Checksum(seg[:ipHdrLen])
		sock.FixTCPChecksum(seg)

		segments = append(segments, segment{data: seg, seq: seq0 + uint32(prevEnd)})
	}

	if len(segments) == 0 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	// Shuffle middle segments, keep edges more stable
	if len(segments) > 3 {
		// Shuffle segments[1:len-1]
		middle := segments[1 : len(segments)-1]
		rand.Shuffle(len(middle), func(i, j int) {
			middle[i], middle[j] = middle[j], middle[i]
		})
	} else if len(segments) > 1 {
		// Just reverse for 2-3 segments
		for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
			segments[i], segments[j] = segments[j], segments[i]
		}
	}

	// Send with variable delays
	for i, seg := range segments {
		_ = w.sock.SendIPv4(seg.data, dst)

		if i < len(segments)-1 {
			// First segment gets longer delay (firstbyte desync effect)
			if i == 0 && cfg.TCP.Seg2Delay > 0 {
				time.Sleep(time.Duration(cfg.TCP.Seg2Delay) * time.Millisecond)
			} else {
				// Small jitter between others
				time.Sleep(time.Duration(rand.Intn(1000)) * time.Microsecond)
			}
		}
	}
}

func uniqueSorted(splits []int, maxVal int) []int {
	seen := make(map[int]bool)
	result := make([]int, 0, len(splits))

	for _, s := range splits {
		if s > 0 && s < maxVal && !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	// Simple sort
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j] < result[i] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}
