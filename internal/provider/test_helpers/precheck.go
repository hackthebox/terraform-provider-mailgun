// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"os"
	"testing"
)

// AccPreCheck validates that required environment variables are set for acceptance tests
func AccPreCheck(t *testing.T) {
	if v := os.Getenv("MAILGUN_API_KEY"); v == "" {
		t.Fatal("MAILGUN_API_KEY must be set for acceptance tests")
	}
}
