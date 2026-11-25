// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domain_sending_keys"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestDomainSendingKeyResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_sending_keys.DomainSendingKeyResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"domain"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"description", "expiration"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "secret", "created_at", "expires_at"}
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

func TestDomainSendingKeysDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := domain_sending_keys.DomainSendingKeysDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing required 'domain' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "keys"}
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

func TestAccDomainSendingKeyResource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	description := fmt.Sprintf("test-key-%s", test_helpers.RandomString(8))

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainSendingKeyDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainSendingKeyResourceConfig(domainName, description),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_sending_key.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_domain_sending_key.test", "description", description),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "secret"),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_domain_sending_key.test",
				ImportState:       true,
				ImportStateVerify: true,
				// secret, expiration cannot be retrieved after import
				ImportStateVerifyIgnore: []string{"secret", "expiration"},
			},
			// Note: Domain sending keys are immutable, so no update test
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDomainSendingKeyResource_WithExpiration(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	description := fmt.Sprintf("test-key-exp-%s", test_helpers.RandomString(8))
	// Set expiration to 1 year in seconds
	expiration := int64(31536000)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainSendingKeyDestroy,
		Steps: []resource.TestStep{
			// Create with expiration
			{
				Config: testAccDomainSendingKeyResourceConfigWithExpiration(domainName, description, expiration),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain_sending_key.test", "domain", domainName),
					resource.TestCheckResourceAttr("mailgun_domain_sending_key.test", "description", description),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "secret"),
					resource.TestCheckResourceAttrSet("mailgun_domain_sending_key.test", "expires_at"),
				),
			},
		},
	})
}

func TestAccDomainSendingKeyResource_DeletionDelayRotation(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	const address = "mailgun_domain_sending_key.test"
	description := fmt.Sprintf("test-key-rot-%s", test_helpers.RandomString(8))
	sameID := statecheck.CompareValue(compare.ValuesSame())
	var oldID string

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainSendingKeyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainSendingKeyResourceConfigWithDeletionDelay(domainName, description, 1),
				Check:  resource.TestCheckResourceAttr(address, "deletion_delay", "1"),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue(address, tfjsonpath.New("id")),
				},
			},
			// Changing the delay updates state in place and keeps the key.
			{
				Config: testAccDomainSendingKeyResourceConfigWithDeletionDelay(domainName, description, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(address, "deletion_delay", "2"),
					resource.TestCheckResourceAttrWith(address, "id", func(id string) error {
						oldID = id
						return nil
					}),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					sameID.AddStateValue(address, tfjsonpath.New("id")),
				},
			},
			// A replacement creates the new key first and deletes the old one after the delay.
			{
				Config: testAccDomainSendingKeyResourceConfigWithDeletionDelay(domainName, description+"-rotated", 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(address, plancheck.ResourceActionCreateBeforeDestroy),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet(address, "secret"),
					func(*terraform.State) error { return testAccCheckSendingKeyGone(domainName, oldID) },
				),
			},
		},
	})
}

func testAccCheckSendingKeyGone(domain, keyID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys, err := mailgun.NewMailgun(os.Getenv("MAILGUN_API_KEY")).ListAPIKeys(ctx, &mailgun.ListAPIKeysOptions{
		DomainName: domain,
		Kind:       "domain",
	})
	if err != nil {
		return fmt.Errorf("listing sending keys: %w", err)
	}
	for _, key := range keys {
		if key.ID == keyID {
			return fmt.Errorf("replaced sending key %s still exists", keyID)
		}
	}
	return nil
}

func TestAccDomainSendingKeysDataSource(t *testing.T) {
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
				Config: testAccDomainSendingKeysDataSourceConfig(domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domain_sending_keys.test", "domain", domainName),
					resource.TestCheckResourceAttrSet("data.mailgun_domain_sending_keys.test", "total_count"),
				),
			},
		},
	})
}

func testAccDomainSendingKeyResourceConfig(domain, description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_sending_key" "test" {
  domain      = "%s"
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain, description)
}

func testAccDomainSendingKeyResourceConfigWithExpiration(domain, description string, expiration int64) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_sending_key" "test" {
  domain      = "%s"
  description = "%s"
  expiration  = %d
}
`, os.Getenv("MAILGUN_API_KEY"), domain, description, expiration)
}

func testAccDomainSendingKeyResourceConfigWithDeletionDelay(domain, description string, deletionDelay int64) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain_sending_key" "test" {
  domain         = "%s"
  description    = "%s"
  deletion_delay = %d

  lifecycle {
    create_before_destroy = true
  }
}
`, os.Getenv("MAILGUN_API_KEY"), domain, description, deletionDelay)
}

func testAccDomainSendingKeysDataSourceConfig(domain string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_domain_sending_keys" "test" {
  domain = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domain)
}

func testAccCheckDomainSendingKeyDestroy(s *terraform.State) error {
	// Key deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
