// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package test_helpers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v5"
	"github.com/mailgun/mailgun-go/v5/mtypes"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
)

// CleanupTestDomains removes all domains matching the test pattern (test-*.example.com)
func CleanupTestDomains() {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mg := mailgun.NewMailgun(apiKey)
	iter := mg.ListDomains(nil)
	var domains []mtypes.Domain

	page := iter.Next(ctx, &domains)
	for page {
		for _, domain := range domains {
			if strings.HasPrefix(domain.Name, "test-") && strings.HasSuffix(domain.Name, ".example.com") {
				fmt.Printf("Cleanup: deleting orphaned test domain %s\n", domain.Name)
				if err := mg.DeleteDomain(ctx, domain.Name); err != nil {
					fmt.Printf("Cleanup: warning - failed to delete domain %s: %v\n", domain.Name, err)
				}
			}
		}
		page = iter.Next(ctx, &domains)
	}
}

// CleanupTestRoutes removes all routes with descriptions matching test patterns
func CleanupTestRoutes() {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mg := mailgun.NewMailgun(apiKey)
	iter := mg.ListRoutes(nil)
	var routes []mtypes.Route

	page := iter.Next(ctx, &routes)
	for page {
		for _, route := range routes {
			if strings.HasPrefix(route.Description, "test-route-") {
				fmt.Printf("Cleanup: deleting orphaned test route %s (%s)\n", route.Id, route.Description)
				if err := mg.DeleteRoute(ctx, route.Id); err != nil {
					fmt.Printf("Cleanup: warning - failed to delete route %s: %v\n", route.Id, err)
				}
			}
		}
		page = iter.Next(ctx, &routes)
	}
}

// CleanupTestIPAllowlistEntries removes all IP allowlist entries with test descriptions
func CleanupTestIPAllowlistEntries() {
	apiKey := os.Getenv("MAILGUN_API_KEY")
	if apiKey == "" {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	mg := mailgun.NewMailgun(apiKey)
	client := ip_allowlist.NewIPAllowlistClient(mg)

	entries, err := client.ListIPAllowlist(ctx)
	if err != nil {
		fmt.Printf("Cleanup: warning - failed to list IP allowlist entries: %v\n", err)
		return
	}

	for _, entry := range entries {
		if strings.HasPrefix(entry.Description, "test-allowlist-") ||
			strings.HasPrefix(entry.Description, "test-cidr-") ||
			strings.HasPrefix(entry.Description, testRunnerDescriptionPrefix) {
			fmt.Printf("Cleanup: deleting orphaned IP allowlist entry %s (%s)\n", entry.IPAddress, entry.Description)
			if err := client.DeleteIPAllowlistEntry(ctx, entry.IPAddress); err != nil {
				fmt.Printf("Cleanup: warning - failed to delete IP allowlist entry %s: %v\n", entry.IPAddress, err)
			}
		}
	}
}
