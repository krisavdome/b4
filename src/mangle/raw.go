package mangle

import (
	"encoding/binary"
	"fmt"
	"net"
	"syscall"

	"golang.org/x/sys/unix"
)

// RawSender handles sending raw IP packets
type RawSender struct {
	fd4  int // IPv4 raw socket
	fd6  int // IPv6 raw socket
	mark int // Socket mark for routing
}

// NewRawSender creates a new raw packet sender
func NewRawSender(mark int) (*RawSender, error) {
	rs := &RawSender{
		fd4:  -1,
		fd6:  -1,
		mark: mark,
	}

	// Create IPv4 raw socket
	fd4, err := unix.Socket(unix.AF_INET, unix.SOCK_RAW, unix.IPPROTO_RAW)
	if err != nil {
		return nil, fmt.Errorf("failed to create IPv4 raw socket: %w", err)
	}
	rs.fd4 = fd4

	// Set socket mark
	if err := unix.SetsockoptInt(fd4, unix.SOL_SOCKET, unix.SO_MARK, mark); err != nil {
		unix.Close(fd4)
		return nil, fmt.Errorf("failed to set IPv4 socket mark: %w", err)
	}

	// Create IPv6 raw socket
	fd6, err := unix.Socket(unix.AF_INET6, unix.SOCK_RAW, unix.IPPROTO_RAW)
	if err != nil {
		unix.Close(fd4)
		return nil, fmt.Errorf("failed to create IPv6 raw socket: %w", err)
	}
	rs.fd6 = fd6

	// Set socket mark for IPv6
	if err := unix.SetsockoptInt(fd6, unix.SOL_SOCKET, unix.SO_MARK, mark); err != nil {
		unix.Close(fd4)
		unix.Close(fd6)
		return nil, fmt.Errorf("failed to set IPv6 socket mark: %w", err)
	}

	return rs, nil
}

// Close closes the raw sockets
func (rs *RawSender) Close() error {
	var err error
	if rs.fd4 >= 0 {
		if e := unix.Close(rs.fd4); e != nil {
			err = e
		}
		rs.fd4 = -1
	}
	if rs.fd6 >= 0 {
		if e := unix.Close(rs.fd6); e != nil {
			err = e
		}
		rs.fd6 = -1
	}
	return err
}

// SendRaw sends a raw IP packet
func (rs *RawSender) SendRaw(packet []byte) error {
	if len(packet) == 0 {
		return fmt.Errorf("empty packet")
	}

	ipVersion := packet[0] >> 4

	switch ipVersion {
	case 4:
		return rs.sendIPv4(packet)
	case 6:
		return rs.sendIPv6(packet)
	default:
		return fmt.Errorf("unknown IP version: %d", ipVersion)
	}
}

// sendIPv4 sends an IPv4 packet
func (rs *RawSender) sendIPv4(packet []byte) error {
	if len(packet) < 20 {
		return fmt.Errorf("IPv4 packet too short")
	}

	// Extract destination IP
	dstIP := packet[16:20]

	// Create sockaddr
	addr := &unix.SockaddrInet4{
		Port: 0, // Raw sockets ignore port
	}
	copy(addr.Addr[:], dstIP)

	// Send packet
	err := unix.Sendto(rs.fd4, packet, 0, addr)
	if err != nil {
		return fmt.Errorf("sendto failed: %w", err)
	}

	return nil
}

// sendIPv6 sends an IPv6 packet
func (rs *RawSender) sendIPv6(packet []byte) error {
	if len(packet) < 40 {
		return fmt.Errorf("IPv6 packet too short")
	}

	// Extract destination IP
	dstIP := packet[24:40]

	// Create sockaddr
	addr := &unix.SockaddrInet6{
		Port: 0, // Raw sockets ignore port
	}
	copy(addr.Addr[:], dstIP)

	// Send packet
	err := unix.Sendto(rs.fd6, packet, 0, addr)
	if err != nil {
		return fmt.Errorf("sendto failed: %w", err)
	}

	return nil
}

// SendWithFragmentation sends a packet, automatically fragmenting if needed
func (rs *RawSender) SendWithFragmentation(packet []byte, mtu int) error {
	if len(packet) <= mtu {
		return rs.SendRaw(packet)
	}

	// Need to fragment
	ipVersion := packet[0] >> 4

	if ipVersion != 4 {
		// IPv6 fragmentation is more complex and typically handled by kernel
		return rs.SendRaw(packet)
	}

	// Calculate fragment size (must be multiple of 8)
	fragmentSize := (mtu - 20) &^ 7 // Subtract IP header, round down to 8-byte boundary

	ihl := int(packet[0]&0x0F) * 4
	totalLen := int(binary.BigEndian.Uint16(packet[2:4]))
	payload := packet[ihl:totalLen]

	// Send fragments
	offset := 0
	fragID := uint16(syscall.Getpid() & 0xFFFF) // Use PID as fragment ID

	for offset < len(payload) {
		fragPayloadSize := fragmentSize
		if offset+fragPayloadSize > len(payload) {
			fragPayloadSize = len(payload) - offset
		}

		isLastFragment := (offset + fragPayloadSize) >= len(payload)

		// Create fragment
		fragLen := ihl + fragPayloadSize
		frag := make([]byte, fragLen)

		// Copy IP header
		copy(frag, packet[:ihl])

		// Copy payload fragment
		copy(frag[ihl:], payload[offset:offset+fragPayloadSize])

		// Update IP header
		binary.BigEndian.PutUint16(frag[2:4], uint16(fragLen)) // Total length
		binary.BigEndian.PutUint16(frag[4:6], fragID)          // Fragment ID

		// Set fragment offset and flags
		fragOffset := uint16(offset / 8) // Offset in 8-byte units
		var flags uint16
		if !isLastFragment {
			flags = 0x2000 // More Fragments flag
		}
		fragOffsetField := flags | fragOffset
		binary.BigEndian.PutUint16(frag[6:8], fragOffsetField)

		// Recalculate checksum
		SetIPv4Checksum(frag[:ihl])

		// Send fragment
		if err := rs.sendIPv4(frag); err != nil {
			return fmt.Errorf("failed to send fragment: %w", err)
		}

		offset += fragPayloadSize
	}

	return nil
}

// CreateSender creates a sender function for use with packet processing
func (rs *RawSender) CreateSender() func([]byte) error {
	return func(packet []byte) error {
		return rs.SendRaw(packet)
	}
}

// CreateSenderWithMTU creates a sender that handles fragmentation
func (rs *RawSender) CreateSenderWithMTU(mtu int) func([]byte) error {
	return func(packet []byte) error {
		return rs.SendWithFragmentation(packet, mtu)
	}
}

// GetLocalIP gets a local IP address for the given destination
func GetLocalIP(dst net.IP) (net.IP, error) {
	// Create a UDP connection to determine local IP
	// This doesn't actually send anything
	conn, err := net.Dial("udp", net.JoinHostPort(dst.String(), "80"))
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP, nil
}
