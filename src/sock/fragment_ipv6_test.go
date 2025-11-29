package sock

import "testing"

func TestIPv6SendTCPSegments_TooShort(t *testing.T) {
	_, ok := IPv6SendTCPSegments(make([]byte, 30), 10)
	if ok {
		t.Error("expected false for short packet")
	}
}

func TestIPv6SendTCPSegments_NotIPv6(t *testing.T) {
	pkt := buildMinimalIPv4TCPPacket(100)
	_, ok := IPv6SendTCPSegments(pkt, 10)
	if ok {
		t.Error("expected false for IPv4 packet")
	}
}

func TestIPv6SendTCPSegments_Valid(t *testing.T) {
	pkt := buildMinimalIPv6TCPPacket(100)
	segs, ok := IPv6SendTCPSegments(pkt, 20)
	if !ok {
		t.Fatal("expected success")
	}
	if len(segs) != 2 {
		t.Errorf("expected 2 segments, got %d", len(segs))
	}
}

func TestIPv6SendTCPSegments_InvalidSplit(t *testing.T) {
	pkt := buildMinimalIPv6TCPPacket(100)

	// Split at 0
	_, ok := IPv6SendTCPSegments(pkt, 0)
	if ok {
		t.Error("expected false for split=0")
	}

	// Split >= payload
	_, ok = IPv6SendTCPSegments(pkt, 200)
	if ok {
		t.Error("expected false for split >= payload")
	}
}

func TestIPv6FragmentPacket_TooShort(t *testing.T) {
	_, ok := IPv6FragmentPacket(make([]byte, 30), 10)
	if ok {
		t.Error("expected false")
	}
}

func TestIPv6FragmentPacket_NotIPv6(t *testing.T) {
	pkt := buildMinimalIPv4TCPPacket(100)
	_, ok := IPv6FragmentPacket(pkt, 10)
	if ok {
		t.Error("expected false for IPv4")
	}
}

func TestIPv6FragmentPacket_Valid(t *testing.T) {
	pkt := buildMinimalIPv6TCPPacket(100)
	frags, ok := IPv6FragmentPacket(pkt, 16)
	if !ok {
		t.Fatal("expected success")
	}
	if len(frags) != 2 {
		t.Errorf("expected 2 fragments, got %d", len(frags))
	}

	// First fragment should have next header = 44 (Fragment)
	if frags[0][6] != 44 {
		t.Errorf("expected fragment header, got %d", frags[0][6])
	}
}
