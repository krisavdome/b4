package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/daniellavrushin/b4/geodat"
	"github.com/daniellavrushin/b4/log"
	"github.com/spf13/cobra"
)

func (c *Config) SaveToFile(path string) error {
	if path == "" {
		log.Tracef("config path is not defined")
		return nil
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return log.Errorf("failed to marshal config: %v", err)
	}

	file, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return log.Errorf("failed to create config file: %v", err)
	}
	defer file.Close()

	_, err = file.Write(data)
	if err != nil {
		return log.Errorf("failed to write config file: %v", err)
	}
	return nil
}

func (c *Config) LoadFromFile(path string) error {
	if path == "" {
		log.Tracef("config path is not defined")
		return nil
	}

	info, err := os.Stat(path)
	if err != nil {
		return log.Errorf("failed to stat config file: %v", err)
	}
	if info.IsDir() {
		return log.Errorf("config path is a directory, not a file: %s", path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return log.Errorf("failed to read config file: %v", err)
	}
	err = json.Unmarshal(data, c)
	if err != nil {
		return log.Errorf("failed to parse config file: %v", err)
	}
	return nil
}

func (c *Config) BindFlags(cmd *cobra.Command) {
	// Config path
	cmd.Flags().StringVar(&c.ConfigPath, "config", c.ConfigPath, "Path to config file")

	// Queue configuration
	cmd.Flags().IntVar(&c.Queue.StartNum, "queue-num", c.Queue.StartNum, "Netfilter queue number")
	cmd.Flags().IntVar(&c.Queue.Threads, "threads", c.Queue.Threads, "Number of worker threads")
	cmd.Flags().UintVar(&c.Queue.Mark, "mark", c.Queue.Mark, "Packet mark value (default 32768)")
	cmd.Flags().BoolVar(&c.Queue.IPv4Enabled, "ipv4", c.Queue.IPv4Enabled, "Enable IPv4 processing")
	cmd.Flags().BoolVar(&c.Queue.IPv6Enabled, "ipv6", c.Queue.IPv6Enabled, "Enable IPv6 processing")

	// TCP bypass configuration
	cmd.Flags().IntVar(&c.Bypass.TCP.ConnBytesLimit, "connbytes-limit", c.Bypass.TCP.ConnBytesLimit, "TCP connection bytes limit (default 19)")
	cmd.Flags().IntVar(&c.Bypass.TCP.Seg2Delay, "seg2delay", c.Bypass.TCP.Seg2Delay, "Delay between segments in ms")

	// UDP bypass configuration
	cmd.Flags().StringVar(&c.Bypass.UDP.Mode, "udp-mode", c.Bypass.UDP.Mode, "UDP handling strategy (drop|fake)")
	cmd.Flags().IntVar(&c.Bypass.UDP.FakeSeqLength, "udp-fake-seq-len", c.Bypass.UDP.FakeSeqLength, "UDP fake packet sequence length")
	cmd.Flags().IntVar(&c.Bypass.UDP.FakeLen, "udp-fake-len", c.Bypass.UDP.FakeLen, "UDP fake packet size in bytes")
	cmd.Flags().StringVar(&c.Bypass.UDP.FakingStrategy, "udp-faking-strategy", c.Bypass.UDP.FakingStrategy, "UDP faking strategy (none|ttl|checksum)")
	cmd.Flags().IntVar(&c.Bypass.UDP.DPortMin, "udp-dport-min", c.Bypass.UDP.DPortMin, "Minimum UDP destination port to handle")
	cmd.Flags().IntVar(&c.Bypass.UDP.DPortMax, "udp-dport-max", c.Bypass.UDP.DPortMax, "Maximum UDP destination port to handle")
	cmd.Flags().StringVar(&c.Bypass.UDP.FilterQUIC, "udp-filter-quic", c.Bypass.UDP.FilterQUIC, "QUIC filtering mode (disabled|all|parse)")
	cmd.Flags().BoolVar(&c.Bypass.UDP.FilterSTUN, "udp-filter-stun", c.Bypass.UDP.FilterSTUN, "STUN filtering mode (disabled|all|parse)")
	cmd.Flags().IntVar(&c.Bypass.UDP.ConnBytesLimit, "udp-conn-bytes-limit", c.Bypass.UDP.ConnBytesLimit, "UDP connection bytes limit (default 8)")

	// Fragmentation configuration
	cmd.Flags().StringVar(&c.Bypass.Fragmentation.Strategy, "frag", c.Bypass.Fragmentation.Strategy, "Fragmentation strategy (tcp|ip|none)")
	cmd.Flags().BoolVar(&c.Bypass.Fragmentation.SNIReverse, "frag-sni-reverse", c.Bypass.Fragmentation.SNIReverse, "Reverse fragment order")
	cmd.Flags().BoolVar(&c.Bypass.Fragmentation.MiddleSNI, "frag-middle-sni", c.Bypass.Fragmentation.MiddleSNI, "Fragment in middle of SNI")
	cmd.Flags().IntVar(&c.Bypass.Fragmentation.SNIPosition, "frag-sni-pos", c.Bypass.Fragmentation.SNIPosition, "SNI fragment position")

	// Faking configuration
	cmd.Flags().StringVar(&c.Bypass.Faking.Strategy, "fake-strategy", c.Bypass.Faking.Strategy, "Faking strategy (ttl|randseq|pastseq|tcp_check|md5sum)")
	cmd.Flags().Uint8Var(&c.Bypass.Faking.TTL, "fake-ttl", c.Bypass.Faking.TTL, "TTL for fake packets")
	cmd.Flags().Int32Var(&c.Bypass.Faking.SeqOffset, "fake-seq-offset", c.Bypass.Faking.SeqOffset, "Sequence offset for fake packets")
	cmd.Flags().BoolVar(&c.Bypass.Faking.SNI, "fake-sni", c.Bypass.Faking.SNI, "Enable fake SNI packets")
	cmd.Flags().IntVar(&c.Bypass.Faking.SNISeqLength, "fake-sni-len", c.Bypass.Faking.SNISeqLength, "Length of fake SNI sequence")
	cmd.Flags().IntVar(&c.Bypass.Faking.SNIType, "fake-sni-type", c.Bypass.Faking.SNIType, "Type of fake SNI payload (0=random, 1=custom, 2=default)")

	// Domain filtering
	cmd.Flags().StringSliceVar(&c.Domains.SNIDomains, "sni-domains", c.Domains.SNIDomains, "List of SNI domains to match")
	cmd.Flags().StringVar(&c.Domains.GeoSitePath, "geosite", c.Domains.GeoSitePath, "Path to geosite file (e.g., geosite.dat)")
	cmd.Flags().StringVar(&c.Domains.GeoIpPath, "geoip", c.Domains.GeoIpPath, "Path to geoip file (e.g., geoip.dat)")
	cmd.Flags().StringSliceVar(&c.Domains.GeoSiteCategories, "geosite-categories", c.Domains.GeoSiteCategories, "Geographic categories to process (e.g., youtube,facebook,amazon)")
	cmd.Flags().StringSliceVar(&c.Domains.GeoIpCategories, "geoip-categories", c.Domains.GeoIpCategories, "Geographic categories to process (e.g., youtube,facebook,amazon)")

	// System configuration
	cmd.Flags().IntVar(&c.System.Tables.MonitorInterval, "tables-monitor-interval", c.System.Tables.MonitorInterval, "Tables monitor interval in seconds (default 10, 0 to disable)")
	cmd.Flags().BoolVar(&c.System.Tables.SkipSetup, "skip-tables", c.System.Tables.SkipSetup, "Skip iptables/nftables setup on startup")

	cmd.Flags().BoolVarP(&c.System.Logging.Instaflush, "instaflush", "i", c.System.Logging.Instaflush, "Flush logs immediately")
	cmd.Flags().BoolVar(&c.System.Logging.Syslog, "syslog", c.System.Logging.Syslog, "Enable syslog output")

	cmd.Flags().IntVar(&c.System.WebServer.Port, "web-port", c.System.WebServer.Port, "Port for internal web server (0 disables)")
}

func (cfg *Config) ApplyLogLevel(level string) {
	switch level {
	case "debug":
		cfg.System.Logging.Level = log.LevelDebug
	case "trace":
		cfg.System.Logging.Level = log.LevelTrace
	case "info":
		cfg.System.Logging.Level = log.LevelInfo
	case "error":
		cfg.System.Logging.Level = log.LevelError
	case "silent":
		cfg.System.Logging.Level = -1
	default:
		cfg.System.Logging.Level = log.LevelInfo
	}
}

func (c *Config) Validate() error {
	c.System.WebServer.IsEnabled = c.System.WebServer.Port > 0 && c.System.WebServer.Port <= 65535

	if len(c.Domains.GeoSiteCategories) > 0 && c.Domains.GeoSitePath == "" {
		return fmt.Errorf("--geosite must be specified when using --geo-categories")
	}

	if c.Queue.Threads < 1 {
		return fmt.Errorf("threads must be at least 1")
	}

	if c.Queue.StartNum < 0 || c.Queue.StartNum > 65535 {
		return fmt.Errorf("queue-num must be between 0 and 65535")
	}

	return nil
}

func (c *Config) LogString() string {
	return ""
}

func (c *Config) LoadDomainsFromGeodata() ([]string, error) {
	return geodat.LoadDomainsFromSites(c.Domains.GeoSitePath, c.Domains.GeoSiteCategories)
}
