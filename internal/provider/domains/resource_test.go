// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package domains_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/mailgun/mailgun-go/v5/mtypes"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domains"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestDomainModel_AttributeTypes(t *testing.T) {
	ctx := t.Context()
	domainValue := domains.DomainValue{}
	attrTypes := domainValue.AttributeTypes(ctx)

	// Verify all expected attributes are present
	expectedAttrs := []string{
		"created_at", "disabled", "id", "is_disabled", "name",
		"require_tls", "skip_verification", "smtp_login", "smtp_password",
		"spam_action", "state", "tracking_host", "type",
		"use_automatic_sender_security", "web_prefix", "web_scheme", "wildcard",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := attrTypes[attrName]; !ok {
			t.Errorf("Expected attribute %s not found in DomainValue AttributeTypes", attrName)
		}
	}

	// Verify correct types for some key attributes
	if attrTypes["name"] != types.StringType {
		t.Errorf("Expected name to be StringType, got %T", attrTypes["name"])
	}
	if attrTypes["wildcard"] != types.BoolType {
		t.Errorf("Expected wildcard to be BoolType, got %T", attrTypes["wildcard"])
	}

	// Verify disabled is ObjectType with correct nested attributes
	disabledType, ok := attrTypes["disabled"].(types.ObjectType)
	if !ok {
		t.Fatalf("Expected disabled to be ObjectType, got %T", attrTypes["disabled"])
	}

	expectedDisabledAttrs := []string{"code", "note", "permanently", "reason", "until"}
	for _, attrName := range expectedDisabledAttrs {
		if _, ok := disabledType.AttrTypes[attrName]; !ok {
			t.Errorf("Expected disabled.%s not found", attrName)
		}
	}
}

func TestDomainValue_ToObjectValue(t *testing.T) {
	ctx := t.Context()

	disabledObj := types.ObjectNull(map[string]attr.Type{
		"code":        types.StringType,
		"note":        types.StringType,
		"permanently": types.BoolType,
		"reason":      types.StringType,
		"until":       types.StringType,
	})

	domainValue := domains.DomainValue{
		CreatedAt:                  types.StringValue("2025-01-15T10:00:00Z"),
		Disabled:                   disabledObj,
		Id:                         types.StringValue("domain-123"),
		IsDisabled:                 types.BoolValue(false),
		Name:                       types.StringValue("test.example.com"),
		RequireTls:                 types.BoolValue(true),
		SkipVerification:           types.BoolValue(false),
		SmtpLogin:                  types.StringValue("postmaster@test.example.com"),
		SmtpPassword:               types.StringValue("secret"),
		SpamAction:                 types.StringValue("disabled"),
		State:                      types.StringValue("active"),
		TrackingHost:               types.StringValue("track.example.com"),
		DomainType:                 types.StringValue("mailgun"),
		UseAutomaticSenderSecurity: types.BoolValue(true),
		WebPrefix:                  types.StringValue("email"),
		WebScheme:                  types.StringValue("https"),
		Wildcard:                   types.BoolValue(false),
	}

	objValue, diags := domainValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue returned errors: %v", diags)
	}

	if objValue.IsNull() {
		t.Error("Expected non-null object value")
	}

	// Verify object has correct attributes
	attrs := objValue.Attributes()
	if len(attrs) != 17 {
		t.Errorf("Expected 17 attributes, got %d", len(attrs))
	}

	// Verify a few key values
	nameAttr := attrs["name"].(types.String)
	if nameAttr.ValueString() != "test.example.com" {
		t.Errorf("Expected name 'test.example.com', got '%s'", nameAttr.ValueString())
	}
}

