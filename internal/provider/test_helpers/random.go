// Copyright (c) HashiCorp, Inc.
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
