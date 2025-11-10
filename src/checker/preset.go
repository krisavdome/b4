package checker

import "github.com/daniellavrushin/b4/config"

type ConfigPreset struct {
	Name        string
	Description string
	Config      config.SetConfig
}

func GetTestPresets() []ConfigPreset {
	return []ConfigPreset{
		// Basic TCP fragmentation strategies
		{
			Name:        "tcp-frag-pos1",
			Description: "TCP fragmentation at position 1 (most common)",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "tcp-frag-pos2",
			Description: "TCP fragmentation at position 2",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 2,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "tcp-frag-reverse",
			Description: "TCP fragmentation with reversed segment order",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  true,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "tcp-middle-sni",
			Description: "TCP fragmentation in middle of SNI",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   true,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},

		// IP fragmentation strategies
		{
			Name:        "ip-frag-pos1",
			Description: "IP-level fragmentation at position 1",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "ip",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "ip-frag-reverse",
			Description: "IP fragmentation with reversed order",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "ip",
					SNIPosition: 1,
					SNIReverse:  true,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},

		// Different faking strategies
		{
			Name:        "fake-ttl-low",
			Description: "Fake packets with low TTL (3)",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          3,
					Strategy:     "ttl",
					SeqOffset:    10000,
					SNISeqLength: 2,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "fake-randseq",
			Description: "Fake packets with random sequence numbers",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "randseq",
					SeqOffset:    10000,
					SNISeqLength: 3,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "fake-md5sum",
			Description: "Fake packets with MD5 checksum strategy",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "md5sum",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},

		// UDP/QUIC strategies
		{
			Name:        "quic-drop",
			Description: "Drop QUIC packets (force TCP fallback)",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "drop", FakeSeqLength: 0, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "all", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "quic-fake-frag",
			Description: "QUIC with fake packets and fragmentation",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 10, FakeLen: 128, FakingStrategy: "ttl", FilterQUIC: "all", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 1,
					SNIType:      2,
				},
			},
		},

		// Aggressive strategies
		{
			Name:        "aggressive-multi-fake",
			Description: "Multiple fake packets with high TTL",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 5},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 12, FakeLen: 64, FakingStrategy: "ttl", FilterQUIC: "all", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  true,
					MiddleSNI:   true,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          5,
					Strategy:     "randseq",
					SeqOffset:    50000,
					SNISeqLength: 5,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "aggressive-delay",
			Description: "Fragmentation with segment delay",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 10},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 2,
					SNIReverse:  true,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 2,
					SNIType:      2,
				},
			},
		},

		// Minimal/stealth strategies
		{
			Name:        "minimal-no-fake",
			Description: "TCP fragmentation without fake packets",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 0, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "tcp",
					SNIPosition: 1,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          false,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 0,
					SNIType:      2,
				},
			},
		},
		{
			Name:        "no-fragmentation",
			Description: "Only fake packets, no fragmentation",
			Config: config.SetConfig{
				TCP: config.TCPConfig{ConnBytesLimit: 19, Seg2Delay: 0},
				UDP: config.UDPConfig{Mode: "fake", FakeSeqLength: 6, FakeLen: 64, FakingStrategy: "none", FilterQUIC: "disabled", FilterSTUN: true, ConnBytesLimit: 8},
				Fragmentation: config.FragmentationConfig{
					Strategy:    "none",
					SNIPosition: 0,
					SNIReverse:  false,
					MiddleSNI:   false,
				},
				Faking: config.FakingConfig{
					SNI:          true,
					TTL:          8,
					Strategy:     "pastseq",
					SeqOffset:    10000,
					SNISeqLength: 3,
					SNIType:      2,
				},
			},
		},
	}
}
