// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials_test

import (
	"fmt"
	"os"
	"testing"

	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/smtp_credentials"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestSmtpCredentialResourceSchema_HasRequiredFields(t *testing.T) {
	schema := smtp_credentials.SmtpCredentialResourceSchema()

	// Verify key fields exist
	requiredFields := []string{"domain", "login"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	computedFields := []string{"id", "full_login", "created_at"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}

	passwordAttr, ok := schema.Attributes["password"].(rschema.StringAttribute)
	if !ok {
		t.Fatal("Schema missing string 'password' attribute")
	}

	if !passwordAttr.Optional {
		t.Error("Password should be optional to support imported credentials")
	}

	if !passwordAttr.Computed {
		t.Error("Password should be computed to preserve state for imported credentials")
	}
}

func TestSmtpCredentialDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := smtp_credentials.SmtpCredentialDataSourceSchema()

	// Verify key fields exist
	requiredFields := []string{"domain", "login"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	computedFields := []string{"id", "full_login", "created_at"}
	for _, field := range computedFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing computed '%s' attribute", field)
		}
	}
}

func TestSmtpCredentialsListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := smtp_credentials.SmtpCredentialsListDataSourceSchema()

	// Verify key fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}
	if schema.Attributes["credentials"] == nil {
		t.Error("Schema missing computed 'credentials' attribute")
	}
	if schema.Attributes["total_count"] == nil {
		t.Error("Schema missing computed 'total_count' attribute")
	}
}

func TestSmtpCredentialResourceSchema_WriteOnlyPassword(t *testing.T) {
	schema := smtp_credentials.SmtpCredentialResourceSchema()

	legacy, ok := schema.Attributes["password"].(rschema.StringAttribute)
	if !ok {
		t.Fatal("Schema missing string 'password' attribute")
	}
	if legacy.DeprecationMessage == "" {
		t.Error("password should be deprecated in favor of password_wo")
	}

	wo, ok := schema.Attributes["password_wo"].(rschema.StringAttribute)
	if !ok {
		t.Fatal("Schema missing string 'password_wo' attribute")
	}
	if !wo.WriteOnly {
		t.Error("password_wo must be WriteOnly")
	}
	if !wo.Optional {
		t.Error("password_wo must be Optional")
	}
	if !wo.Sensitive {
		t.Error("password_wo must be Sensitive")
	}
	if wo.Computed {
		t.Error("password_wo must not be Computed (WriteOnly forbids Computed)")
	}

	if _, ok := schema.Attributes["password_wo_version"].(rschema.Int64Attribute); !ok {
		t.Fatal("Schema missing Int64 'password_wo_version' attribute")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccSmtpCredentialResource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	// Note: This test requires an existing domain
	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	loginName := test_helpers.RandomString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSmtpCredentialDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccSmtpCredentialResourceConfig(domainName, loginName, "initial-password-123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "login", loginName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "full_login", fmt.Sprintf("%s@%s", loginName, domainName)),
					resource.TestCheckResourceAttrSet("mailgun_smtp_credential.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_smtp_credential.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mailgun_smtp_credential.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password"}, // Password cannot be imported
			},
			// Ensure an imported credential remains stable when password is omitted
			{
				Config: testAccSmtpCredentialImportedResourceConfig(domainName, loginName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "login", loginName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "full_login", fmt.Sprintf("%s@%s", loginName, domainName)),
				),
			},
			// Update password testing
			{
				Config: testAccSmtpCredentialResourceConfig(domainName, loginName, "updated-password-456"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "login", loginName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.test", "password", "updated-password-456"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccSmtpCredentialsDataSource(t *testing.T) {
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
				Config: testAccSmtpCredentialsListDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_smtp_credentials.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_smtp_credentials.test", "total_count"),
				),
			},
		},
	})
}

func testAccSmtpCredentialResourceConfig(domain, login, password string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_smtp_credential" "test" {
  domain   = "%s"
  login    = "%s"
  password = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, login, password)
}

func testAccSmtpCredentialImportedResourceConfig(domain, login string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_smtp_credential" "test" {
  domain = "%s"
  login  = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, login)
}

func testAccSmtpCredentialsListDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_smtp_credentials" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckSmtpCredentialDestroy(s *terraform.State) error {
	// Credential deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}

func testAccSmtpCredentialWriteOnlyConfig(domain, login, password string, version int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_smtp_credential" "wo" {
  domain              = "%s"
  login               = "%s"
  password_wo         = "%s"
  password_wo_version = %d
}
`, os.Getenv("MAILGUN_API_KEY"), domain, login, password, version)
}

func testAccSmtpCredentialMigrationLegacyConfig(domain, login, password string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_smtp_credential" "migrate" {
  domain   = "%s"
  login    = "%s"
  password = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, login, password)
}

func testAccSmtpCredentialMigrationWriteOnlyConfig(domain, login, password string, version int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_smtp_credential" "migrate" {
  domain              = "%s"
  login               = "%s"
  password_wo         = "%s"
  password_wo_version = %d
}
`, os.Getenv("MAILGUN_API_KEY"), domain, login, password, version)
}

// Migrating an existing credential from the deprecated stateful password to
// password_wo must drop password from state without tripping Terraform's
// "known planned value must be identical in the new state" rule.
func TestAccSmtpCredentialResource_LegacyToWriteOnlyMigration(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	loginName := test_helpers.RandomString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSmtpCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSmtpCredentialMigrationLegacyConfig(domainName, loginName, "legacy-initial-123"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.migrate", "password", "legacy-initial-123"),
				),
			},
			{
				Config: testAccSmtpCredentialMigrationWriteOnlyConfig(domainName, loginName, "wo-migrated-456", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.migrate", "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr("mailgun_smtp_credential.migrate", "password_wo"),
					resource.TestCheckNoResourceAttr("mailgun_smtp_credential.migrate", "password"),
				),
			},
			{
				Config:   testAccSmtpCredentialMigrationWriteOnlyConfig(domainName, loginName, "wo-migrated-456", 1),
				PlanOnly: true,
			},
		},
	})
}

func TestAccSmtpCredentialResource_WriteOnly(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}
	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	loginName := test_helpers.RandomString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckSmtpCredentialDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSmtpCredentialWriteOnlyConfig(domainName, loginName, "wo-initial-123", 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.wo", "login", loginName),
					resource.TestCheckResourceAttr("mailgun_smtp_credential.wo", "password_wo_version", "1"),
					resource.TestCheckNoResourceAttr("mailgun_smtp_credential.wo", "password_wo"),
					resource.TestCheckResourceAttrSet("mailgun_smtp_credential.wo", "created_at"),
				),
			},
			{
				Config:   testAccSmtpCredentialWriteOnlyConfig(domainName, loginName, "wo-initial-123", 1),
				PlanOnly: true,
			},
			{
				Config: testAccSmtpCredentialWriteOnlyConfig(domainName, loginName, "wo-rotated-456", 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_smtp_credential.wo", "password_wo_version", "2"),
				),
			},
		},
	})
}
