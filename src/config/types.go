package config

import "github.com/daniellavrushin/b4/log"

const (
	FakePayloadRandom = iota
	FakePayloadCustom
	FakePayloadDefault
)

// QueueConfig contains netfilter queue and networking parameters
type QueueConfig struct {
	StartNum    int  `json:"start_num" bson:"start_num"`
	Threads     int  `json:"threads" bson:"threads"`
	Mark        uint `json:"mark" bson:"mark"`
	IPv4Enabled bool `json:"ipv4" bson:"ipv4"`
	IPv6Enabled bool `json:"ipv6" bson:"ipv6"`
}

// BypassConfig contains all DPI bypass strategies and settings
type BypassConfig struct {
	TCP           TCPConfig     `json:"tcp" bson:"tcp"`
	UDP           UDPConfig     `json:"udp" bson:"udp"`
	Fragmentation Fragmentation `json:"fragmentation" bson:"fragmentation"`
	Faking        Faking        `json:"faking" bson:"faking"`
}

// TCPConfig contains TCP-specific bypass settings
type TCPConfig struct {
	ConnBytesLimit int `json:"conn_bytes_limit" bson:"conn_bytes_limit"`
	Seg2Delay      int `json:"seg2delay" bson:"seg2delay"`
}

// UDPConfig contains UDP-specific bypass settings
type UDPConfig struct {
	Mode           string `json:"mode" bson:"mode"`
	FakeSeqLength  int    `json:"fake_seq_length" bson:"fake_seq_length"`
	FakeLen        int    `json:"fake_len" bson:"fake_len"`
	FakingStrategy string `json:"faking_strategy" bson:"faking_strategy"`
	DPortMin       int    `json:"dport_min" bson:"dport_min"`
	DPortMax       int    `json:"dport_max" bson:"dport_max"`
	FilterQUIC     string `json:"filter_quic" bson:"filter_quic"`
	FilterSTUN     bool   `json:"filter_stun" bson:"filter_stun"`
	ConnBytesLimit int    `json:"conn_bytes_limit" bson:"conn_bytes_limit"`
}

// Fragmentation defines packet fragmentation strategy
type Fragmentation struct {
	Strategy    string `json:"strategy" bson:"strategy"`
	SNIReverse  bool   `json:"sni_reverse" bson:"sni_reverse"`
	MiddleSNI   bool   `json:"middle_sni" bson:"middle_sni"`
	SNIPosition int    `json:"sni_position" bson:"sni_position"`
}

// Faking defines fake packet generation strategy
type Faking struct {
	SNI           bool   `json:"sni" bson:"sni"`
	TTL           uint8  `json:"ttl" bson:"ttl"`
	Strategy      string `json:"strategy" bson:"strategy"`
	SeqOffset     int32  `json:"seq_offset" bson:"seq_offset"`
	SNISeqLength  int    `json:"sni_seq_length" bson:"sni_seq_length"`
	SNIType       int    `json:"sni_type" bson:"sni_type"`
	CustomPayload string `json:"custom_payload" bson:"custom_payload"`
}

// DomainsConfig defines domain filtering rules
type DomainsConfig struct {
	GeoSitePath       string   `json:"geosite_path" bson:"geosite_path"`
	GeoIpPath         string   `json:"geoip_path" bson:"geoip_path"`
	SNIDomains        []string `json:"sni_domains" bson:"sni_domains"`
	GeoSiteCategories []string `json:"geosite_categories" bson:"geosite_categories"`
	GeoIpCategories   []string `json:"geoip_categories" bson:"geoip_categories"`
}

// SystemConfig contains infrastructure settings
type SystemConfig struct {
	Tables    TablesConfig  `json:"tables" bson:"tables"`
	Logging   Logging       `json:"logging" bson:"logging"`
	WebServer WebServer     `json:"web_server" bson:"web_server"`
	Checker   CheckerConfig `json:"checker" bson:"checker"`
}

type TablesConfig struct {
	MonitorInterval int  `json:"monitor_interval" bson:"monitor_interval"`
	SkipSetup       bool `json:"skip_setup" bson:"skip_setup"`
}

type WebServer struct {
	Port      int  `json:"port" bson:"port"`
	IsEnabled bool `json:"-" bson:"-"`
}

type CheckerConfig struct {
	TimeoutSeconds int      `json:"timeout" bson:"timeout"`
	Domains        []string `json:"domains" bson:"domains"`
	MaxConcurrent  int      `json:"max_concurrent" bson:"max_concurrent"`
}

type Logging struct {
	Level      log.Level `json:"level" bson:"level"`
	Instaflush bool      `json:"instaflush" bson:"instaflush"`
	Syslog     bool      `json:"syslog" bson:"syslog"`
}
