package nfq

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

func ExtractPacketInfoV4(packet []byte) (PacketInfo, bool) {
	if len(packet) < 40 {
		return PacketInfo{}, false
	}
	ipHdrLen := int((packet[0] & 0x0F) * 4)
	if len(packet) < ipHdrLen+20 {
		return PacketInfo{}, false
	}
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen
	payloadLen := len(packet) - payloadStart

	return PacketInfo{
		IPHdrLen:     ipHdrLen,
		TCPHdrLen:    tcpHdrLen,
		PayloadStart: payloadStart,
		PayloadLen:   payloadLen,
		Payload:      packet[payloadStart:],
		Seq0:         binary.BigEndian.Uint32(packet[ipHdrLen+4 : ipHdrLen+8]),
		ID0:          binary.BigEndian.Uint16(packet[4:6]),
		IsIPv6:       false,
	}, true
}

func BuildSegmentV4(packet []byte, pi PacketInfo, payloadSlice []byte, seqOffset uint32, idOffset uint16) []byte {
	segLen := pi.PayloadStart + len(payloadSlice)
	seg := make([]byte, segLen)
	copy(seg[:pi.PayloadStart], packet[:pi.PayloadStart])
	copy(seg[pi.PayloadStart:], payloadSlice)

	binary.BigEndian.PutUint32(seg[pi.IPHdrLen+4:pi.IPHdrLen+8], pi.Seq0+seqOffset)
	binary.BigEndian.PutUint16(seg[4:6], pi.ID0+idOffset)
	binary.BigEndian.PutUint16(seg[2:4], uint16(segLen))

	sock.FixIPv4Checksum(seg[:pi.IPHdrLen])
	sock.FixTCPChecksum(seg)
	return seg
}

func ShuffleSegments(segments []Segment, mode string, r *rand.Rand) {
	switch mode {
	case "full":
		for i := len(segments) - 1; i > 0; i-- {
			j := r.Intn(i + 1)
			segments[i], segments[j] = segments[j], segments[i]
		}
	case "reverse":
		for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
			segments[i], segments[j] = segments[j], segments[i]
		}
	case "middle":
		if len(segments) > 3 {
			middle := segments[1 : len(segments)-1]
			for i := len(middle) - 1; i > 0; i-- {
				j := r.Intn(i + 1)
				middle[i], middle[j] = middle[j], middle[i]
			}
		} else if len(segments) > 1 {
			for i, j := 0, len(segments)-1; i < j; i, j = i+1, j-1 {
				segments[i], segments[j] = segments[j], segments[i]
			}
		}
	}
}

func (w *Worker) SendWithDelayV4(seg []byte, dst net.IP, delayMs int) {
	_ = w.sock.SendIPv4(seg, dst)
	if delayMs > 0 {
		time.Sleep(time.Duration(delayMs) * time.Millisecond)
	}
}

func (w *Worker) SendSegmentsV4(segs [][]byte, dst net.IP, cfg *config.SetConfig) {
	delay := cfg.TCP.Seg2Delay
	if cfg.Fragmentation.ReverseOrder {
		for i := len(segs) - 1; i >= 0; i-- {
			_ = w.sock.SendIPv4(segs[i], dst)
			if i > 0 && delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	} else {
		for i, seg := range segs {
			_ = w.sock.SendIPv4(seg, dst)
			if i < len(segs)-1 && delay > 0 {
				time.Sleep(time.Duration(delay) * time.Millisecond)
			}
		}
	}
}

func SetPSH(seg []byte, ipHdrLen int) {
	seg[ipHdrLen+13] |= 0x08
}

func ClearPSH(seg []byte, ipHdrLen int) {
	seg[ipHdrLen+13] &^= 0x08
}

func GetSNISplitPoints(payload []byte, payloadLen int, middleSNI bool, sniPosition int) []int {
	var splits []int

	if middleSNI {
		if sniStart, sniEnd, ok := locateSNI(payload); ok && sniEnd > sniStart {
			sniLen := sniEnd - sniStart
			splits = append(splits, sniStart)
			if sniLen > 6 {
				splits = append(splits, sniStart+sniLen/2)
			}
			splits = append(splits, sniEnd)
		}
	}

	if len(splits) == 0 && sniPosition > 0 && sniPosition < payloadLen {
		splits = append(splits, sniPosition)
	}

	return splits
}
