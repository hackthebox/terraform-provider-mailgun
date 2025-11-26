// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package api_keys_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/api_keys"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestApiKeyResourceSchema_HasRequiredFields(t *testing.T) {
	schema := api_keys.ApiKeyResourceSchema()

	// Verify key fields exist
	requiredFields := []string{"role"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	optionalFields := []string{"description", "domain_name", "kind", "expiration"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	computedFields := []string{"id", "secret", "created_at", "updated_at", "expires_at", "is_disabled", "disabled_reason", "requestor", "user_name"}
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

func TestApiKeyDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := api_keys.ApiKeyDataSourceSchema()

	// Verify key fields exist
	if schema.Attributes["id"] == nil {
		t.Error("Schema missing required 'id' attribute")
	}

	computedFields := []string{"role", "kind", "description", "domain_name", "created_at", "updated_at", "expires_at", "is_disabled", "disabled_reason", "requestor", "user_name"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}
}

func TestApiKeysListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := api_keys.ApiKeysListDataSourceSchema()

	// Verify optional filter fields exist
	if schema.Attributes["domain_name"] == nil {
		t.Error("Schema missing optional 'domain_name' attribute")
	}
	if schema.Attributes["kind"] == nil {
		t.Error("Schema missing optional 'kind' attribute")
	}

	// Verify computed fields
	if schema.Attributes["keys"] == nil {
		t.Error("Schema missing computed 'keys' attribute")
	}
	if schema.Attributes["total_count"] == nil {
		t.Error("Schema missing computed 'total_count' attribute")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccApiKeyResource_Sending(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// Note: This test requires an existing domain for sending keys
	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckApiKeyDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccApiKeyResourceConfig_Sending(domainName, "Terraform test sending key"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_api_key.test", "role", "sending"),
					resource.TestCheckResourceAttr("mailgun_api_key.test", "kind", "domain"),
					resource.TestCheckResourceAttr("mailgun_api_key.test", "domain_name", domainName),
					resource.TestCheckResourceAttr("mailgun_api_key.test", "description", "Terraform test sending key"),
					resource.TestCheckResourceAttrSet("mailgun_api_key.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_api_key.test", "secret"),
					resource.TestCheckResourceAttrSet("mailgun_api_key.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mailgun_api_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"secret", "expiration"}, // Secret cannot be imported
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccApiKeysDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccApiKeysListDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_api_keys.test", "total_count"),
				),
			},
		},
	})
}

func testAccApiKeyResourceConfig_Sending(domain, description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_api_key" "test" {
  role        = "sending"
  kind        = "domain"
  domain_name = "%s"
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, description)
}

func testAccApiKeysListDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_api_keys" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}

func testAccCheckApiKeyDestroy(s *terraform.State) error {
	// API key deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
