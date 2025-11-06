export type FakingStrategy =
  | "ttl"
  | "pastseq"
  | "randseq"
  | "tcp_check"
  | "md5sum";
export enum FakingPayloadType {
  RANDOM = 0,
  CUSTOM = 1,
  DEFAULT = 2,
}
export interface IFaking {
  strategy: FakingStrategy;
  sni: boolean;
  ttl: number;
  seq_offset: number;
  sni_seq_length: number;
  sni_type: FakingPayloadType;
  custom_payload: string;
}

export type FragmentationStrategy = "tcp" | "ip" | "none";
export interface IFragmentation {
  strategy: FragmentationStrategy;
  sni_position: number;
  sni_reverse: boolean;
  middle_sni: boolean;
}

export enum LogLevel {
  ERROR = 0,
  INFO = 1,
  TRACE = 2,
  DEBUG = 3,
}
export interface ILogging {
  level: LogLevel;
  instaflush: boolean;
  syslog: boolean;
}

export interface IDomainConfig {
  sni_domains: string[];
  geosite_path: string;
  geoip_path: string;
  geosite_categories: string[];
  geoip_categories: string[];
  block_domains: string[];
  block_geosite_categories: string[];
}

export interface IDomainStatistics {
  manual_domains: number;
  geosite_domains: number;
  total_domains: number;
  category_breakdown?: Record<string, number>;
  geosite_available: boolean;
}

export interface ICategoryPreview {
  category: string;
  total_domains: number;
  preview_count: number;
  preview: string[];
}

export type UdpMode = "drop" | "fake";
export type UdpFilterQuicMode = "disabled" | "all" | "parse";
export type UdpFakingStrategy = "none" | "ttl" | "checksum";
export interface IUdpConfig {
  mode: UdpMode;
  fake_seq_length: number;
  fake_len: number;
  faking_strategy: UdpFakingStrategy;
  dport_min: number;
  dport_max: number;
  filter_quic: UdpFilterQuicMode;
  conn_bytes_limit: number;
  filter_stun: boolean;
}
export interface IQueueConfig {
  start_num: number;
  threads: number;
  mark: number;
  ipv4: boolean;
  ipv6: boolean;
}

export interface IB4Config {
  queue: IQueueConfig;
  domains: IDomainConfig;
  system: ISystemConfig;
}

export interface ICheckerConfig {
  timeout: number;
  max_concurrent: number;
  domains: string[];
}

export interface ITcpConfig {
  conn_bytes_limit: number;
  seg2delay: number;
}

export interface IBypassConfig {
  tcp: ITcpConfig;
  udp: IUdpConfig;
  fragmentation: IFragmentation;
  faking: IFaking;
}
export interface IWebServerConfig {
  port: number;
}
export interface ITableConfig {
  monitor_interval: number;
  skip_setup: false;
}
export interface ISystemConfig {
  logging: ILogging;
  web_server: IWebServerConfig;
  tables: ITableConfig;
  checker: ICheckerConfig;
}

export default class B4Config implements IB4Config {
  queue: IQueueConfig = {
    start_num: 0,
    threads: 4,
    mark: 32768,
    ipv4: true,
    ipv6: false,
  };

  system: ISystemConfig = {
    logging: {
      level: LogLevel.INFO,
      instaflush: true,
      syslog: false,
    },
    web_server: {
      port: 7000,
    },
    tables: {
      monitor_interval: 10,
      skip_setup: false,
    },
    checker: {
      timeout: 15,
      max_concurrent: 4,
      domains: [],
    },
  };

  bypass: IBypassConfig = {
    tcp: {
      conn_bytes_limit: 19,
      seg2delay: 0,
    },
    udp: {
      mode: "fake",
      fake_seq_length: 6,
      fake_len: 64,
      faking_strategy: "none",
      dport_min: 0,
      dport_max: 0,
      filter_quic: "disabled",
      conn_bytes_limit: 8,
      filter_stun: true,
    },
    fragmentation: {
      strategy: "tcp",
      sni_position: 5,
      sni_reverse: true,
      middle_sni: true,
    },
    faking: {
      strategy: "ttl",
      sni: true,
      ttl: 64,
      seq_offset: 1000,
      sni_seq_length: 6,
      sni_type: FakingPayloadType.DEFAULT,
      custom_payload: "",
    },
  };

  domains: IDomainConfig = {
    sni_domains: [],
    geosite_path: "",
    geoip_path: "",
    geosite_categories: [],
    geoip_categories: [],
    block_domains: [],
    block_geosite_categories: [],
  };
}
