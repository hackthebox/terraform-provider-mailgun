// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package templates_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/templates"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestTemplateResourceSchema_HasRequiredFields(t *testing.T) {
	schema := templates.TemplateResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "name"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"description", "template", "engine", "comment"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "created_at", "version_tag", "version_active"}
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

func TestTemplatesDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := templates.TemplatesDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "templates"}
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

func TestAccTemplateResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	templateName := test_helpers.RandomName("test-template")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTemplateResourceConfig(domainName, templateName, "Test template description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_template.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_template.test", "name", templateName),
					resource.TestCheckResourceAttr("mailgun_template.test", "description", "Test template description"),
					resource.TestCheckResourceAttrSet("mailgun_template.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_template.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update description testing
			{
				Config: testAccTemplateResourceConfig(domainName, templateName, "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_template.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccTemplateResource_WithContent(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	templateName := test_helpers.RandomName("test-template-content")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateDestroy,
		Steps: []resource.TestStep{
			// Create with template content
			{
				Config: testAccTemplateResourceConfigWithContent(domainName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_template.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_template.test", "name", templateName),
					resource.TestCheckResourceAttr("mailgun_template.test", "engine", "handlebars"),
					resource.TestCheckResourceAttrSet("mailgun_template.test", "version_tag"),
				),
			},
		},
	})
}

func TestAccTemplatesDataSource(t *testing.T) {
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
				Config: testAccTemplatesDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_templates.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_templates.test", "total_count"),
				),
			},
		},
	})
}

func testAccTemplateResourceConfig(domain, name, description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, name, description)
}

func testAccTemplateResourceConfigWithContent(domain, name string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "Template with content"
  template    = "<html><body>Hello {{name}}!</body></html>"
  engine      = "handlebars"
  comment     = "Initial version"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, name)
}

func testAccTemplatesDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_templates" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckTemplateDestroy(s *terraform.State) error {
	// Template deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
