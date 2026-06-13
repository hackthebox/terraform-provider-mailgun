// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"net"
	"testing"
)

// RFC 5737 documentation ranges (never routed, safe for tests).
var docRangesForTest = []string{"192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24"}

func parseDocNets(t *testing.T) []*net.IPNet {
	t.Helper()
	nets := make([]*net.IPNet, 0, len(docRangesForTest))
	for _, cidr := range docRangesForTest {
		_, n, err := net.ParseCIDR(cidr)
		if err != nil {
			t.Fatalf("bad doc range %q: %v", cidr, err)
		}
		nets = append(nets, n)
	}
	return nets
}

func TestRandomDocIP_WithinDocumentationRange(t *testing.T) {
	nets := parseDocNets(t)
	for range 200 {
		got := RandomDocIP()
		ip := net.ParseIP(got)
		if ip == nil {
			t.Fatalf("RandomDocIP() = %q, not a valid IP", got)
		}
		inRange := false
		for _, n := range nets {
			if n.Contains(ip) {
				inRange = true
				break
			}
		}
		if !inRange {
			t.Errorf("RandomDocIP() = %q, not in any RFC 5737 documentation range", got)
		}
	}
}

func TestRandomDocIP_ProducesVariety(t *testing.T) {
	seen := make(map[string]struct{})
	for range 50 {
		seen[RandomDocIP()] = struct{}{}
	}
	if len(seen) < 2 {
		t.Errorf("RandomDocIP() produced %d distinct values over 50 calls, expected variety", len(seen))
	}
}

func TestRandomDocCIDR_WithinDocumentationRange(t *testing.T) {
	nets := parseDocNets(t)
	for range 200 {
		got := RandomDocCIDR()
		ip, network, err := net.ParseCIDR(got)
		if err != nil {
			t.Fatalf("RandomDocCIDR() = %q, not a valid CIDR: %v", got, err)
		}
		if network.String() != got {
			t.Errorf("RandomDocCIDR() = %q is not a canonical (aligned) network address; got %q", got, network.String())
		}
		inRange := false
		for _, n := range nets {
			if n.Contains(ip) {
				inRange = true
				break
			}
		}
		if !inRange {
			t.Errorf("RandomDocCIDR() = %q, not in any RFC 5737 documentation range", got)
		}
	}
}

func TestRandomDocCIDR_ProducesVariety(t *testing.T) {
	seen := make(map[string]struct{})
	for range 50 {
		seen[RandomDocCIDR()] = struct{}{}
	}
	if len(seen) < 2 {
		t.Errorf("RandomDocCIDR() produced %d distinct values over 50 calls, expected variety", len(seen))
	}
}
