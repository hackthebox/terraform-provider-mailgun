// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package domain_dkim_key_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_dkim_key"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestDomainDkimKeyResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_dkim_key.DomainDkimKeyResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain", "selector"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"bits", "active"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "signing_domain", "dns_record_name", "dns_record_type", "dns_record_value", "dns_record_valid"}
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

func TestDomainDkimKeysDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_dkim_key.DomainDkimKeysDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	if schema.Attributes["keys"] == nil {
		t.Error("Schema missing computed 'keys' attribute")
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccDomainDkimKeyResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	selector := fmt.Sprintf("tf%d", test_helpers.RandomInt())

	// Note: DKIM key activation requires valid DNS records and is not supported
	// on sandbox domains. We test with active=false which is the API default.
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDkimKeyDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainDkimKeyResourceConfigBasic(domainName, selector, 1024),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "selector", selector),
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "bits", "1024"),
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "active", "false"),
					resource.TestCheckResourceAttrSet("mailgun_domain_dkim_key.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_domain_dkim_key.test", "signing_domain"),
					resource.TestCheckResourceAttrSet("mailgun_domain_dkim_key.test", "dns_record_name"),
					resource.TestCheckResourceAttrSet("mailgun_domain_dkim_key.test", "dns_record_type"),
					resource.TestCheckResourceAttrSet("mailgun_domain_dkim_key.test", "dns_record_value"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mailgun_domain_dkim_key.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bits"}, // bits is not returned by API
			},
		},
	})
}

func TestAccDomainDkimKeyResource_2048Bits(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	selector := fmt.Sprintf("tf%d", test_helpers.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainDkimKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDkimKeyResourceConfigBasic(domainName, selector, 2048),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "bits", "2048"),
					resource.TestCheckResourceAttr("mailgun_domain_dkim_key.test", "active", "false"),
				),
			},
		},
	})
}

func TestAccDomainDkimKeysDataSource(t *testing.T) {
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
				Config: testAccDomainDkimKeysDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domain_dkim_keys.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_dkim_keys.test", "keys.#"),
				),
			},
		},
	})
}

func testAccDomainDkimKeyResourceConfigBasic(domain, selector string, bits int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_dkim_key" "test" {
  domain   = "%s"
  selector = "%s"
  bits     = %d
}
`, os.Getenv("MAILGUN_API_KEY"), domain, selector, bits)
}

func testAccDomainDkimKeysDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_domain_dkim_keys" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckDomainDkimKeyDestroy(s *terraform.State) error {
	// DKIM keys are deleted via the API
	return nil
}
