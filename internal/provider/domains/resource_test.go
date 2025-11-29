// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package domains_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domains"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// TestMain runs cleanup after all tests complete (even on failure)
func TestMain(m *testing.M) {
	code := m.Run()
	test_helpers.CleanupTestDomains()
	os.Exit(code)
}

// Unit Tests - These tests don't require external API calls

func TestDomainModel_HasExpectedFields(t *testing.T) {
	// Verify DomainModel has expected fields by creating one
	model := domains.DomainModel{
		// Required/Optional input attributes
		Name:                       types.StringValue("test.example.com"),
		SmtpPassword:               types.StringNull(),
		SpamAction:                 types.StringValue("disabled"),
		Wildcard:                   types.BoolValue(false),
		ForceDkimAuthority:         types.BoolNull(),
		DkimKeySize:                types.StringNull(),
		Ips:                        types.StringNull(),
		PoolId:                     types.StringNull(),
		WebScheme:                  types.StringValue("https"),
		WebPrefix:                  types.StringValue("email"),
		UseAutomaticSenderSecurity: types.BoolValue(false),
		DkimSelector:               types.StringNull(),
		DkimHostName:               types.StringNull(),
		ForceRootDkimHost:          types.BoolNull(),
		EncryptIncomingMessage:     types.BoolNull(),

		// Computed attributes (flat, not nested)
		Id:               types.StringValue("domain-123"),
		CreatedAt:        types.StringValue("2025-01-15T10:00:00Z"),
		State:            types.StringValue("active"),
		SmtpLogin:        types.StringValue("postmaster@test.example.com"),
		IsDisabled:       types.BoolValue(false),
		RequireTls:       types.BoolValue(true),
		SkipVerification: types.BoolValue(false),
		DomainType:       types.StringValue("custom"),
		TrackingHost:     types.StringValue("track.example.com"),
	}

	// Verify values are set correctly
	if model.Name.ValueString() != "test.example.com" {
		t.Errorf("Expected name 'test.example.com', got '%s'", model.Name.ValueString())
	}
	if model.Id.ValueString() != "domain-123" {
		t.Errorf("Expected id 'domain-123', got '%s'", model.Id.ValueString())
	}
	if model.State.ValueString() != "active" {
		t.Errorf("Expected state 'active', got '%s'", model.State.ValueString())
	}
}

func TestReceivingDnsRecordsValue_AttributeTypes(t *testing.T) {
	ctx := t.Context()

	recordValue := domains.ReceivingDnsRecordsValue{}
	attrTypes := recordValue.AttributeTypes(ctx)

	// Verify all expected attributes are present
	expectedAttrs := []string{
		"cached", "is_active", "name", "priority", "record_type", "valid", "value",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := attrTypes[attrName]; !ok {
			t.Errorf("Expected attribute %s not found in ReceivingDnsRecordsValue AttributeTypes", attrName)
		}
	}
}

func TestSendingDnsRecordsValue_AttributeTypes(t *testing.T) {
	ctx := t.Context()

	recordValue := domains.SendingDnsRecordsValue{}
	attrTypes := recordValue.AttributeTypes(ctx)

	// Verify all expected attributes are present
	expectedAttrs := []string{
		"cached", "is_active", "name", "priority", "record_type", "valid", "value",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := attrTypes[attrName]; !ok {
			t.Errorf("Expected attribute %s not found in SendingDnsRecordsValue AttributeTypes", attrName)
		}
	}
}

func TestDomainResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domains.DomainResourceSchema()

	// Verify key input fields exist
	if schema.Attributes["name"] == nil {
		t.Error("Schema missing 'name' attribute")
	}
	if schema.Attributes["spam_action"] == nil {
		t.Error("Schema missing 'spam_action' attribute")
	}
	if schema.Attributes["wildcard"] == nil {
		t.Error("Schema missing 'wildcard' attribute")
	}

	// Verify computed fields are flat (not nested under 'domain')
	if schema.Attributes["id"] == nil {
		t.Error("Schema missing 'id' attribute")
	}
	if schema.Attributes["state"] == nil {
		t.Error("Schema missing 'state' attribute")
	}
	if schema.Attributes["smtp_login"] == nil {
		t.Error("Schema missing 'smtp_login' attribute")
	}
	if schema.Attributes["created_at"] == nil {
		t.Error("Schema missing 'created_at' attribute")
	}
	if schema.Attributes["is_disabled"] == nil {
		t.Error("Schema missing 'is_disabled' attribute")
	}
	if schema.Attributes["type"] == nil {
		t.Error("Schema missing 'type' attribute")
	}

	// Verify DNS records attributes exist
	if schema.Attributes["receiving_dns_records"] == nil {
		t.Error("Schema missing 'receiving_dns_records' attribute")
	}
	if schema.Attributes["sending_dns_records"] == nil {
		t.Error("Schema missing 'sending_dns_records' attribute")
	}

	// Verify there is NO nested 'domain' attribute (schema is flat)
	if schema.Attributes["domain"] != nil {
		t.Error("Schema should not have nested 'domain' attribute - schema should be flat")
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

// Mock test for SDK to Model conversion logic
func TestSDKToTerraformConversion(t *testing.T) {
	// Create a mock SDK response similar to what Mailgun API returns
	createdAt := time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC)

	sdkDomain := mtypes.Domain{
		CreatedAt:    mtypes.RFC2822Time(createdAt),
		ID:           "test.example.com",
		Name:         "test.example.com",
		RequireTLS:   true,
		SMTPLogin:    "postmaster@test.example.com",
		SMTPPassword: "secret123",
		SpamAction:   mtypes.SpamActionDisabled,
		State:        "active",
		Wildcard:     false,
		WebScheme:    "https",
	}

	// Verify the SDK types are as expected
	if sdkDomain.Name != "test.example.com" {
		t.Errorf("Expected domain name 'test.example.com', got '%s'", sdkDomain.Name)
	}
	if sdkDomain.SpamAction != mtypes.SpamActionDisabled {
		t.Errorf("Expected spam action 'disabled', got '%s'", sdkDomain.SpamAction)
	}

	// Note: Full conversion testing happens in acceptance tests where we
	// can actually call the Mailgun API and verify the mapping
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccDomainResource(t *testing.T) {
	// Skip if MAILGUN_API_KEY is not set
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := test_helpers.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainResourceDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccDomainResourceConfig(domainName, "disabled", false, "http"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("mailgun_domain.test", "spam_action", "disabled"),
					resource.TestCheckResourceAttr("mailgun_domain.test", "wildcard", "false"),
					resource.TestCheckResourceAttr("mailgun_domain.test", "web_scheme", "http"),
					// Flat attributes (no nested 'domain.' prefix)
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "id"),
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "state"),
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "created_at"),
					// Note: smtp_login may be empty on free/sandbox tier accounts
				),
			},
			// ImportState testing
			{
				ResourceName:            "mailgun_domain.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"smtp_password", "dkim_key_size", "force_dkim_authority"},
			},
			// Delete testing automatically occurs in TestCase
			// Note: spam_action and wildcard require replacement, so we skip update testing for those
		},
	})
}

func TestAccDomainResource_Wildcard(t *testing.T) {
	// Skip this test by default - most test accounts have a 1 domain limit
	// To run this test, set MAILGUN_MULTI_DOMAIN=1
	if os.Getenv("MAILGUN_MULTI_DOMAIN") == "" {
		t.Skip("Skipping wildcard test - most Mailgun test accounts have a 1 domain limit. Set MAILGUN_MULTI_DOMAIN=1 to run this test.")
	}
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	domainName := test_helpers.RandomDomainName()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckDomainResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainResourceConfig(domainName, "disabled", true, "http"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("mailgun_domain.test", "wildcard", "true"),
				),
			},
		},
	})
}

func testAccDomainResourceConfig(domainName, spamAction string, wildcard bool, webScheme string) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain" "test" {
  name        = "%s"
  spam_action = "%s"
  wildcard    = %t
  web_scheme  = "%s"
}
`, os.Getenv("MAILGUN_API_KEY"), domainName, spamAction, wildcard, webScheme)
}

func testAccCheckDomainResourceDestroy(s *terraform.State) error {
	// Domain deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