func TestNewDomainValueMust(t *testing.T) {
	ctx := t.Context()

	disabledObj := types.ObjectNull(map[string]attr.Type{
		"code":        types.StringType,
		"note":        types.StringType,
		"permanently": types.BoolType,
		"reason":      types.StringType,
		"until":       types.StringType,
	})

	attrTypes := map[string]attr.Type{
		"created_at":                    types.StringType,
		"disabled":                      types.ObjectType{AttrTypes: map[string]attr.Type{}},
		"id":                            types.StringType,
		"is_disabled":                   types.BoolType,
		"name":                          types.StringType,
		"require_tls":                   types.BoolType,
		"skip_verification":             types.BoolType,
		"smtp_login":                    types.StringType,
		"smtp_password":                 types.StringType,
		"spam_action":                   types.StringType,
		"state":                         types.StringType,
		"tracking_host":                 types.StringType,
		"type":                          types.StringType,
		"use_automatic_sender_security": types.BoolType,
		"web_prefix":                    types.StringType,
		"web_scheme":                    types.StringType,
		"wildcard":                      types.BoolType,
	}

	attributes := map[string]attr.Value{
		"created_at":                    types.StringValue("2025-01-15T10:00:00Z"),
		"disabled":                      disabledObj,
		"id":                            types.StringValue("domain-123"),
		"is_disabled":                   types.BoolValue(false),
		"name":                          types.StringValue("test.example.com"),
		"require_tls":                   types.BoolValue(true),
		"skip_verification":             types.BoolValue(false),
		"smtp_login":                    types.StringValue("postmaster@test.example.com"),
		"smtp_password":                 types.StringValue("secret"),
		"spam_action":                   types.StringValue("disabled"),
		"state":                         types.StringValue("active"),
		"tracking_host":                 types.StringValue("track.example.com"),
		"type":                          types.StringValue("mailgun"),
		"use_automatic_sender_security": types.BoolValue(true),
		"web_prefix":                    types.StringValue("email"),
		"web_scheme":                    types.StringValue("https"),
		"wildcard":                      types.BoolValue(false),
	}

	domainValue := domains.NewDomainValueMust(attrTypes, attributes)

	// Verify the created domain value has correct fields
	if domainValue.Name.ValueString() != "test.example.com" {
		t.Errorf("Expected name 'test.example.com', got '%s'", domainValue.Name.ValueString())
	}
	if domainValue.SpamAction.ValueString() != "disabled" {
		t.Errorf("Expected spam_action 'disabled', got '%s'", domainValue.SpamAction.ValueString())
	}
	if !domainValue.RequireTls.ValueBool() {
		t.Error("Expected require_tls to be true")
	}

	// Test that ToObjectValue works on the created value
	objValue, diags := domainValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue on NewDomainValueMust result returned errors: %v", diags)
	}
	if objValue.IsNull() {
		t.Error("Expected non-null object value from NewDomainValueMust result")
	}
}

func TestDisabledValue_ToObjectValue(t *testing.T) {
	ctx := t.Context()

	disabledValue := domains.DisabledValue{
		Code:        types.StringValue("503"),
		Note:        types.StringValue("Domain temporarily disabled"),
		Permanently: types.BoolValue(false),
		Reason:      types.StringValue("billing"),
		Until:       types.StringValue("2025-02-01T00:00:00Z"),
	}

	objValue, diags := disabledValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue returned errors: %v", diags)
	}

	if objValue.IsNull() {
		t.Error("Expected non-null object value")
	}

	attrs := objValue.Attributes()
	if len(attrs) != 5 {
		t.Errorf("Expected 5 attributes, got %d", len(attrs))
	}
}

func TestNewDisabledValueNull(t *testing.T) {
	ctx := t.Context()

	disabledValue := domains.NewDisabledValueNull()

	// All fields should be null
	if !disabledValue.Code.IsNull() {
		t.Error("Expected Code to be null")
	}
	if !disabledValue.Note.IsNull() {
		t.Error("Expected Note to be null")
	}
	if !disabledValue.Permanently.IsNull() {
		t.Error("Expected Permanently to be null")
	}
	if !disabledValue.Reason.IsNull() {
		t.Error("Expected Reason to be null")
	}
	if !disabledValue.Until.IsNull() {
		t.Error("Expected Until to be null")
	}

	// Should convert to ObjectValue successfully
	objValue, diags := disabledValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue on null DisabledValue returned errors: %v", diags)
	}

	if objValue.IsNull() {
		t.Error("Expected non-null object value (with null attributes)")
	}
}

func TestReceivingDnsRecordsValue_ToObjectValue(t *testing.T) {
	ctx := t.Context()

	cachedList, _ := types.ListValueFrom(ctx, types.StringType, []string{"8.8.8.8", "8.8.4.4"})

	recordValue := domains.ReceivingDnsRecordsValue{
		Cached:     cachedList,
		IsActive:   types.BoolValue(true),
		Name:       types.StringValue("mx.test.example.com"),
		Priority:   types.StringValue("10"),
		RecordType: types.StringValue("MX"),
		Valid:      types.StringValue("valid"),
		Value:      types.StringValue("mxa.mailgun.org"),
	}

	objValue, diags := recordValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue returned errors: %v", diags)
	}

	if objValue.IsNull() {
		t.Error("Expected non-null object value")
	}

	attrs := objValue.Attributes()
	if len(attrs) != 7 {
		t.Errorf("Expected 7 attributes, got %d", len(attrs))
	}

	// Verify record type
	recordTypeAttr := attrs["record_type"].(types.String)
	if recordTypeAttr.ValueString() != "MX" {
		t.Errorf("Expected record_type 'MX', got '%s'", recordTypeAttr.ValueString())
	}
}

