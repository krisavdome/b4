// path: src/config/config.go
package config

import (
	"fmt"

	"github.com/daniellavrushin/b4/geodat"
	"github.com/daniellavrushin/b4/log"
	"github.com/spf13/cobra"
)

type Config struct {
	QueueStartNum  int
	Mark           uint
	ConnBytesLimit int
	Logging        Logging
	SNIDomains     []string
	Threads        int
	UseGSO         bool
	UseConntrack   bool
	SkipIpTables   bool
	GeoSitePath    string
	GeoIpPath      string
	GeoCategories  []string
	Seg2Delay      int

	FragmentStrategy string
	FragSNIReverse   bool
	FragMiddleSNI    bool
	FragSNIPosition  int

	FakeSNI           bool
	FakeTTL           uint8
	FakeStrategy      string
	FakeSeqOffset     int32
	FakeSNISeqLength  int
	FakeSNIType       int
	FakeCustomPayload string

	UDPMode           string
	UDPFakeSeqLength  int
	UDPFakeLen        int
	UDPFakingStrategy string
	UDPDPortMin       int
	UDPDPortMax       int
	UDPFilterQUIC     string
}

type Logging struct {
	Level      log.Level
	Instaflush bool
	Syslog     bool
}

const (
	FakePayloadRandom = iota
	FakePayloadCustom
	FakePayloadDefault
)

var DefaultConfig = Config{
	QueueStartNum:  537,
	Mark:           1 << 15,
	Threads:        4,
	ConnBytesLimit: 19,
	UseConntrack:   false,
	UseGSO:         false,
	SkipIpTables:   false,
	GeoSitePath:    "",
	GeoIpPath:      "",
	GeoCategories:  []string{},
	Seg2Delay:      0,

	FragmentStrategy: "tcp",
	FragSNIReverse:   true,
	FragMiddleSNI:    true,
	FragSNIPosition:  1,

	FakeSNI:           true,
	FakeTTL:           8,
	FakeSNISeqLength:  1,
	FakeSNIType:       FakePayloadDefault,
	FakeCustomPayload: "",
	FakeStrategy:      "pastseq",
	FakeSeqOffset:     10000,

	UDPMode:           "drop",
	UDPFakeSeqLength:  6,
	UDPFakeLen:        64,
	UDPFakingStrategy: "none",
	UDPDPortMin:       0,
	UDPDPortMax:       0,
	UDPFilterQUIC:     "parse",

	Logging: Logging{
		Level:      log.LevelInfo,
		Instaflush: true,
		Syslog:     false,
	},
}

