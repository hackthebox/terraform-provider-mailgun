// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package webhooks_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/webhooks"
)

// Unit Tests - These tests don't require external API calls

func TestWebhookResourceSchema_HasRequiredFields(t *testing.T) {
	schema := webhooks.WebhookResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "webhook_type", "urls"}
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

func TestWebhooksDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := webhooks.WebhooksDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "webhooks"}
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

func TestAccWebhookResource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	webhookURL := "https://example.com/webhook"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebhookResourceConfig(domainName, "delivered", webhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_webhook.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "webhook_type", "delivered"),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "urls.#", "1"),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "urls.0", webhookURL),
					resource.TestCheckResourceAttrSet("mailgun_webhook.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_webhook.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update URLs testing
			{
				Config: testAccWebhookResourceConfig(domainName, "delivered", "https://example.com/webhook-updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_webhook.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "webhook_type", "delivered"),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "urls.0", "https://example.com/webhook-updated"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccWebhookResource_MultipleURLs(t *testing.T) {
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
		CheckDestroy:             testAccCheckWebhookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccWebhookResourceConfigMultipleURLs(domainName, "opened"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_webhook.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "webhook_type", "opened"),
					resource.TestCheckResourceAttr("mailgun_webhook.test", "urls.#", "2"),
				),
			},
		},
	})
}

func TestAccWebhooksDataSource(t *testing.T) {
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
				Config: testAccWebhooksDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_webhooks.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_webhooks.test", "total_count"),
				),
			},
		},
	})
}

func testAccWebhookResourceConfig(domain, webhookType, url string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_webhook" "test" {
  domain       = "%s"
  webhook_type = "%s"
  urls         = ["%s"]
}
`, os.Getenv("MAILGUN_API_KEY"), domain, webhookType, url)
}

func testAccWebhookResourceConfigMultipleURLs(domain, webhookType string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_webhook" "test" {
  domain       = "%s"
  webhook_type = "%s"
  urls         = ["https://example.com/webhook1", "https://example.com/webhook2"]
}
`, os.Getenv("MAILGUN_API_KEY"), domain, webhookType)
}

func testAccWebhooksDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_webhooks" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckWebhookDestroy(s *terraform.State) error {
	// Webhook deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
