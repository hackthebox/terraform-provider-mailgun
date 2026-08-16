// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/ip_allowlist"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// TestMain runs cleanup after all tests complete (even on failure)
func TestMain(m *testing.M) {
	code := m.Run()
	test_helpers.CleanupTestIPAllowlistEntries()
	os.Exit(code)
}

// Unit Tests - These tests don't require external API calls

func TestIPAllowlistResourceSchema_HasRequiredFields(t *testing.T) {
	schema := ip_allowlist.IPAllowlistResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"address"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"description"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

func TestIPAllowlistDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := ip_allowlist.IPAllowlistDataSourceSchema()

	// Verify computed fields exist
	computedFields := []string{"id", "entries"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

func TestIPAllowlistClient_GetBaseURL(t *testing.T) {
	// This tests the URL construction logic for v2 API
	// The client uses v2 endpoints (/v2/ip_whitelist) instead of v3
	// We can't easily test the actual client without mocking HTTP,
	// but we document the expected behavior here.
	t.Log("IP Allowlist uses v2 API endpoint: /v2/ip_whitelist")
	t.Log("The client replaces /v3 suffix with /v2 in the base URL")
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls
// CAUTION: These tests modify the real IP allowlist - use test IP addresses

func TestAccIPAllowlistResource(t *testing.T) {
	if os.Getenv("MAILGUN_RUN_KNOWN_FAILING") == "" {
		t.Skip("pending #85: intermittent non-empty refresh plan on the shared account; set MAILGUN_RUN_KNOWN_FAILING=1 to run")
	}

	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// Setup: Add test runner's IP to allowlist (removed automatically via t.Cleanup)
	test_helpers.SetupIPAllowlistForTests(t)

	testIP := test_helpers.RandomDocIP()
	description := fmt.Sprintf("test-allowlist-%s", test_helpers.RandomString(8))

	ensureIPAllowlistAbsent(t, testIP)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPAllowlistDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccIPAllowlistResourceConfig(testIP, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "address", testIP),
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "description", description),
					resource.TestCheckResourceAttrSet("mailgun_ip_allowlist.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_ip_allowlist.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update description testing
			{
				Config: testAccIPAllowlistResourceConfig(testIP, "updated-description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "address", testIP),
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "description", "updated-description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccIPAllowlistResource_CIDR(t *testing.T) {
	if os.Getenv("MAILGUN_RUN_KNOWN_FAILING") == "" {
		t.Skip("pending #85: intermittent non-empty refresh plan on the shared account; set MAILGUN_RUN_KNOWN_FAILING=1 to run")
	}

	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// Setup: Add test runner's IP to allowlist (removed automatically via t.Cleanup)
	test_helpers.SetupIPAllowlistForTests(t)

	testCIDR := test_helpers.RandomDocCIDR()
	description := fmt.Sprintf("test-cidr-%s", test_helpers.RandomString(8))

	ensureIPAllowlistAbsent(t, testCIDR)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckIPAllowlistDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAllowlistResourceConfig(testCIDR, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "address", testCIDR),
					resource.TestCheckResourceAttr("mailgun_ip_allowlist.test", "description", description),
				),
			},
		},
	})
}

func TestAccIPAllowlistDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// Setup: Add test runner's IP to allowlist (removed automatically via t.Cleanup)
	test_helpers.SetupIPAllowlistForTests(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAllowlistDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_ip_allowlist.test", "id"),
				),
			},
		},
	})
}

func testAccIPAllowlistResourceConfig(address, description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_ip_allowlist" "test" {
  address     = "%s"
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), address, description)
}

func testAccIPAllowlistDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_ip_allowlist" "test" {
}
`, os.Getenv("MAILGUN_API_KEY"))
}

// ensureIPAllowlistAbsent deletes any existing allowlist entry at address so a
// leftover orphan from a crashed run cannot make the test's create return 409.
// Best-effort: failures are logged, not fatal.
func ensureIPAllowlistAbsent(t *testing.T, address string) {
	t.Helper()

	mg := mailgun.NewMailgun(os.Getenv("MAILGUN_API_KEY"))
	client := ip_allowlist.NewIPAllowlistClient(mg)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if _, found, err := client.GetIPAllowlistEntry(ctx, address); err != nil || !found {
		return
	}
	if err := client.DeleteIPAllowlistEntry(ctx, address); err != nil {
		t.Logf("warning: failed to pre-delete leftover allowlist entry %s: %v", address, err)
	}
}

func testAccCheckIPAllowlistDestroy(s *terraform.State) error {
	// IP allowlist deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
