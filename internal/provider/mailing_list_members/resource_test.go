// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mailing_list_members"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestMailingListMemberResourceSchema_HasRequiredFields(t *testing.T) {
	schema := mailing_list_members.MailingListMemberResourceSchema()

	// Verify required fields exist
	requiredFields := []string{"list_address", "member_address"}
	for _, field := range requiredFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing required '%s' attribute", field)
		}
	}

	// Verify optional fields exist
	optionalFields := []string{"name", "subscribed", "vars"}
	for _, field := range optionalFields {
		if schema.Attributes[field] == nil {
			t.Errorf("Schema missing optional '%s' attribute", field)
		}
	}

	// Verify computed fields exist
	computedFields := []string{"id"}
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

func TestMailingListMembersDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := mailing_list_members.MailingListMembersDataSourceSchema()

	// Verify required fields exist
	if schema.Attributes["list_address"] == nil {
		t.Error("Schema missing required 'list_address' attribute")
	}

	// Verify computed fields exist
	computedFields := []string{"total_count", "members"}
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

func TestAccMailingListMemberResource_Basic(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	listName := test_helpers.RandomName("testlist")
	listAddress := fmt.Sprintf("%s@%s", listName, domainName)
	memberAddress := fmt.Sprintf("member-%d@example.com", test_helpers.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMailingListMemberDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccMailingListMemberResourceConfig(listAddress, memberAddress, "Test Member"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "list_address", listAddress),
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "member_address", memberAddress),
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "name", "Test Member"),
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "subscribed", "true"),
					resource.TestCheckResourceAttrSet("mailgun_mailing_list_member.test", "id"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "mailgun_mailing_list_member.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update testing
			{
				Config: testAccMailingListMemberResourceConfig(listAddress, memberAddress, "Updated Name"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "name", "Updated Name"),
				),
			},
		},
	})
}

func TestAccMailingListMemberResource_Unsubscribed(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	listName := test_helpers.RandomName("testlist-unsub")
	listAddress := fmt.Sprintf("%s@%s", listName, domainName)
	memberAddress := fmt.Sprintf("member-%d@example.com", test_helpers.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMailingListMemberDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMailingListMemberResourceConfigUnsubscribed(listAddress, memberAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_mailing_list_member.test", "subscribed", "false"),
				),
			},
		},
	})
}

func TestAccMailingListMembersDataSource(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := os.Getenv("MAILGUN_TEST_DOMAIN")
	if domainName == "" {
		t.Skip("MAILGUN_TEST_DOMAIN environment variable is not set")
	}

	listName := test_helpers.RandomName("testlist-ds")
	listAddress := fmt.Sprintf("%s@%s", listName, domainName)
	memberAddress := fmt.Sprintf("member-%d@example.com", test_helpers.RandomInt())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccMailingListMembersDataSourceConfig(listAddress, memberAddress),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_mailing_list_members.test", "list_address", listAddress),
					resource.TestCheckResourceAttrSet("data.mailgun_mailing_list_members.test", "total_count"),
				),
			},
		},
	})
}

func testAccMailingListMemberResourceConfig(listAddress, memberAddress, name string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_mailing_list" "test" {
  address = "%s"
  name    = "Test List"
}

resource "mailgun_mailing_list_member" "test" {
  list_address   = mailgun_mailing_list.test.address
  member_address = "%s"
  name           = "%s"
  subscribed     = true
}
`, os.Getenv("MAILGUN_API_KEY"), listAddress, memberAddress, name)
}

func testAccMailingListMemberResourceConfigUnsubscribed(listAddress, memberAddress string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_mailing_list" "test" {
  address = "%s"
  name    = "Test List"
}

resource "mailgun_mailing_list_member" "test" {
  list_address   = mailgun_mailing_list.test.address
  member_address = "%s"
  subscribed     = false
}
`, os.Getenv("MAILGUN_API_KEY"), listAddress, memberAddress)
}

func testAccMailingListMembersDataSourceConfig(listAddress, memberAddress string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_mailing_list" "test" {
  address = "%s"
  name    = "Test List"
}

resource "mailgun_mailing_list_member" "test" {
  list_address   = mailgun_mailing_list.test.address
  member_address = "%s"
}

data "mailgun_mailing_list_members" "test" {
  list_address = mailgun_mailing_list.test.address
  depends_on   = [mailgun_mailing_list_member.test]
}
`, os.Getenv("MAILGUN_API_KEY"), listAddress, memberAddress)
}

func testAccCheckMailingListMemberDestroy(s *terraform.State) error {
	// Member deletion is handled by the provider
	return nil
}
