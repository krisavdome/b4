package nfq

import (
	"bytes"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendOverlapFragmentsV6 - IPv6 version: exploits TCP segment overlap behavior
func (w *Worker) sendOverlapFragmentsV6(cfg *config.SetConfig, packet []byte, dst net.IP) {
	pi, ok := ExtractPacketInfoV6(packet)
	if !ok || pi.PayloadLen < 20 {
		_ = w.sock.SendIPv6(packet, dst)
		return
	}

	sniStart, sniEnd, ok := locateSNI(pi.Payload)
	if !ok || sniEnd <= sniStart || sniEnd > pi.PayloadLen || sniStart < 0 {
		w.sendTCPSegmentsv6(cfg, packet, dst)
		return
	}

	// Segment 1: Contains REAL SNI (sent FIRST - server keeps)
	overlapStart := sniStart - 4
	if overlapStart < 0 {
		overlapStart = 0
	}

	seg1 := BuildSegmentV6(packet, pi, pi.Payload[overlapStart:], uint32(overlapStart))

	// Segment 2: With FAKE SNI (sent SECOND - DPI sees, server discards overlap)
	seg2End := sniEnd + 4
	if seg2End > pi.PayloadLen {
		seg2End = pi.PayloadLen
	}

	seg2 := BuildSegmentV6(packet, pi, pi.Payload[:seg2End], 0)

	sniLen := sniEnd - sniStart
	fakeDomains := cfg.Fragmentation.Overlap.FakeSNIs
	if len(fakeDomains) == 0 {
		fakeDomains = config.DefaultSetConfig.Fragmentation.Overlap.FakeSNIs
	}
	if len(fakeDomains) == 0 {
		w.sendTCPSegmentsv6(cfg, packet, dst)
		return
	}
	fakeSNI := []byte(fakeDomains[pi.Seq0%uint32(len(fakeDomains))])
	if len(fakeSNI) < sniLen {
		fakeSNI = append(fakeSNI, bytes.Repeat([]byte{'.'}, sniLen-len(fakeSNI))...)
	}

	destStart := pi.PayloadStart + sniStart
	destEnd := pi.PayloadStart + sniEnd
	if destStart < 0 || destEnd > len(seg2) || destStart > destEnd || sniLen > len(fakeSNI) {
		w.sendTCPSegmentsv6(cfg, packet, dst)
		return
	}
	copy(seg2[destStart:destEnd], fakeSNI[:sniLen])

	ClearPSH(seg2, pi.IPHdrLen)
	sock.FixTCPChecksumV6(seg2)

	delay := cfg.TCP.Seg2Delay

	_ = w.sock.SendIPv6(seg1, dst)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	_ = w.sock.SendIPv6(seg2, dst)
}
