// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
)

// GetPublicIP retrieves the current machine's public IP address.
func GetPublicIP(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org", nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get public IP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return strings.TrimSpace(string(body)), nil
}

// testRunnerDescriptionPrefix is used to identify IPs added by the test framework
const testRunnerDescriptionPrefix = "terraform-provider-test-runner-"

// SetupIPAllowlistForTests adds the test runner's IP to the Mailgun allowlist
// and returns the IP address. Uses t.Cleanup() to ensure the IP is removed
// even if the test fails or panics.
//
// If the IP is already in the allowlist with a test-runner description (from a
// previous crashed test run), it will be cleaned up after the test.
// If the IP was manually added (different description), it won't be removed.
func SetupIPAllowlistForTests(t *testing.T) string {
	t.Helper()

	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		t.Fatal("MAILGUN_API_KEY environment variable is required")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()

	// Get current public IP
	currentIP, err := GetPublicIP(ctx)
	if err != nil {
		t.Fatalf("Failed to get public IP: %v", err)
	}

	// Create Mailgun client and IP allowlist client
	mg := mailgun.NewMailgun(apiKey)
	client := ip_allowlist.NewIPAllowlistClient(mg)

	// Check if IP is already in allowlist
	existingEntry, err := client.GetIPAllowlistEntry(ctx, currentIP)
	shouldCleanup := true

	if err == nil {
		// IP already exists - check if it was added by a previous test run
		if strings.HasPrefix(existingEntry.Description, testRunnerDescriptionPrefix) {
			// This was added by a previous test run that didn't clean up - we'll clean it up
			t.Logf("Test runner IP %s already in allowlist from previous test run, will clean up after test", currentIP)
		} else {
			// This was manually added, don't clean it up
			t.Logf("Test runner IP %s already in allowlist (manually added), skipping setup and cleanup", currentIP)
			shouldCleanup = false
		}
	} else {
		// IP not in allowlist, add it
		description := fmt.Sprintf("%s%d", testRunnerDescriptionPrefix, time.Now().Unix())
		if err := client.CreateIPAllowlistEntry(ctx, currentIP, description); err != nil {
			t.Fatalf("Failed to add test runner IP to allowlist: %v", err)
		}
		t.Logf("Added test runner IP %s to allowlist", currentIP)
	}

	// Register cleanup to remove IP after test completes (if we should clean up)
	if shouldCleanup {
		// Note: Using context.Background() here because t.Context() is cancelled before cleanup runs
		t.Cleanup(func() {
			cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 30*time.Second) //nolint:usetesting // t.Context() is cancelled before cleanup
			defer cleanupCancel()

			if err := client.DeleteIPAllowlistEntry(cleanupCtx, currentIP); err != nil {
				t.Logf("Warning: Failed to remove test runner IP from allowlist: %v", err)
			} else {
				t.Logf("Removed test runner IP %s from allowlist", currentIP)
			}
		})
	}

	return currentIP
}
