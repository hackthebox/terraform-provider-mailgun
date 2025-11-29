// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package routes_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/routes"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// TestMain runs cleanup after all tests complete (even on failure)
func TestMain(m *testing.M) {
	code := m.Run()
	test_helpers.CleanupTestRoutes()
	os.Exit(code)
}

// Unit Tests - These tests don't require external API calls

func TestRouteResourceSchema_HasRequiredFields(t *testing.T) {
	schema := routes.RouteResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"expression", "actions"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"priority", "description"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "created_at"}
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

func TestRoutesDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := routes.RoutesDataSourceSchema()

	// Verify optional fields exist
	if schema.Attributes["limit"] == nil {
		t.Error("Schema missing optional 'limit' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "routes"}
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

func TestAccRouteResource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	randomSuffix := test_helpers.RandomString(8)
	description := fmt.Sprintf("test-route-%s", randomSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccRouteResourceConfig(description, 10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_route.test", "description", description),
					resource.TestCheckResourceAttr("mailgun_route.test", "priority", "10"),
					resource.TestCheckResourceAttrSet("mailgun_route.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_route.test", "expression"),
					resource.TestCheckResourceAttrSet("mailgun_route.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_route.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update priority testing
			{
				Config: testAccRouteResourceConfig(description, 20),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_route.test", "description", description),
					resource.TestCheckResourceAttr("mailgun_route.test", "priority", "20"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccRouteResource_WithMultipleActions(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	randomSuffix := test_helpers.RandomString(8)
	description := fmt.Sprintf("test-route-multi-%s", randomSuffix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRouteResourceConfigMultipleActions(description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_route.test", "description", description),
					resource.TestCheckResourceAttrSet("mailgun_route.test", "id"),
					// Check that actions list has 2 items
					resource.TestCheckResourceAttr("mailgun_route.test", "actions.#", "2"),
				),
			},
		},
	})
}

func TestAccRoutesDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_routes.test", "total_count"),
				),
			},
		},
	})
}

func TestAccRoutesDataSource_WithLimit(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRoutesDataSourceConfigWithLimit(5),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_routes.test", "total_count"),
				),
			},
		},
	})
}

func testAccRouteResourceConfig(description string, priority int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_route" "test" {
  expression  = "catch_all()"
  actions     = ["stop()"]
  priority    = %d
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), priority, description)
}

func testAccRouteResourceConfigMultipleActions(description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_route" "test" {
  expression  = "catch_all()"
  actions     = ["forward(\"https://example.com/webhook\")", "stop()"]
  priority    = 100
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), description)
}

func testAccRoutesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_routes" "test" {
}
`, os.Getenv("MAILGUN_API_KEY"))
}

func testAccRoutesDataSourceConfigWithLimit(limit int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_routes" "test" {
  limit = %d
}
`, os.Getenv("MAILGUN_API_KEY"), limit)
}

func testAccCheckRouteDestroy(s *terraform.State) error {
	// Route deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