func (c *Config) BindFlags(cmd *cobra.Command) {
	// Network configuration
	cmd.Flags().IntVar(&c.QueueStartNum, "queue-num", c.QueueStartNum, "Netfilter queue number")
	cmd.Flags().IntVar(&c.Threads, "threads", c.Threads, "Number of worker threads")
	cmd.Flags().UintVar(&c.Mark, "mark", c.Mark, "Packet mark value")
	cmd.Flags().IntVar(&c.ConnBytesLimit, "connbytes-limit", c.ConnBytesLimit, "Connection bytes limit")
	cmd.Flags().StringSliceVar(&c.SNIDomains, "sni-domains", c.SNIDomains, "List of SNI domains to match")
	cmd.Flags().IntVar(&c.Seg2Delay, "seg2delay", 0, "Delay between segments in ms")

	// Geodata and site filtering
	cmd.Flags().StringVar(&c.GeoSitePath, "geosite", c.GeoSitePath, "Path to geosite file (e.g., geosite.dat)")
	cmd.Flags().StringVar(&c.GeoIpPath, "geoip", c.GeoIpPath, "Path to geoip file (e.g., geoip.dat)")
	cmd.Flags().StringSliceVar(&c.GeoCategories, "geo-categories", c.GeoCategories, "Geographic categories to process (e.g., youtube,facebook,amazon)")

	// Fake SNI and TTL configuration
	cmd.Flags().StringVar(&c.FragmentStrategy, "frag", "tcp", "Fragmentation strategy (tcp/ip/none)")
	cmd.Flags().BoolVar(&c.FragSNIReverse, "frag-sni-reverse", true, "Reverse fragment order")
	cmd.Flags().BoolVar(&c.FragMiddleSNI, "frag-middle-sni", true, "Fragment in middle of SNI")
	cmd.Flags().IntVar(&c.FragSNIPosition, "frag-sni-pos", 1, "SNI fragment position")

	cmd.Flags().StringVar(&c.FakeStrategy, "fake-strategy", "ttl", "Faking strategy (ttl/randseq/pastseq/tcp_check/md5sum)")
	cmd.Flags().Uint8Var(&c.FakeTTL, "fake-ttl", 8, "TTL for fake packets")
	cmd.Flags().Int32Var(&c.FakeSeqOffset, "fake-seq-offset", 10000, "Sequence offset for fake packets")
	cmd.Flags().BoolVar(&c.FakeSNI, "fake-sni", c.FakeSNI, "Enable fake SNI packets")
	cmd.Flags().IntVar(&c.FakeSNISeqLength, "fake-sni-len", c.FakeSNISeqLength, "Length of fake SNI sequence")
	cmd.Flags().IntVar(&c.FakeSNIType, "fake-sni-type", c.FakeSNIType, "Type of fake SNI payload (0=random, 1=custom, 2=default)")

	cmd.Flags().StringVar(&c.UDPMode, "udp-mode", c.UDPMode, "UDP handling strategy (drop|fake)")
	cmd.Flags().IntVar(&c.UDPFakeSeqLength, "udp-fake-seq-len", c.UDPFakeSeqLength, "UDP fake packet sequence length")
	cmd.Flags().IntVar(&c.UDPFakeLen, "udp-fake-len", c.UDPFakeLen, "UDP fake packet size in bytes")
	cmd.Flags().StringVar(&c.UDPFakingStrategy, "udp-faking-strategy", c.UDPFakingStrategy, "UDP faking strategy (none|ttl|checksum)")
	cmd.Flags().IntVar(&c.UDPDPortMin, "udp-dport-min", c.UDPDPortMin, "Minimum UDP destination port to handle")
	cmd.Flags().IntVar(&c.UDPDPortMax, "udp-dport-max", c.UDPDPortMax, "Maximum UDP destination port to handle")
	cmd.Flags().StringVar(&c.UDPFilterQUIC, "udp-filter-quic", c.UDPFilterQUIC, "QUIC filtering mode (disabled|all|parse)")

	// Feature flags
	cmd.Flags().BoolVar(&c.UseGSO, "gso", c.UseGSO, "Enable Generic Segmentation Offload")
	cmd.Flags().BoolVar(&c.UseConntrack, "conntrack", c.UseConntrack, "Enable connection tracking")
	cmd.Flags().BoolVar(&c.SkipIpTables, "skip-iptables", c.SkipIpTables, "Skip iptables rules setup")

	// Logging configuration
	cmd.Flags().BoolVarP(&c.Logging.Instaflush, "instaflush", "i", c.Logging.Instaflush, "Flush logs immediately")
	cmd.Flags().BoolVar(&c.Logging.Syslog, "syslog", c.Logging.Syslog, "Enable syslog output")
}

func (cfg *Config) ApplyLogLevel(level string) {
	switch level {
	case "debug":
		cfg.Logging.Level = log.LevelDebug
	case "trace":
		cfg.Logging.Level = log.LevelTrace
	case "info":
		cfg.Logging.Level = log.LevelInfo
	case "error":
		cfg.Logging.Level = log.LevelError
	case "silent":
		cfg.Logging.Level = -1
	default:
		cfg.Logging.Level = log.LevelInfo
	}
}

func (c *Config) Validate() error {
	// If sites are specified, geodata path must be provided
	if len(c.GeoCategories) > 0 && c.GeoSitePath == "" {
		return fmt.Errorf("--geosite must be specified when using --geo-categories")
	}

	if c.Threads < 1 {
		return fmt.Errorf("threads must be at least 1")
	}

	if c.QueueStartNum < 0 || c.QueueStartNum > 65535 {
		return fmt.Errorf("queue-num must be between 0 and 65535")
	}

	return nil
}

func (c *Config) LogString() string {
	return ""
}

// LoadDomainsFromGeodata loads domains from geodata file for specified sites
// and returns them as a slice
func (c *Config) LoadDomainsFromGeodata() ([]string, error) {
	return geodat.LoadDomainsFromSites(c.GeoSitePath, c.GeoCategories)
}
