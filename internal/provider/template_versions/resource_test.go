// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package template_versions_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/template_versions"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestTemplateVersionResourceSchema_HasRequiredFields(t *testing.T) {
	schema := template_versions.TemplateVersionResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "template_name", "template"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"tag", "engine", "comment", "active"}
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

func TestTemplateVersionsDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := template_versions.TemplateVersionsDataSourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "template_name"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "versions"}
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

func TestAccTemplateVersionResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	templateName := test_helpers.RandomName("test-template")
	versionTag := "v1"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateVersionDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccTemplateVersionResourceConfig(domainName, templateName, versionTag),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_template_version.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_template_version.test", "template_name", templateName),
					resource.TestCheckResourceAttr("mailgun_template_version.test", "tag", versionTag),
					resource.TestCheckResourceAttr("mailgun_template_version.test", "engine", "handlebars"),
					resource.TestCheckResourceAttrSet("mailgun_template_version.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_template_version.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_template_version.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update template content testing
			{
				Config: testAccTemplateVersionResourceConfigUpdated(domainName, templateName, versionTag),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_template_version.test", "comment", "Updated version"),
				),
			},
		},
	})
}

func TestAccTemplateVersionResource_ActiveAttribute(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	templateName := test_helpers.RandomName("test-template-active")

	// Test that the active attribute is correctly read from the API.
	// We don't set active=true because that creates a version that can't be deleted
	// (Mailgun doesn't allow deleting active versions, and there's no API to deactivate).
	// The new version will have active=false since the template's initial version is active.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckTemplateVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateVersionResourceConfigWithActive(domainName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// New versions are not active by default (template's initial version is active)
					resource.TestCheckResourceAttr("mailgun_template_version.test", "active", "false"),
				),
			},
		},
	})
}

func TestAccTemplateVersionsDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	templateName := test_helpers.RandomName("test-template-ds")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTemplateVersionsDataSourceConfig(domainName, templateName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_template_versions.test", "domain", domainName),
					resource.TestCheckResourceAttr("data.mailgun_template_versions.test", "template_name", templateName),
					resource.TestCheckResourceAttrSet("data.mailgun_template_versions.test", "total_count"),
				),
			},
		},
	})
}

func testAccTemplateVersionResourceConfig(domain, templateName, versionTag string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "Test template for version testing"
  template    = "<html><body>Initial template content</body></html>"
}

resource "mailgun_template_version" "test" {
  domain        = mailgun_template.test.domain
  template_name = mailgun_template.test.name
  tag           = "%s"
  template      = "<html><body>Hello {{name}}!</body></html>"
  engine        = "handlebars"
  comment       = "Test version"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, templateName, versionTag)
}

func testAccTemplateVersionResourceConfigUpdated(domain, templateName, versionTag string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "Test template for version testing"
  template    = "<html><body>Initial template content</body></html>"
}

resource "mailgun_template_version" "test" {
  domain        = mailgun_template.test.domain
  template_name = mailgun_template.test.name
  tag           = "%s"
  template      = "<html><body>Hello {{name}}! Updated content.</body></html>"
  engine        = "handlebars"
  comment       = "Updated version"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, templateName, versionTag)
}

// testAccTemplateVersionResourceConfigWithActive creates a template with a second version
// to test the active attribute. The second version will be inactive (active=false).
func testAccTemplateVersionResourceConfigWithActive(domain, templateName string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "Test template for active attribute testing"
  template    = "<html><body>Initial version content</body></html>"
}

resource "mailgun_template_version" "test" {
  domain        = mailgun_template.test.domain
  template_name = mailgun_template.test.name
  tag           = "v2"
  template      = "<html><body>Second version content</body></html>"
  engine        = "handlebars"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, templateName)
}

func testAccTemplateVersionsDataSourceConfig(domain, templateName string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_template" "test" {
  domain      = "%s"
  name        = "%s"
  description = "Test template for data source"
  template    = "<html><body>Initial</body></html>"
}

data "mailgun_template_versions" "test" {
  domain        = mailgun_template.test.domain
  template_name = mailgun_template.test.name
}
`, os.Getenv("MAILGUN_API_KEY"), domain, templateName)
}

func testAccCheckTemplateVersionDestroy(s *terraform.State) error {
	// Template version deletion is handled by the provider
	return nil
}
