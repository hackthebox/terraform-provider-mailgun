// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_ip_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_ip"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestDomainIPResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_ip.DomainIPResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "ip"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
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

func TestDomainIPsDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_ip.DomainIPsDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	if schema.Attributes["ips"] == nil {
		t.Error("Schema missing computed 'ips' attribute")
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls
// Note: These tests also require MAILGUN_TEST_IP to be set to a valid dedicated IP

func TestAccDomainIPResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	testIP := os.Getenv("MAILGUN_TEST_IP")
	if testIP == "" {
		t.Skip("MAILGUN_TEST_IP environment variable is not set (requires a dedicated IP)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainIPDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainIPResourceConfig(domainName, testIP),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_ip.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_domain_ip.test", "ip", testIP),
					resource.TestCheckResourceAttrSet("mailgun_domain_ip.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_domain_ip.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccDomainIPsDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainIPsDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domain_ips.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_ips.test", "ips.#"),
				),
			},
		},
	})
}

func testAccDomainIPResourceConfig(domain, ip string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_ip" "test" {
  domain = "%s"
  ip     = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, ip)
}

func testAccDomainIPsDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_domain_ips" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckDomainIPDestroy(s *terraform.State) error {
	// Domain IP associations are deleted via the API
	return nil
}
