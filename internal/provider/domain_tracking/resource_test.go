// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_tracking_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_tracking"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestDomainTrackingResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_tracking.DomainTrackingResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"click_active", "open_active", "unsubscribe_active", "unsubscribe_html_footer", "unsubscribe_text_footer"}
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

func TestDomainTrackingDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_tracking.DomainTrackingDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"click_active", "open_active", "unsubscribe_active", "unsubscribe_html_footer", "unsubscribe_text_footer"}
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

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccDomainTrackingResource_Basic(t *testing.T) {
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
		CheckDestroy:             testAccCheckDomainTrackingDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainTrackingResourceConfig(domainName, true, false, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "click_active", "true"),
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "open_active", "false"),
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "unsubscribe_active", "false"),
					resource.TestCheckResourceAttrSet("mailgun_domain_tracking.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_domain_tracking.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccDomainTrackingResourceConfig(domainName, true, true, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "click_active", "true"),
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "open_active", "true"),
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "unsubscribe_active", "true"),
				),
			},
		},
	})
}

func TestAccDomainTrackingResource_WithFooters(t *testing.T) {
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
		CheckDestroy:             testAccCheckDomainTrackingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainTrackingResourceConfigWithFooters(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_tracking.test", "unsubscribe_active", "true"),
					resource.TestCheckResourceAttrSet("mailgun_domain_tracking.test", "unsubscribe_html_footer"),
					resource.TestCheckResourceAttrSet("mailgun_domain_tracking.test", "unsubscribe_text_footer"),
				),
			},
		},
	})
}

func TestAccDomainTrackingDataSource(t *testing.T) {
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
				Config: testAccDomainTrackingDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domain_tracking.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_tracking.test", "click_active"),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_tracking.test", "open_active"),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_tracking.test", "unsubscribe_active"),
				),
			},
		},
	})
}

func testAccDomainTrackingResourceConfig(domain string, clickActive, openActive, unsubActive bool) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_tracking" "test" {
  domain             = "%s"
  click_active       = %t
  open_active        = %t
  unsubscribe_active = %t
}
`, os.Getenv("MAILGUN_API_KEY"), domain, clickActive, openActive, unsubActive)
}

func testAccDomainTrackingResourceConfigWithFooters(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_tracking" "test" {
  domain                   = "%s"
  click_active             = false
  open_active              = false
  unsubscribe_active       = true
  unsubscribe_html_footer  = "<p>Click <a href=\"%%unsubscribe_url%%\">here</a> to unsubscribe.</p>"
  unsubscribe_text_footer  = "Click %%unsubscribe_url%% to unsubscribe."
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccDomainTrackingDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_domain_tracking" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckDomainTrackingDestroy(s *terraform.State) error {
	// Tracking settings are reset on delete
	return nil
}
