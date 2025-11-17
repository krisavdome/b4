package checker

import (
	"fmt"

	"github.com/daniellavrushin/b4/config"
)

type ConfigPreset struct {
	Name        string
	Description string
	Config      config.SetConfig
}

// GetTestPresets generates focused preset combinations
func GetTestPresets() []ConfigPreset {
	presets := []ConfigPreset{}

	// 1. SYN Fake Strategies (only meaningful combos)
	synLengths := []int{0, 64, 256}
	synTTLs := []uint8{3, 5, 8}

	for _, synLen := range synLengths {
		for _, ttl := range synTTLs {
			presets = append(presets, ConfigPreset{
				Name:        fmt.Sprintf("syn-%d-ttl%d", synLen, ttl),
				Description: fmt.Sprintf("SYN fake len=%d, TTL=%d", synLen, ttl),
				Config: config.SetConfig{
					TCP:           config.TCPConfig{ConnBytesLimit: 19, SynFake: true, SynFakeLen: synLen},
					UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
					Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: 1},
					Faking:        config.FakingConfig{SNI: true, TTL: ttl, Strategy: "ttl", SeqOffset: 10000, SNISeqLength: 1, SNIType: 2},
				},
			})
		}
	}

	// 2. TCP Fragmentation Strategies (reduced set)
	fragConfigs := []struct {
		pos     int
		reverse bool
		middle  bool
	}{
		{1, false, false}, // standard
		{1, true, false},  // reverse
		{1, false, true},  // middle SNI
		{3, false, false}, // different position
		{5, false, false}, // another position
	}

	for _, fc := range fragConfigs {
		name := fmt.Sprintf("tcp-pos%d", fc.pos)
		if fc.reverse {
			name += "-rev"
		}
		if fc.middle {
			name += "-mid"
		}

		presets = append(presets, ConfigPreset{
			Name:        name,
			Description: fmt.Sprintf("TCP frag pos=%d reverse=%v middle=%v", fc.pos, fc.reverse, fc.middle),
			Config: config.SetConfig{
				TCP:           config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: fc.pos, SNIReverse: fc.reverse, MiddleSNI: fc.middle},
				Faking:        config.FakingConfig{SNI: true, TTL: 8, Strategy: "pastseq", SeqOffset: 10000, SNISeqLength: 1, SNIType: 2},
			},
		})
	}

	// 3. Faking Strategies with proper variations
	fakeConfigs := []struct {
		strategy string
		ttl      uint8
		seqLen   int
		offset   int32
	}{
		{"ttl", 3, 1, 0},
		{"ttl", 5, 2, 0},
		{"ttl", 8, 3, 0},
		{"pastseq", 8, 1, 10000},
		{"pastseq", 8, 2, 50000},
		{"randseq", 8, 1, 100000},
		{"md5sum", 8, 1, 0},
		{"tcp_check", 8, 2, 0}, // ADD tcp_check strategy
	}

	for _, fc := range fakeConfigs {
		presets = append(presets, ConfigPreset{
			Name:        fmt.Sprintf("fake-%s-ttl%d-len%d", fc.strategy, fc.ttl, fc.seqLen),
			Description: fmt.Sprintf("Fake %s TTL=%d seqLen=%d", fc.strategy, fc.ttl, fc.seqLen),
			Config: config.SetConfig{
				TCP:           config.TCPConfig{ConnBytesLimit: 19},
				UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: 1},
				Faking:        config.FakingConfig{SNI: true, TTL: fc.ttl, Strategy: fc.strategy, SeqOffset: fc.offset, SNISeqLength: fc.seqLen, SNIType: 2},
			},
		})
	}

	// 4. UDP/QUIC strategies (fixed)
	udpConfigs := []struct {
		mode       string
		fakeLen    int
		fakeSeq    int
		strategy   string
		quicFilter string
	}{
		{"fake", 64, 6, "ttl", "disabled"},
		{"fake", 128, 10, "checksum", "parse"},
		{"fake", 256, 12, "none", "all"},
		{"drop", 0, 0, "none", "all"},
	}

	for _, uc := range udpConfigs {
		presets = append(presets, ConfigPreset{
			Name:        fmt.Sprintf("udp-%s-q%s", uc.mode, uc.quicFilter),
			Description: fmt.Sprintf("UDP %s QUIC=%s", uc.mode, uc.quicFilter),
			Config: config.SetConfig{
				TCP:           config.TCPConfig{ConnBytesLimit: 19},
				UDP:           config.UDPConfig{Mode: uc.mode, FakeSeqLength: uc.fakeSeq, FakeLen: uc.fakeLen, FakingStrategy: uc.strategy, FilterQUIC: uc.quicFilter, FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: 1},
				Faking:        config.FakingConfig{SNI: true, TTL: 8, Strategy: "pastseq", SeqOffset: 10000, SNISeqLength: 1, SNIType: 2},
			},
		})
	}

	// 5. IP Fragmentation
	presets = append(presets, ConfigPreset{
		Name:        "ip-frag",
		Description: "IP fragmentation",
		Config: config.SetConfig{
			TCP:           config.TCPConfig{ConnBytesLimit: 19},
			UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
			Fragmentation: config.FragmentationConfig{Strategy: "ip", SNIPosition: 1},
			Faking:        config.FakingConfig{SNI: true, TTL: 8, Strategy: "pastseq", SeqOffset: 10000, SNISeqLength: 1, SNIType: 2},
		},
	})

	// 6. Delay variations (only test meaningful delays)
	delays := []int{5, 10, 20}
	for _, delay := range delays {
		presets = append(presets, ConfigPreset{
			Name:        fmt.Sprintf("delay-%d", delay),
			Description: fmt.Sprintf("Segment delay %dms", delay),
			Config: config.SetConfig{
				TCP:           config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: delay},
				UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8, Seg2Delay: delay},
				Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: 1, SNIReverse: true},
				Faking:        config.FakingConfig{SNI: true, TTL: 5, Strategy: "randseq", SeqOffset: 50000, SNISeqLength: 3, SNIType: 2},
			},
		})
	}

	// 7. Aggressive combo (truly aggressive)
	presets = append(presets, ConfigPreset{
		Name:        "aggressive",
		Description: "Max bypass: SYN fake + multi-fake + reverse frag",
		Config: config.SetConfig{
			TCP:           config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 10, SynFake: true, SynFakeLen: 256},
			UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 15, FakeLen: 256, FakingStrategy: "checksum", FilterQUIC: "all", FilterSTUN: true, ConnBytesLimit: 8},
			Fragmentation: config.FragmentationConfig{Strategy: "tcp", SNIPosition: 1, SNIReverse: true, MiddleSNI: true},
			Faking:        config.FakingConfig{SNI: true, TTL: 3, Strategy: "pastseq", SeqOffset: 100000, SNISeqLength: 5, SNIType: 2},
		},
	})

	// 8. Baseline (no bypass)
	presets = append(presets, ConfigPreset{
		Name:        "baseline",
		Description: "No bypass techniques",
		Config: config.SetConfig{
			TCP:           config.TCPConfig{ConnBytesLimit: 19},
			UDP:           config.UDPConfig{Mode: "fake", FakeSeqLength: 0, FakeLen: 0, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: false, ConnBytesLimit: 8},
			Fragmentation: config.FragmentationConfig{Strategy: "none"},
			Faking:        config.FakingConfig{SNI: false, TTL: 8, Strategy: "none", SeqOffset: 0, SNISeqLength: 0, SNIType: 2},
		},
	})

	return presets
}
