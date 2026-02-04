// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package subaccounts_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/subaccounts"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestSubaccountDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := subaccounts.SubaccountDataSourceSchema()

	// Verify required fields exist
	requiredFields := []string{"id"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"name", "status"}
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

func TestSubaccountsListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := subaccounts.SubaccountsListDataSourceSchema()

	// Verify optional fields exist
	optionalFields := []string{"enabled"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"subaccounts", "total_count"}
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

// Acceptance Tests - These tests require MAILGUN_API_KEY and real subaccounts

func TestAccSubaccountDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	subaccountID := os.Getenv("MAILGUN_TEST_SUBACCOUNT_ID")
	if subaccountID == "" {
		t.Skip("MAILGUN_TEST_SUBACCOUNT_ID environment variable is not set (requires existing subaccount)")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubaccountDataSourceConfig(subaccountID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_subaccount.test", "id", subaccountID),
					resource.TestCheckResourceAttrSet("data.mailgun_subaccount.test", "name"),
					resource.TestCheckResourceAttrSet("data.mailgun_subaccount.test", "status"),
				),
			},
		},
	})
}

func TestAccSubaccountsListDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// This test just lists subaccounts - doesn't require any specific subaccount to exist
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubaccountsListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_subaccounts.test", "total_count"),
				),
			},
		},
	})
}

func TestAccSubaccountsListDataSource_FilterEnabled(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSubaccountsListDataSourceConfigEnabled(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_subaccounts.test", "total_count"),
				),
			},
		},
	})
}

func testAccSubaccountDataSourceConfig(subaccountID string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_subaccount" "test" {
  id = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), subaccountID)
}

func testAccSubaccountsListDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_subaccounts" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}

func testAccSubaccountsListDataSourceConfigEnabled() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_subaccounts" "test" {
  enabled = true
}
`, os.Getenv("MAILGUN_API_KEY"))
}
