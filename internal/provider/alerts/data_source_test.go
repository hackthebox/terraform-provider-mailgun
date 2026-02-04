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

func TestAlertDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := alerts.AlertDataSourceSchema()

	// Verify required fields exist
	requiredFields := []string{"id"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"event_type", "channel", "emails", "webhook_url", "slack_ids"}
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

func TestAlertsListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := alerts.AlertsListDataSourceSchema()

	// Verify computed fields exist
	computedFields := []string{"alerts", "total_count"}
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

func TestAlertEventsDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := alerts.AlertEventsDataSourceSchema()

	// Verify computed fields exist
	computedFields := []string{"events"}
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

// Acceptance Tests - These tests require MAILGUN_API_KEY

func TestAccAlertDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	alertID := os.Getenv("MAILGUN_TEST_ALERT_ID")
	if alertID == "" {
		t.Skip("MAILGUN_TEST_ALERT_ID environment variable is not set (requires existing alert)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertDataSourceConfig(alertID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_alert.test", "id", alertID),
					resource.TestCheckResourceAttrSet("data.mailgun_alert.test", "event_type"),
					resource.TestCheckResourceAttrSet("data.mailgun_alert.test", "channel"),
				),
			},
		},
	})
}

func TestAccAlertsListDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// This test just lists alerts - doesn't require any specific alert to exist
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertsListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_alerts.test", "total_count"),
				),
			},
		},
	})
}

func TestAccAlertEventsDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// This test lists available event types
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAlertEventsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_alert_events.test", "events.#"),
				),
			},
		},
	})
}

func testAccAlertDataSourceConfig(alertID string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_alert" "test" {
  id = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), alertID)
}

func testAccAlertsListDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_alerts" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}

func testAccAlertEventsDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_alert_events" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}
