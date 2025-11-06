// path: src/config/config.go
package config

import (
	_ "embed"

	"github.com/daniellavrushin/b4/log"
)

type Config struct {
	ConfigPath string `json:"-" bson:"-"`

	// Netfilter Queue Configuration
	Queue QueueConfig `json:"queue" bson:"queue"`

	// DPI Bypass Configuration
	Bypass BypassConfig `json:"bypass" bson:"bypass"`

	// Domain Filtering Configuration
	Domains DomainsConfig `json:"domains" bson:"domains"`

	// System Configuration
	System SystemConfig `json:"system" bson:"system"`
}

var DefaultConfig = Config{
	ConfigPath: "",

	Queue: QueueConfig{
		StartNum:    537,
		Mark:        1 << 15,
		Threads:     4,
		IPv4Enabled: true,
		IPv6Enabled: false,
	},

	Bypass: BypassConfig{
		TCP: TCPConfig{
			ConnBytesLimit: 19,
			Seg2Delay:      0,
		},

		UDP: UDPConfig{
			Mode:           "fake",
			FakeSeqLength:  6,
			FakeLen:        64,
			FakingStrategy: "none",
			DPortMin:       0,
			DPortMax:       0,
			FilterQUIC:     "disabled",
			FilterSTUN:     true,
			ConnBytesLimit: 8,
		},

		Fragmentation: Fragmentation{
			Strategy:    "tcp",
			SNIReverse:  true,
			MiddleSNI:   true,
			SNIPosition: 1,
		},

		Faking: Faking{
			SNI:           true,
			TTL:           8,
			SNISeqLength:  1,
			SNIType:       FakePayloadDefault,
			CustomPayload: "",
			Strategy:      "pastseq",
			SeqOffset:     10000,
		},
	},

	Domains: DomainsConfig{
		GeoSitePath:       "",
		GeoIpPath:         "",
		SNIDomains:        []string{},
		GeoSiteCategories: []string{},
		GeoIpCategories:   []string{},
	},

	System: SystemConfig{
		Tables: TablesConfig{
			MonitorInterval: 10,
			SkipSetup:       false,
		},

		WebServer: WebServer{
			Port:      7000,
			IsEnabled: true,
		},

		Logging: Logging{
			Level:      log.LevelInfo,
			Instaflush: true,
			Syslog:     false,
		},

		Checker: CheckerConfig{
			TimeoutSeconds: 15,
			MaxConcurrent:  4,
			Domains:        []string{},
		},
	},
}
