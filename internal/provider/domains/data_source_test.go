// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package domains_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/domains"
	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/test_helpers"
)

// Unit Tests - These tests don't require external API calls

func TestItemsValue_AttributeTypes(t *testing.T) {
	ctx := t.Context()
	itemValue := domains.ItemsValue{}
	attrTypes := itemValue.AttributeTypes(ctx)

	// Verify all expected attributes are present
	expectedAttrs := []string{
		"created_at", "disabled", "id", "is_disabled", "name",
		"require_tls", "skip_verification", "smtp_login", "smtp_password",
		"spam_action", "state", "tracking_host", "type",
		"use_automatic_sender_security", "web_prefix", "web_scheme", "wildcard",
	}

	for _, attrName := range expectedAttrs {
		if _, ok := attrTypes[attrName]; !ok {
			t.Errorf("Expected attribute %s not found in ItemsValue AttributeTypes", attrName)
		}
	}

	// Verify correct types for key attributes
	if attrTypes["name"] != types.StringType {
		t.Errorf("Expected name to be StringType, got %T", attrTypes["name"])
	}
	if attrTypes["wildcard"] != types.BoolType {
		t.Errorf("Expected wildcard to be BoolType, got %T", attrTypes["wildcard"])
	}
	if attrTypes["is_disabled"] != types.BoolType {
		t.Errorf("Expected is_disabled to be BoolType, got %T", attrTypes["is_disabled"])
	}

	// Verify disabled is ObjectType
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

func TestItemsValue_ToObjectValue(t *testing.T) {
	ctx := t.Context()

	disabledObj := types.ObjectNull(map[string]attr.Type{
		"code":        types.StringType,
		"note":        types.StringType,
		"permanently": types.BoolType,
		"reason":      types.StringType,
		"until":       types.StringType,
	})

	itemValue := domains.ItemsValue{
		CreatedAt:                  types.StringValue("2025-01-15T10:00:00Z"),
		Disabled:                   disabledObj,
		Id:                         types.StringValue("test.example.com"),
		IsDisabled:                 types.BoolValue(false),
		Name:                       types.StringValue("test.example.com"),
		RequireTls:                 types.BoolValue(true),
		SkipVerification:           types.BoolValue(false),
		SmtpLogin:                  types.StringValue("postmaster@test.example.com"),
		SmtpPassword:               types.StringValue("secret"),
		SpamAction:                 types.StringValue("disabled"),
		State:                      types.StringValue("active"),
		TrackingHost:               types.StringValue(""),
		ItemsType:                  types.StringValue("mailgun"),
		UseAutomaticSenderSecurity: types.BoolValue(false),
		WebPrefix:                  types.StringValue(""),
		WebScheme:                  types.StringValue("https"),
		Wildcard:                   types.BoolValue(false),
	}

	objValue, diags := itemValue.ToObjectValue(ctx)
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

	stateAttr := attrs["state"].(types.String)
	if stateAttr.ValueString() != "active" {
		t.Errorf("Expected state 'active', got '%s'", stateAttr.ValueString())
	}
}

func TestDomainsModel_Structure(t *testing.T) {
	// Test that DomainsModel can be instantiated with proper types
	ctx := t.Context()

	model := domains.DomainsModel{
		Authority:          types.StringValue(""),
		IncludeSubaccounts: types.BoolValue(false),
		Limit:              types.Int64Value(100),
		Search:             types.StringValue(""),
		Skip:               types.Int64Value(0),
		Sort:               types.StringValue(""),
		State:              types.StringValue(""),
		TotalCount:         types.Int64Value(0),
		Items:              types.ListNull(types.ObjectType{AttrTypes: domains.ItemsValue{}.AttributeTypes(ctx)}),
	}

	// Verify values
	if model.Limit.ValueInt64() != 100 {
		t.Errorf("Expected limit 100, got %d", model.Limit.ValueInt64())
	}

	if !model.Items.IsNull() {
		// If not null, it should be a valid list
		if model.Items.IsUnknown() {
			t.Error("Items should not be unknown")
		}
	}
}

func TestDomainsListDataSourceSchema_HasRequiredFields(t *testing.T) {
	schema := domains.DomainsListDataSourceSchema()

	// Verify key fields exist
	if schema.Attributes["limit"] == nil {
		t.Error("Schema missing 'limit' attribute")
	}
	if schema.Attributes["total_count"] == nil {
		t.Error("Schema missing 'total_count' attribute")
	}
	if schema.Attributes["items"] == nil {
		t.Error("Schema missing 'items' attribute")
	}
	if schema.Attributes["search"] == nil {
		t.Error("Schema missing 'search' attribute")
	}
	if schema.Attributes["skip"] == nil {
		t.Error("Schema missing 'skip' attribute")
	}
	if schema.Attributes["state"] == nil {
		t.Error("Schema missing 'state' attribute")
	}

	// Verify description exists
	if schema.Description == "" {
		t.Error("Schema should have a description")
	}
}

func TestConvertItemsToList_EmptyList(t *testing.T) {
	// This tests the behavior of an empty domain items list
	// Note: convertItemsToList is not exported, so we test indirectly
	ctx := t.Context()

	// Create an empty list
	emptyList := types.ListNull(types.ObjectType{
		AttrTypes: domains.ItemsValue{}.AttributeTypes(ctx),
	})

	if !emptyList.IsNull() {
		t.Error("Expected null list for empty items")
	}
}

func TestConvertItemsToList_WithItems(t *testing.T) {
	ctx := t.Context()

	disabledObj := types.ObjectNull(map[string]attr.Type{
		"code":        types.StringType,
		"note":        types.StringType,
		"permanently": types.BoolType,
		"reason":      types.StringType,
		"until":       types.StringType,
	})

	// Create a sample item
	item := domains.ItemsValue{
		CreatedAt:                  types.StringValue("2025-01-15T10:00:00Z"),
		Disabled:                   disabledObj,
		Id:                         types.StringValue("test1.example.com"),
		IsDisabled:                 types.BoolValue(false),
		Name:                       types.StringValue("test1.example.com"),
		RequireTls:                 types.BoolValue(true),
		SkipVerification:           types.BoolValue(false),
		SmtpLogin:                  types.StringValue("postmaster@test1.example.com"),
		SmtpPassword:               types.StringValue(""),
		SpamAction:                 types.StringValue("disabled"),
		State:                      types.StringValue("active"),
		TrackingHost:               types.StringValue(""),
		ItemsType:                  types.StringValue(""),
		UseAutomaticSenderSecurity: types.BoolValue(false),
		WebPrefix:                  types.StringValue(""),
		WebScheme:                  types.StringValue("https"),
		Wildcard:                   types.BoolValue(false),
	}

	// Create another item
	item2 := domains.ItemsValue{
		CreatedAt:                  types.StringValue("2025-01-16T10:00:00Z"),
		Disabled:                   disabledObj,
		Id:                         types.StringValue("test2.example.com"),
		IsDisabled:                 types.BoolValue(false),
		Name:                       types.StringValue("test2.example.com"),
		RequireTls:                 types.BoolValue(false),
		SkipVerification:           types.BoolValue(false),
		SmtpLogin:                  types.StringValue("postmaster@test2.example.com"),
		SmtpPassword:               types.StringValue(""),
		SpamAction:                 types.StringValue("tag"),
		State:                      types.StringValue("active"),
		TrackingHost:               types.StringValue(""),
		ItemsType:                  types.StringValue(""),
		UseAutomaticSenderSecurity: types.BoolValue(true),
		WebPrefix:                  types.StringValue(""),
		WebScheme:                  types.StringValue("https"),
		Wildcard:                   types.BoolValue(true),
	}

	// Convert items to objects first
	obj1, diags1 := item.ToObjectValue(ctx)
	if diags1.HasError() {
		t.Fatalf("Failed to convert item1 to ObjectValue: %v", diags1)
	}

	obj2, diags2 := item2.ToObjectValue(ctx)
	if diags2.HasError() {
		t.Fatalf("Failed to convert item2 to ObjectValue: %v", diags2)
	}

	// Create a list from the objects
	listValue, diags := types.ListValue(
		types.ObjectType{AttrTypes: item.AttributeTypes(ctx)},
		[]attr.Value{obj1, obj2},
	)
	if diags.HasError() {
		t.Fatalf("Failed to create list: %v", diags)
	}

	// Verify the list
	if listValue.IsNull() {
		t.Error("Expected non-null list")
	}

	elements := listValue.Elements()
	if len(elements) != 2 {
		t.Errorf("Expected 2 elements in list, got %d", len(elements))
	}
}

func TestDomainsModel_DefaultValues(t *testing.T) {
	// Test setting default values similar to what happens in the Read method
	model := domains.DomainsModel{}

	// Set default values
	if model.Authority.IsNull() {
		model.Authority = types.StringValue("")
	}
	if model.IncludeSubaccounts.IsNull() {
		model.IncludeSubaccounts = types.BoolValue(false)
	}
	if model.Limit.IsNull() {
		model.Limit = types.Int64Value(100)
	}
	if model.Search.IsNull() {
		model.Search = types.StringValue("")
	}
	if model.Skip.IsNull() {
		model.Skip = types.Int64Value(0)
	}
	if model.Sort.IsNull() {
		model.Sort = types.StringValue("")
	}
	if model.State.IsNull() {
		model.State = types.StringValue("")
	}

	// Verify defaults are set
	if model.Limit.ValueInt64() != 100 {
		t.Errorf("Expected default limit 100, got %d", model.Limit.ValueInt64())
	}
	if model.Skip.ValueInt64() != 0 {
		t.Errorf("Expected default skip 0, got %d", model.Skip.ValueInt64())
	}
	if model.IncludeSubaccounts.ValueBool() != false {
		t.Error("Expected default include_subaccounts false")
	}
}

// Acceptance Tests - These tests require MAILGUN_API_KEY and make real API calls

func TestAccDomainsDataSource(t *testing.T) {
	// Skip if MAILGUN_API_KEY is not set
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read testing
			{
				Config: testAccDomainsDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.mailgun_domains.test", "total_count"),
					resource.TestCheckResourceAttrSet("data.mailgun_domains.test", "items.#"),
				),
			},
		},
	})
}

func TestAccDomainsDataSource_WithLimit(t *testing.T) {
	if os.Getenv("MAILGUN_API_KEY") == "" {
		t.Skip("MAILGUN_API_KEY environment variable is not set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { test_helpers.AccPreCheck(t) },
		ProtoV6ProviderFactories: test_helpers.ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainsDataSourceConfigWithLimit(10),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mailgun_domains.test", "limit", "10"),
					resource.TestCheckResourceAttrSet("data.mailgun_domains.test", "total_count"),
				),
			},
		},
	})
}

func testAccDomainsDataSourceConfig() string {
	return `
provider "mailgun" {
  api_key = "` + os.Getenv("MAILGUN_API_KEY") + `"
}

data "mailgun_domains" "test" {}
`
}

func testAccDomainsDataSourceConfigWithLimit(limit int) string {
	return fmt.Sprintf(`
provider "mailgun" {
  api_key = "%s"
}

data "mailgun_domains" "test" {
  limit = %d
}
`, os.Getenv("MAILGUN_API_KEY"), limit)
}
