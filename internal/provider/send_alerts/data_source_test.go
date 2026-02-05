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

func TestSendAlertDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := send_alerts.SendAlertDataSourceSchema()

	// Verify required input fields exist
	requiredFields := []string{"name"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "metric", "comparator", "limit", "dimension", "description", "period", "alert_channels", "filters"}
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

func TestSendAlertsListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := send_alerts.SendAlertsListDataSourceSchema()

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

// Acceptance Tests - These tests require MAILGUN_API_KEY

func TestAccSendAlertDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	alertName := os.Getenv("MAILGUN_TEST_SEND_ALERT_NAME")
	if alertName == "" {
		t.Skip("MAILGUN_TEST_SEND_ALERT_NAME environment variable is not set (requires existing send alert)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendAlertDataSourceConfig(alertName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_send_alert.test", "name", alertName),
					resource.TestCheckResourceAttrSet("data.mailgun_send_alert.test", "metric"),
					resource.TestCheckResourceAttrSet("data.mailgun_send_alert.test", "limit"),
					resource.TestCheckResourceAttrSet("data.mailgun_send_alert.test", "dimension"),
				),
			},
		},
	})
}

func TestAccSendAlertsListDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	if os.Getenv("MAILGUN_TEST_SEND_ALERTS_ENABLED") == "" {
		t.Skip("MAILGUN_TEST_SEND_ALERTS_ENABLED environment variable is not set (send alerts feature may not be available on account)")
	}

	// This test just lists alerts - doesn't require any specific alert to exist
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSendAlertsListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_send_alerts.test", "total_count"),
				),
			},
		},
	})
}

func testAccSendAlertDataSourceConfig(alertName string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_send_alert" "test" {
  name = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), alertName)
}

func testAccSendAlertsListDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_send_alerts" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}
