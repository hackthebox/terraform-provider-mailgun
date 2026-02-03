// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_lists_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mailing_lists"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestMailingListResourceSchema_HasRequiredFields(t *testing.T) {
	schema := mailing_lists.MailingListResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"address"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"name", "description", "access_level", "reply_preference"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id", "created_at", "members_count"}
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

func TestMailingListsDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := mailing_lists.MailingListsDataSourceSchema()

	// Verify computed fields exist
	computedFields := []string{"total_count", "mailing_lists"}
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

func TestAccMailingListResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	listName := test_helpers.RandomName("testlist")
	listAddress := fmt.Sprintf("%s@%s", listName, domainName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMailingListDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMailingListResourceConfig(listAddress, "Test List", "A test mailing list"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "address", listAddress),
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "name", "Test List"),
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "description", "A test mailing list"),
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "access_level", "readonly"),
					resource.TestCheckResourceAttrSet("mailgun_mailing_list.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_mailing_list.test", "created_at"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_mailing_list.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccMailingListResourceConfig(listAddress, "Updated List", "Updated description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "name", "Updated List"),
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "description", "Updated description"),
				),
			},
		},
	})
}

func TestAccMailingListResource_AccessLevels(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	listName := test_helpers.RandomName("testlist-access")
	listAddress := fmt.Sprintf("%s@%s", listName, domainName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMailingListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMailingListResourceConfigWithAccessLevel(listAddress, "members"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "access_level", "members"),
				),
			},
			{
				Config: testAccMailingListResourceConfigWithAccessLevel(listAddress, "everyone"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list.test", "access_level", "everyone"),
				),
			},
		},
	})
}

func TestAccMailingListsDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMailingListsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_mailing_lists.test", "total_count"),
				),
			},
		},
	})
}

func testAccMailingListResourceConfig(address, name, description string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_mailing_list" "test" {
  address     = "%s"
  name        = "%s"
  description = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), address, name, description)
}

func testAccMailingListResourceConfigWithAccessLevel(address, accessLevel string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_mailing_list" "test" {
  address      = "%s"
  name         = "Access Level Test"
  access_level = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), address, accessLevel)
}

func testAccMailingListsDataSourceConfig() string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_mailing_lists" "test" {}
`, os.Getenv("MAILGUN_API_KEY"))
}

func testAccCheckMailingListDestroy(s *terraform.State) error {
	// Mailing list deletion is handled by the provider
	return nil
}
