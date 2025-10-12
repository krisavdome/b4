package nfq

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/daniellavrushin/b4/log"
	"github.com/daniellavrushin/b4/sni"
	"github.com/daniellavrushin/b4/sock"
	"github.com/florianl/go-nfqueue"
)

func (w *Worker) Start() error {
	s, err := sock.NewSenderWithMark(int(w.cfg.Mark))
	if err != nil {
		return err
	}
	w.sock = s
	w.frag = &sock.Fragmenter{}

	c := nfqueue.Config{
		NfQueue:      w.qnum,
		MaxPacketLen: 0xffff,
		MaxQueueLen:  4096,
		Copymode:     nfqueue.NfQnlCopyPacket,
	}
	q, err := nfqueue.Open(&c)
	if err != nil {
		return err
	}
	w.q = q

	w.wg.Add(1)
	go w.gc()

	go func() {
		pid := os.Getpid()
		log.Infof("NFQ bound pid=%d queue=%d", pid, w.qnum)
		_ = q.RegisterWithErrorFunc(w.ctx, func(a nfqueue.Attribute) int {
			if a.PacketID == nil || a.Payload == nil || len(*a.Payload) == 0 {
				return 0
			}
			id := *a.PacketID
			raw := *a.Payload

			v := raw[0] >> 4
			if v != 4 && v != 6 {
				_ = q.SetVerdict(id, nfqueue.NfAccept)
				return 0
			}
			var proto uint8
			var src, dst net.IP
			var ihl int
			if v == 4 {
				if len(raw) < 20 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				ihl = int(raw[0]&0x0f) * 4
				if len(raw) < ihl {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				proto = raw[9]
				src = net.IP(raw[12:16])
				dst = net.IP(raw[16:20])
			} else {
				if len(raw) < 40 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				ihl = 40
				proto = raw[6]
				src = net.IP(raw[8:24])
				dst = net.IP(raw[24:40])
			}

			if proto == 6 && len(raw) >= ihl+20 {
				tcp := raw[ihl:]
				if len(tcp) < 20 {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				datOff := int((tcp[12]>>4)&0x0f) * 4
				if len(tcp) < datOff {
					_ = q.SetVerdict(id, nfqueue.NfAccept)
					return 0
				}
				payload := tcp[datOff:]
				sport := binary.BigEndian.Uint16(tcp[0:2])
				dport := binary.BigEndian.Uint16(tcp[2:4])
				if dport == 443 && len(payload) > 0 {
					k := fmt.Sprintf("%s:%d>%s:%d", src.String(), sport, dst.String(), dport)
					host, ok := w.feed(k, payload)
					if ok && w.matcher.Match(host) {
						log.Infof("TCP: %s %s:%d -> %s:%d", host, src.String(), sport, dst.String(), dport)
						go w.dropAndInjectTCP(raw, dst)
						_ = q.SetVerdict(id, nfqueue.NfDrop)
						return 0
					}
				}
			}

			if proto == 17 && len(raw) >= ihl+8 {
				udp := raw[ihl:]
				if len(udp) >= 8 {
					payload := udp[8:]
					sport := binary.BigEndian.Uint16(udp[0:2])
					dport := binary.BigEndian.Uint16(udp[2:4])
					if dport == 443 {
						if host, ok := sni.ParseQUICClientHelloSNI(payload); ok && w.matcher.Match(host) {
							log.Infof("UDP: %s %s:%d -> %s:%d", host, src.String(), sport, dst.String(), dport)
							go w.dropAndInjectQUIC(raw, dst)
							_ = q.SetVerdict(id, nfqueue.NfDrop)
							return 0
						}
					}
				}
			}

			_ = q.SetVerdict(id, nfqueue.NfAccept)
			return 0
		}, func(err error) int {
			log.Errorf("nfq: %v", err)
			return 0
		})
	}()

	return nil
}

func (w *Worker) dropAndInjectQUIC(raw []byte, dst net.IP) {
	// Send fake UDP packets first
	if w.cfg.FakeSNI { // Using same config for UDP faking
		for i := 0; i < 6; i++ { //  default is 6 for UDP
			fake, ok := sock.BuildFakeUDPFromOriginal(raw, 64, w.cfg.FakeTTL)
			if ok {
				_ = w.sock.SendIPv4(fake, dst)
				time.Sleep(1 * time.Millisecond)
			}
		}
	}

	splitPos := 24

	frags, ok := sock.IPv4FragmentUDP(raw, splitPos)
	if !ok {
		_ = w.sock.SendIPv4(raw, dst)
		return
	}

	// Send fragments with proper ordering and delay
	if w.cfg.FragSNIReverse {
		_ = w.sock.SendIPv4(frags[0], dst) // Second fragment first
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frags[1], dst) // First fragment second
	} else {
		_ = w.sock.SendIPv4(frags[1], dst) // First fragment
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frags[0], dst) // Second fragment
	}
}

