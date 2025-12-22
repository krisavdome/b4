package nfq

import (
	"net"
	"sort"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
	"github.com/daniellavrushin/b4/utils"
)

// sendComboFragments combines multiple evasion techniques
// Strategy: split at multiple points + send out of order + optional delay
func (w *Worker) sendComboFragments(cfg *config.SetConfig, packet []byte, dst net.IP) {
	pi, ok := ExtractPacketInfoV4(packet)
	if !ok || pi.PayloadLen < 20 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	combo := &cfg.Fragmentation.Combo
	splits := []int{}

	if combo.FirstByteSplit {
		splits = append(splits, 1)
	}

	if combo.ExtensionSplit {
		if extSplit := findPreSNIExtensionPoint(pi.Payload); extSplit > 1 && extSplit < pi.PayloadLen-5 {
			splits = append(splits, extSplit)
		}
	}

	if cfg.Fragmentation.MiddleSNI {
		if sniStart, sniEnd, ok := locateSNI(pi.Payload); ok && sniEnd > sniStart {
			sniLen := sniEnd - sniStart
			if sniStart > 2 {
				splits = append(splits, sniStart-1)
			}
			splits = append(splits, sniStart+sniLen/2)
			if sniLen > 15 {
				splits = append(splits, sniStart+sniLen*3/4)
			}
		}
	}

	splits = uniqueSorted(splits, pi.PayloadLen)
	if len(splits) < 1 {
		splits = []int{pi.PayloadLen / 2}
	}

	segments := make([]Segment, 0, len(splits)+1)
	prevEnd := 0

	for i, splitPos := range splits {
		if splitPos <= prevEnd {
			continue
		}
		seg := BuildSegmentV4(packet, pi, pi.Payload[prevEnd:splitPos], uint32(prevEnd), uint16(i))
		ClearPSH(seg, pi.IPHdrLen)
		sock.FixTCPChecksum(seg)
		segments = append(segments, Segment{Data: seg, Seq: pi.Seq0 + uint32(prevEnd)})
		prevEnd = splitPos
	}

	if prevEnd < pi.PayloadLen {
		seg := BuildSegmentV4(packet, pi, pi.Payload[prevEnd:], uint32(prevEnd), uint16(len(segments)))
		segments = append(segments, Segment{Data: seg, Seq: pi.Seq0 + uint32(prevEnd)})
	}

	if len(segments) == 0 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	r := utils.NewRand()
	ShuffleSegments(segments, combo.ShuffleMode, r)

	// Set PSH flag on highest-sequence segment
	maxSeqIdx := 0
	for i := range segments {
		ClearPSH(segments[i].Data, pi.IPHdrLen)
		sock.FixTCPChecksum(segments[i].Data)
		if segments[i].Seq > segments[maxSeqIdx].Seq {
			maxSeqIdx = i
		}
	}

	SetPSH(segments[maxSeqIdx].Data, pi.IPHdrLen)
	sock.FixTCPChecksum(segments[maxSeqIdx].Data)

	// Send with delays
	firstDelayMs := combo.FirstDelayMs
	if firstDelayMs <= 0 {
		firstDelayMs = 100
	}
	jitterMaxUs := combo.JitterMaxUs
	if jitterMaxUs <= 0 {
		jitterMaxUs = 2000
	}

	for i, seg := range segments {
		_ = w.sock.SendIPv4(seg.Data, dst)

		if i == 0 {
			jitter := r.Intn(firstDelayMs/3 + 1)
			time.Sleep(time.Duration(firstDelayMs+jitter) * time.Millisecond)
		} else if i < len(segments)-1 {
			time.Sleep(time.Duration(r.Intn(jitterMaxUs)) * time.Microsecond)
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

	sort.Ints(result)
	return result
}
