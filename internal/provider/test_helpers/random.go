// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"fmt"
	"math/rand"
)

// RandomInt returns a random integer for generating unique test resource names
func RandomInt() int {
	return rand.Intn(100000)
}

// RandomDomainName generates a unique domain name for testing
func RandomDomainName() string {
	return fmt.Sprintf("test-%d.example.com", RandomInt())
}

// RandomString generates a random string of the specified length
func RandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// RandomName generates a unique name with a given prefix for testing
func RandomName(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, RandomInt())
}

var docRangePrefixes = []string{"192.0.2", "198.51.100", "203.0.113"}

// RandomDocIP returns a random IPv4 address in an RFC 5737 documentation range,
// giving 768 distinct values to avoid collisions on the shared test account.
func RandomDocIP() string {
	prefix := docRangePrefixes[rand.Intn(len(docRangePrefixes))]
	return fmt.Sprintf("%s.%d", prefix, rand.Intn(256))
}

// RandomDocCIDR returns a random aligned /30 CIDR in an RFC 5737 documentation
// range, giving 192 distinct values to avoid collisions on the shared test account.
func RandomDocCIDR() string {
	prefix := docRangePrefixes[rand.Intn(len(docRangePrefixes))]
	base := rand.Intn(64) * 4
	return fmt.Sprintf("%s.%d/30", prefix, base)
}