func (w *Worker) dropAndInjectTCP(raw []byte, dst net.IP) {
	if len(raw) < 40 || raw[0]>>4 != 4 {
		_ = w.sock.SendIPv4(raw, dst)
		return
	}

	ipHdrLen := int((raw[0] & 0x0F) * 4)
	tcpHdrLen := int((raw[ipHdrLen+12] >> 4) * 4)
	payloadStart := ipHdrLen + tcpHdrLen
	payloadLen := len(raw) - payloadStart

	if payloadLen <= 0 {
		_ = w.sock.SendIPv4(raw, dst)
		return
	}

	// Send fake SNI packets first
	if w.cfg.FakeSNI {
		for i := 0; i < w.cfg.FakeSNISeqLength; i++ {
			fake := sock.BuildFakeSNIPacket(raw, w.cfg)
			if fake != nil {
				_ = w.sock.SendIPv4(fake, dst)
				// Small delay between fakes
				if i < w.cfg.FakeSNISeqLength-1 {
					time.Sleep(1 * time.Millisecond)
				}
			}
		}
	}

	// Find SNI position for smart splitting
	splitPos := w.cfg.FragSNIPosition
	if splitPos <= 0 {
		splitPos = 1
	}

	// If FragMiddleSNI is set, try to find and split in middle of SNI
	if w.cfg.FragMiddleSNI {
		// Simple heuristic: look for "www." or common TLD patterns
		payload := raw[payloadStart:]
		for i := 0; i < min(len(payload)-20, 100); i++ {
			if i+4 < len(payload) &&
				payload[i] == '.' &&
				((payload[i+1] == 'c' && payload[i+2] == 'o' && payload[i+3] == 'm') ||
					(payload[i+1] == 'o' && payload[i+2] == 'r' && payload[i+3] == 'g') ||
					(payload[i+1] == 'n' && payload[i+2] == 'e' && payload[i+3] == 't')) {
				splitPos = i + 2 // Split in middle of domain
				break
			}
		}
	}

	// Ensure split position is valid
	splitPos = min(splitPos, payloadLen-1)
	splitPos = max(splitPos, 1)

	// Fragment based on strategy
	switch w.cfg.FragmentStrategy {
	case "tcp":
		w.sendTCPFragments(raw, payloadStart+splitPos, dst)
	case "ip":
		w.sendIPFragments(raw, payloadStart+splitPos, dst)
	case "none":
		_ = w.sock.SendIPv4(raw, dst)
	default:
		w.sendTCPFragments(raw, payloadStart+splitPos, dst)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (w *Worker) feed(key string, chunk []byte) (string, bool) {
	w.mu.Lock()
	st := w.flows[key]
	if st == nil {
		st = &flowState{buf: nil, last: time.Now()}
		w.flows[key] = st
	}
	if len(st.buf) < w.limit {
		need := w.limit - len(st.buf)
		if len(chunk) < need {
			st.buf = append(st.buf, chunk...)
		} else {
			st.buf = append(st.buf, chunk[:need]...)
		}
	}
	st.last = time.Now()
	buf := append([]byte(nil), st.buf...)
	w.mu.Unlock()
	host, ok := sni.ParseTLSClientHelloSNI(buf)
	if ok && host != "" {
		w.mu.Lock()
		delete(w.flows, key)
		w.mu.Unlock()
		return host, true
	}
	return "", false
}

func (w *Worker) sendTCPFragments(packet []byte, splitPos int, dst net.IP) {
	if splitPos <= 0 || splitPos >= len(packet) {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	ipHdrLen := int((packet[0] & 0x0F) * 4)
	tcpHdrLen := int((packet[ipHdrLen+12] >> 4) * 4)
	hdrTotal := ipHdrLen + tcpHdrLen

	if splitPos <= hdrTotal {
		splitPos = hdrTotal + 1
	}

	// Create two segments
	seg1Len := splitPos
	seg1 := make([]byte, seg1Len)
	copy(seg1, packet[:seg1Len])

	seg2Len := len(packet) - splitPos + hdrTotal
	seg2 := make([]byte, seg2Len)

	// Copy headers to segment 2
	copy(seg2, packet[:hdrTotal])
	// Copy remaining payload
	copy(seg2[hdrTotal:], packet[splitPos:])

	// Fix segment 1
	binary.BigEndian.PutUint16(seg1[2:4], uint16(seg1Len))
	sock.FixIPv4Checksum(seg1[:ipHdrLen])
	sock.FixTCPChecksum(seg1)

	// Fix segment 2
	binary.BigEndian.PutUint16(seg2[2:4], uint16(seg2Len))
	// Adjust sequence number
	seq := binary.BigEndian.Uint32(seg2[ipHdrLen+4 : ipHdrLen+8])
	binary.BigEndian.PutUint32(seg2[ipHdrLen+4:ipHdrLen+8],
		seq+uint32(splitPos-hdrTotal))
	sock.FixIPv4Checksum(seg2[:ipHdrLen])
	sock.FixTCPChecksum(seg2)
	log.Tracef("Segmented TCP packet into two fragments at pos=%d: seg1Len=%d, seg2Len=%d", splitPos, seg1Len, seg2Len)
	// Send in order based on config
	if w.cfg.FragSNIReverse {
		_ = w.sock.SendIPv4(seg2, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(seg1, dst)
	} else {
		_ = w.sock.SendIPv4(seg1, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(seg2, dst)
	}
}

func (w *Worker) sendIPFragments(packet []byte, splitPos int, dst net.IP) {
	if splitPos <= 0 || splitPos >= len(packet) {
		_ = w.sock.SendIPv4(packet, dst)
		return
	}

	ipHdrLen := int((packet[0] & 0x0F) * 4)

	// Align to 8-byte boundary for IP fragmentation
	splitPos = (splitPos + 7) &^ 7
	if splitPos >= len(packet) {
		splitPos = len(packet) - 8
	}

	// Fragment 1
	frag1 := make([]byte, splitPos)
	copy(frag1, packet[:splitPos])

	// Set MF flag
	frag1[6] |= 0x20
	binary.BigEndian.PutUint16(frag1[2:4], uint16(splitPos))
	sock.FixIPv4Checksum(frag1[:ipHdrLen])

	// Fragment 2
	frag2Len := ipHdrLen + len(packet) - splitPos
	frag2 := make([]byte, frag2Len)
	copy(frag2, packet[:ipHdrLen])
	copy(frag2[ipHdrLen:], packet[splitPos:])

	// Set fragment offset
	fragOff := uint16(splitPos-ipHdrLen) / 8
	binary.BigEndian.PutUint16(frag2[6:8], fragOff)
	binary.BigEndian.PutUint16(frag2[2:4], uint16(frag2Len))
	sock.FixIPv4Checksum(frag2[:ipHdrLen])
	log.Tracef("Fragmented IP packet into two fragments at pos=%d: frag1Len=%d, frag2Len=%d", splitPos, len(frag1), len(frag2))
	// Send fragments
	if w.cfg.FragSNIReverse {
		_ = w.sock.SendIPv4(frag2, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frag1, dst)
	} else {
		_ = w.sock.SendIPv4(frag1, dst)
		time.Sleep(time.Duration(w.cfg.Seg2Delay) * time.Millisecond)
		_ = w.sock.SendIPv4(frag2, dst)
	}
}

func (w *Worker) Stop() {
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
	if w.q != nil {
		_ = w.q.Close()
	}
	if w.sock != nil {
		w.sock.Close()
	}
}

func (w *Worker) gc() {
	defer w.wg.Done()
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case now := <-t.C:
			w.mu.Lock()
			for k, st := range w.flows {
				if now.Sub(st.last) > w.ttl {
					delete(w.flows, k)
				}
			}
			w.mu.Unlock()
		}
	}
}
