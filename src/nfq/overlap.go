package nfq

import (
	"bytes"
	"net"
	"time"

	"github.com/daniellavrushin/b4/config"
	"github.com/daniellavrushin/b4/sock"
)

// sendOverlapFragments exploits TCP segment overlap behavior
func (w *Worker) sendOverlapFragments(cfg *config.SetConfig, packet []byte, dst net.IP) {
	pi, ok := ExtractPacketInfoV4(packet)
	if !ok || pi.PayloadLen < 20 {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	sniStart, sniEnd, ok := locateSNI(pi.Payload)
	if !ok || sniEnd <= sniStart || sniEnd > pi.PayloadLen || sniStart < 0 {
		w.sendTCPFragments(cfg, packet, dst)
		return
	}

	// Segment 1: From before SNI to end, contains REAL SNI (sent FIRST - server keeps this)
	overlapStart := sniStart - 4
	if overlapStart < 0 {
		overlapStart = 0
	}

	seg1 := BuildSegmentV4(packet, pi, pi.Payload[overlapStart:], uint32(overlapStart), 0)

	// Segment 2: From start through SNI, with FAKE SNI (sent SECOND - DPI sees, server discards overlap)
	seg2End := sniEnd + 4
	if seg2End > pi.PayloadLen {
		seg2End = pi.PayloadLen
	}

	seg2 := BuildSegmentV4(packet, pi, pi.Payload[:seg2End], 0, 1)

	// Inject fake SNI
	sniLen := sniEnd - sniStart
	fakeDomains := cfg.Fragmentation.Overlap.FakeSNIs
	if len(fakeDomains) == 0 {
		fakeDomains = config.DefaultSetConfig.Fragmentation.Overlap.FakeSNIs
	}
	if len(fakeDomains) == 0 {
		w.sendTCPFragments(cfg, packet, dst)
		return
	}
	fakeSNI := []byte(fakeDomains[pi.Seq0%uint32(len(fakeDomains))])
	if len(fakeSNI) < sniLen {
		fakeSNI = append(fakeSNI, bytes.Repeat([]byte{'.'}, sniLen-len(fakeSNI))...)
	}

	// Validate bounds before copy to prevent panic
	destStart := pi.PayloadStart + sniStart
	destEnd := pi.PayloadStart + sniEnd
	if destStart < 0 || destEnd > len(seg2) || destStart > destEnd || sniLen > len(fakeSNI) {
		w.sendTCPFragments(cfg, packet, dst)
		return
	}
	copy(seg2[destStart:destEnd], fakeSNI[:sniLen])

	ClearPSH(seg2, pi.IPHdrLen)
	sock.FixTCPChecksum(seg2)

	delay := cfg.TCP.Seg2Delay

	// REAL first (server keeps), then FAKE (DPI sees but server discards)
	_ = w.sock.SendIPv4(seg1, dst)
	if delay > 0 {
		time.Sleep(time.Duration(delay) * time.Millisecond)
	}
	_ = w.sock.SendIPv4(seg2, dst)
}
