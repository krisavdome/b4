package sni

import (
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"github.com/daniellavrushin/b4/config"
	"github.com/yl2chen/cidranger"
)

type ipRange struct {
	ipNet *net.IPNet
	set   *config.SetConfig
}

// rangerEntry implements cidranger.RangerEntry
type rangerEntry struct {
	ipNet *net.IPNet
	set   *config.SetConfig
}

func (r *rangerEntry) Network() net.IPNet {
	return *r.ipNet
}

type portRange struct {
	min int
	max int
	set *config.SetConfig
}

type SuffixSet struct {
	sets       map[string]*config.SetConfig
	regexes    []*regexWithSet
	regexCache sync.Map
	ipRanger   cidranger.Ranger // Optimized radix tree for IP lookups
	portRanges []portRange
}

type regexWithSet struct {
	regex *regexp.Regexp
	set   *config.SetConfig
}

type cacheEntry struct {
	matched bool
	set     *config.SetConfig
}

func NewSuffixSet(sets []*config.SetConfig) *SuffixSet {
	s := &SuffixSet{
		sets:     make(map[string]*config.SetConfig),
		regexes:  make([]*regexWithSet, 0),
		ipRanger: cidranger.NewPCTrieRanger(), // Initialize radix tree
	}

	seenRegexes := make(map[string]bool)

	for _, set := range sets {
		if !set.Enabled {
			continue
		}

		for _, d := range set.Targets.DomainsToMatch {
			d = strings.ToLower(strings.TrimSpace(d))
			if d == "" {
				continue
			}

			// Handle regex patterns
			if strings.HasPrefix(d, "regexp:") {
				pattern := strings.TrimPrefix(d, "regexp:")
				if seenRegexes[pattern] {
					continue
				}
				if re, err := regexp.Compile(pattern); err == nil {
					s.regexes = append(s.regexes, &regexWithSet{regex: re, set: set})
					seenRegexes[pattern] = true
				}
				continue
			}

			// Regular domain
			d = strings.TrimRight(d, ".")
			if _, exists := s.sets[d]; !exists {
				s.sets[d] = set
			}
		}

		for _, ipStr := range set.Targets.IpsToMatch {
			ipStr = strings.TrimSpace(ipStr)
			if ipStr == "" {
				continue
			}

			// Parse CIDR or single IP
			var ipNet *net.IPNet
			var err error

			if strings.Contains(ipStr, "/") {
				_, ipNet, err = net.ParseCIDR(ipStr)
			} else {
				ip := net.ParseIP(ipStr)
				if ip != nil {
					if ip.To4() != nil {
						_, ipNet, _ = net.ParseCIDR(ipStr + "/32")
					} else {
						_, ipNet, _ = net.ParseCIDR(ipStr + "/128")
					}
				}
			}

			if err == nil && ipNet != nil {
				// Insert into radix tree for O(log n) lookups
				entry := &rangerEntry{ipNet: ipNet, set: set}
				s.ipRanger.Insert(entry)
			}
		}

		if set.UDP.DPortFilter != "" {
			ports := strings.Split(set.UDP.DPortFilter, ",")
			for _, part := range ports {
				part = strings.TrimSpace(part)
				if part == "" {
					continue
				}

				if strings.Contains(part, "-") {
					bounds := strings.SplitN(part, "-", 2)
					if len(bounds) == 2 {
						min, err1 := strconv.Atoi(bounds[0])
						max, err2 := strconv.Atoi(bounds[1])
						if err1 == nil && err2 == nil {
							if min >= 0 && max >= 0 && min <= max {
								s.portRanges = append(s.portRanges, portRange{min: min, max: max, set: set})
							}
						}
					}
				} else {
					port, err := strconv.Atoi(part)
					if err == nil && port >= 0 {
						s.portRanges = append(s.portRanges, portRange{min: port, max: port, set: set})
					}
				}
			}
		}
	}

	return s
}

func (s *SuffixSet) MatchUDPPort(dport uint16) (bool, *config.SetConfig) {
	if s == nil || len(s.portRanges) == 0 {
		return false, nil
	}

	port := int(dport)

	for _, r := range s.portRanges {
		matched := false

		if r.min > 0 && r.max > 0 {
			matched = port >= r.min && port <= r.max
		} else if r.min > 0 {
			matched = port >= r.min
		} else if r.max > 0 {
			matched = port <= r.max
		}

		if matched {
			return true, r.set
		}
	}

	return false, nil
}

func (s *SuffixSet) MatchSNI(host string) (bool, *config.SetConfig) {
	if s == nil || (len(s.sets) == 0 && len(s.regexes) == 0) || host == "" {
		return false, nil
	}

	lower := strings.ToLower(host)

	// Check exact/suffix match first (fast)
	if matched, set := s.matchDomain(lower); matched {
		return true, set
	}

	if len(s.regexes) > 0 {
		return s.matchRegex(lower)
	}

	return false, nil
}

func (s *SuffixSet) MatchIP(ip net.IP) (bool, *config.SetConfig) {
	if s == nil || s.ipRanger == nil || ip == nil {
		return false, nil
	}

	// Use radix tree for O(log n) lookup instead of O(n) linear search
	contains, err := s.ipRanger.Contains(ip)
	if err != nil || !contains {
		return false, nil
	}

	// Get the matching entry to retrieve the set config
	entries, err := s.ipRanger.ContainingNetworks(ip)
	if err != nil || len(entries) == 0 {
		return false, nil
	}

	// Return the first matching set (most specific match)
	if entry, ok := entries[0].(*rangerEntry); ok {
		return true, entry.set
	}

	return false, nil
}

func (s *SuffixSet) matchDomain(host string) (bool, *config.SetConfig) {
	// Exact match
	if set, ok := s.sets[host]; ok {
		return true, set
	}

	// Check suffixes
	for {
		idx := strings.IndexByte(host, '.')
		if idx == -1 {
			break
		}
		host = host[idx+1:]
		if set, ok := s.sets[host]; ok {
			return true, set
		}
	}

	return false, nil
}

func (s *SuffixSet) matchRegex(host string) (bool, *config.SetConfig) {
	// Check cache
	if cached, ok := s.regexCache.Load(host); ok {
		entry := cached.(cacheEntry)
		return entry.matched, entry.set
	}

	// Test patterns
	for _, rws := range s.regexes {
		if rws.regex.MatchString(host) {
			s.regexCache.Store(host, cacheEntry{matched: true, set: rws.set})
			return true, rws.set
		}
	}

	s.regexCache.Store(host, cacheEntry{matched: false, set: nil})
	return false, nil
}
