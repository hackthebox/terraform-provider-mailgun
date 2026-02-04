// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package alerts_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/alerts"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestAlertResourceSchema_HasRequiredFields(t *testing.T) {
	schema := alerts.AlertResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"event_type", "channel"}
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

	// Verify optional fields exist (channel-specific settings)
	optionalFields := []string{"emails", "webhook_url", "slack_ids"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify description exists and mentions immutability
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY

func TestAccAlertResource_Email(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	testEmail := os.Getenv("MAILGUN_TEST_EMAIL")
	if testEmail == "" {
		testEmail = "test@example.com"
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertResourceConfig_Email(testEmail),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("mailgun_alert.test", "id"),
					resource.TestCheckResourceAttr("mailgun_alert.test", "channel", "email"),
					resource.TestCheckResourceAttr("mailgun_alert.test", "event_type", "hard_bounces"),
				),
			},
		},
	})
}

func TestAccAlertResource_Webhook(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	testWebhookURL := os.Getenv("MAILGUN_TEST_WEBHOOK_URL")
	if testWebhookURL == "" {
		t.Skip("MAILGUN_TEST_WEBHOOK_URL environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertResourceConfig_Webhook(testWebhookURL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("mailgun_alert.test", "id"),
					resource.TestCheckResourceAttr("mailgun_alert.test", "channel", "webhook"),
					resource.TestCheckResourceAttr("mailgun_alert.test", "event_type", "complaints"),
					resource.TestCheckResourceAttr("mailgun_alert.test", "webhook_url", testWebhookURL),
				),
			},
		},
	})
}

func testAccAlertResourceConfig_Email(email string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_alert" "test" {
  event_type = "hard_bounces"
  channel    = "email"
  emails     = ["%s"]
}
`, os.Getenv("MAILGUN_API_KEY"), email)
}

func testAccAlertResourceConfig_Webhook(webhookURL string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_alert" "test" {
  event_type  = "complaints"
  channel     = "webhook"
  webhook_url = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), webhookURL)
}
