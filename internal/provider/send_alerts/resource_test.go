// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/send_alerts"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestSendAlertResourceSchema_HasRequiredFields(t *testing.T) {
	schema := send_alerts.SendAlertResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"name", "metric", "comparator", "limit", "dimension"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "parent_account_id", "subaccount_id", "account_group", "created_at", "updated_at", "last_checked"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"description", "period", "alert_channels", "filters"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY

func TestAccSendAlertResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	if os.Getenv("MAILGUN_TEST_SEND_ALERTS_ENABLED") == "" {
		t.Skip("MAILGUN_TEST_SEND_ALERTS_ENABLED environment variable is not set (send alerts feature may not be available on account)")
	}

	alertName := fmt.Sprintf("tf-test-alert-%d", os.Getpid())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendAlertResourceConfig_Basic(alertName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("mailgun_send_alert.test", "id"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "name", alertName),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "metric", "hard_bounce_rate"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "comparator", ">"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "limit", "0.05"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "dimension", "domain"),
				),
			},
			// Import test
			{
				ResourceName:      "mailgun_send_alert.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccSendAlertResource_WithOptionalFields(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	if os.Getenv("MAILGUN_TEST_SEND_ALERTS_ENABLED") == "" {
		t.Skip("MAILGUN_TEST_SEND_ALERTS_ENABLED environment variable is not set (send alerts feature may not be available on account)")
	}

	alertName := fmt.Sprintf("tf-test-alert-full-%d", os.Getpid())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendAlertResourceConfig_WithOptionalFields(alertName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("mailgun_send_alert.test", "id"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "name", alertName),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "metric", "delivered_rate"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "comparator", "<"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "limit", "0.95"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "dimension", "domain"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "description", "Test alert for delivery rate"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "period", "1d"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "alert_channels.#", "1"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "alert_channels.0", "email"),
				),
			},
		},
	})
}

func TestAccSendAlertResource_Update(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	if os.Getenv("MAILGUN_TEST_SEND_ALERTS_ENABLED") == "" {
		t.Skip("MAILGUN_TEST_SEND_ALERTS_ENABLED environment variable is not set (send alerts feature may not be available on account)")
	}

	alertName := fmt.Sprintf("tf-test-alert-update-%d", os.Getpid())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendAlertResourceConfig_Basic(alertName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "limit", "0.05"),
				),
			},
			{
				Config: testAccSendAlertResourceConfig_Updated(alertName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "limit", "0.10"),
					resource.TestCheckResourceAttr("mailgun_send_alert.test", "description", "Updated description"),
				),
			},
		},
	})
}

func testAccSendAlertResourceConfig_Basic(name string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_send_alert" "test" {
  name       = "%s"
  metric     = "hard_bounce_rate"
  comparator = ">"
  limit      = "0.05"
  dimension  = "domain"
}
`, os.Getenv("MAILGUN_API_KEY"), name)
}

func testAccSendAlertResourceConfig_WithOptionalFields(name string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_send_alert" "test" {
  name           = "%s"
  metric         = "delivered_rate"
  comparator     = "<"
  limit          = "0.95"
  dimension      = "domain"
  description    = "Test alert for delivery rate"
  period         = "1d"
  alert_channels = ["email"]
}
`, os.Getenv("MAILGUN_API_KEY"), name)
}

func testAccSendAlertResourceConfig_Updated(name string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_send_alert" "test" {
  name        = "%s"
  metric      = "hard_bounce_rate"
  comparator  = ">"
  limit       = "0.10"
  dimension   = "domain"
  description = "Updated description"
}
`, os.Getenv("MAILGUN_API_KEY"), name)
}
