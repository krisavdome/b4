package sni

import (
	"net"
	"testing"

	"github.com/daniellavrushin/b4/config"
)

func TestMatchIP_BasicFunctionality(t *testing.T) {
	// Create test set config
	testSet := &config.SetConfig{
		Name:    "test-set",
		Enabled: true,
		Targets: config.TargetsConfig{
			IpsToMatch: []string{
				"192.168.1.0/24",
				"10.0.0.1",
				"2001:db8::/32",
			},
		},
	}

	// Build matcher
	matcher := NewSuffixSet([]*config.SetConfig{testSet})

	// Test cases
	tests := []struct {
		name     string
		ip       string
		expected bool
	}{
		{"Match subnet", "192.168.1.100", true},
		{"Match single IP", "10.0.0.1", true},
		{"Match IPv6", "2001:db8::1", true},
		{"No match", "8.8.8.8", false},
		{"No match different subnet", "192.168.2.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			if ip == nil {
				t.Fatalf("Failed to parse IP: %s", tt.ip)
			}

			matched, set := matcher.MatchIP(ip)
			if matched != tt.expected {
				t.Errorf("MatchIP(%s) = %v, want %v", tt.ip, matched, tt.expected)
			}

			if matched && set != testSet {
				t.Errorf("MatchIP(%s) returned wrong set", tt.ip)
			}
		})
	}
}

func TestMatchIP_LargeIPList(t *testing.T) {
	// Simulate the real-world scenario with large IP list
	largeIPList := make([]string, 0, 90000)

	// Add 89,457 IPs similar to the user's example
	for i := 0; i < 255; i++ {
		for j := 0; j < 255; j++ {
			largeIPList = append(largeIPList, "10."+string(rune(i))+"."+string(rune(j))+".0/24")
			if len(largeIPList) >= 89457 {
				break
			}
		}
		if len(largeIPList) >= 89457 {
			break
		}
	}

	testSet := &config.SetConfig{
		Name:    "large-set",
		Enabled: true,
		Targets: config.TargetsConfig{
			IpsToMatch: largeIPList,
		},
	}

	// Build matcher (should be fast with radix tree)
	matcher := NewSuffixSet([]*config.SetConfig{testSet})

	// Verify it works
	testIP := net.ParseIP("10.1.1.1")
	matched, _ := matcher.MatchIP(testIP)
	if !matched {
		t.Error("Expected to match IP in large list")
	}

	// Test non-matching IP
	nonMatchIP := net.ParseIP("8.8.8.8")
	matched, _ = matcher.MatchIP(nonMatchIP)
	if matched {
		t.Error("Expected not to match IP outside large list")
	}
}

func TestMatchIP_NilHandling(t *testing.T) {
	testSet := &config.SetConfig{
		Name:    "test-set",
		Enabled: true,
		Targets: config.TargetsConfig{
			IpsToMatch: []string{"192.168.1.0/24"},
		},
	}

	matcher := NewSuffixSet([]*config.SetConfig{testSet})

	// Test nil IP
	matched, _ := matcher.MatchIP(nil)
	if matched {
		t.Error("Expected no match for nil IP")
	}

	// Test nil matcher
	var nilMatcher *SuffixSet
	matched, _ = nilMatcher.MatchIP(net.ParseIP("192.168.1.1"))
	if matched {
		t.Error("Expected no match for nil matcher")
	}
}