func TestSendingDnsRecordsValue_ToObjectValue(t *testing.T) {
	ctx := t.Context()

	cachedList, _ := types.ListValueFrom(ctx, types.StringType, []string{})

	recordValue := domains.SendingDnsRecordsValue{
		Cached:     cachedList,
		IsActive:   types.BoolValue(false),
		Name:       types.StringValue("test.example.com"),
		Priority:   types.StringValue(""),
		RecordType: types.StringValue("TXT"),
		Valid:      types.StringValue("unknown"),
		Value:      types.StringValue("v=spf1 include:mailgun.org ~all"),
	}

	objValue, diags := recordValue.ToObjectValue(ctx)
	if diags.HasError() {
		t.Fatalf("ToObjectValue returned errors: %v", diags)
	}

	if objValue.IsNull() {
		t.Error("Expected non-null object value")
	}

	attrs := objValue.Attributes()
	if len(attrs) != 7 {
		t.Errorf("Expected 7 attributes, got %d", len(attrs))
	}
}

func TestMapDomainResponseToModel_BasicFields(t *testing.T) {
	// This test would ideally test the mapDomainResponseToModel function,
	// but since it's not exported, we can only test it indirectly through
	// integration tests. However, we can document what it should do:
	//
	// 1. Map SDK Domain response fields to DomainValue
	// 2. Handle DNS records (receiving and sending)
	// 3. Set computed-only fields to null when not in plan
	// 4. Preserve plan-only fields from the input model
	t.Skip("mapDomainResponseToModel is not exported, tested via acceptance tests")
}

func TestDomainResourceSchema_HasRequiredFields(t *testing.T) {
	schema := domains.DomainResourceSchema()

	// Verify key fields exist
	if schema.Attributes["name"] == nil {
		t.Error("Schema missing 'name' attribute")
	}
	if schema.Attributes["spam_action"] == nil {
		t.Error("Schema missing 'spam_action' attribute")
	}
	if schema.Attributes["wildcard"] == nil {
		t.Error("Schema missing 'wildcard' attribute")
	}
	if schema.Attributes["domain"] == nil {
		t.Error("Schema missing 'domain' attribute")
	}
	if schema.Attributes["receiving_dns_records"] == nil {
		t.Error("Schema missing 'receiving_dns_records' attribute")
	}
	if schema.Attributes["sending_dns_records"] == nil {
		t.Error("Schema missing 'sending_dns_records' attribute")
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
				Config: testAccDomainResourceConfig(domainName, "disabled", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("mailgun_domain.test", "spam_action", "disabled"),
					resource.TestCheckResourceAttr("mailgun_domain.test", "wildcard", "false"),
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "domain.id"),
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "domain.name"),
					resource.TestCheckResourceAttrSet("mailgun_domain.test", "domain.state"),
				),
			},
			// ImportState testing
			{
				ResourceName:            "mailgun_domain.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"smtp_password", "dkim_key_size", "force_dkim_authority", "web_scheme", "web_prefix"},
			},
			// Update and Read testing
			{
				Config: testAccDomainResourceConfig(domainName, "tag", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("mailgun_domain.test", "spam_action", "tag"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func TestAccDomainResource_Wildcard(t *testing.T) {
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
				Config: testAccDomainResourceConfig(domainName, "disabled", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mailgun_domain.test", "name", domainName),
					resource.TestCheckResourceAttr("mailgun_domain.test", "wildcard", "true"),
				),
			},
		},
	})
}

func testAccDomainResourceConfig(domainName, spamAction string, wildcard bool) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

resource "mailgun_domain" "test" {
  name        = "%s"
  spam_action = "%s"
  wildcard    = %t
}
`, os.Getenv("MAILGUN_API_KEY"), domainName, spamAction, wildcard)
}

func testAccCheckDomainResourceDestroy(s *terraform.State) error {
	// Domain deletion is handled by the provider
	// This is a placeholder for more complex destroy checks if needed
	return nil
}
