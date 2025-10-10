package mangle

import (
	"net"
)

// Packet verdict types
const (
	PktAccept   = 0
	PktDrop     = 1
	PktContinue = 2
)

// Faking strategies
const (
	FakeStratNone     = 0
	FakeStratRandSeq  = 1 << 0 // Random sequence number
	FakeStratTTL      = 1 << 1 // Low TTL to expire before destination
	FakeStratPastSeq  = 1 << 2 // Past sequence number
	FakeStratTCPCheck = 1 << 3 // Invalid TCP checksum
	FakeStratTCPMD5   = 1 << 4 // TCP MD5 signature option
	FakeStratUDPCheck = 1 << 5 // Invalid UDP checksum
)

// FragmentationStrategy types
const (
	FragStratNone = iota
	FragStratTCP
	FragStratIP
)

// FakePayloadType types
const (
	FakePayloadRandom = iota
	FakePayloadCustom
	FakePayloadDefault
)

// FailingStrategy describes how to invalidate a packet
type FailingStrategy struct {
	Strategy      uint32 // Bitmask of FakeStrat* constants
	FakingTTL     uint8
	RandseqOffset uint32
}

// FakeType describes how to generate a fake packet
type FakeType struct {
	Type        int    // FakePayloadType
	FakeLen     uint16 // Length of fake payload
	FakeData    []byte // Custom payload data
	SequenceLen uint   // Number of fake packets to send
	Seg2Delay   uint   // Delay in milliseconds
	Strategy    FailingStrategy
}

// PacketContext holds packet processing state
type PacketContext struct {
	RawPacket []byte
	IPHeader  []byte
	TCPHeader []byte
	Payload   []byte
	IsIPv6    bool
	SrcIP     net.IP
	DstIP     net.IP
	SrcPort   uint16
	DstPort   uint16
}

// TCPFragmentPosition defines where to split TCP packets
type TCPFragmentPosition struct {
	Offset int  // Offset relative to TCP payload start
	Delay  uint // Delay in ms before sending second fragment
}
