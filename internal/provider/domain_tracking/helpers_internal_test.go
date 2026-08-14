// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFillUnknownsFromKeepsConfiguredValues(t *testing.T) {
	plan := DomainTrackingModel{
		ClickActive:           types.BoolValue(true),
		OpenActive:            types.BoolValue(false),
		UnsubscribeActive:     types.BoolValue(true),
		UnsubscribeHtmlFooter: types.StringValue("<p>configured</p>"),
		UnsubscribeTextFooter: types.StringValue("configured"),
	}
	actual := DomainTrackingModel{
		ClickActive:           types.BoolValue(false),
		OpenActive:            types.BoolValue(true),
		UnsubscribeActive:     types.BoolValue(false),
		UnsubscribeHtmlFooter: types.StringNull(),
		UnsubscribeTextFooter: types.StringNull(),
	}

	fillUnknownsFrom(&plan, &actual)

	if !plan.UnsubscribeActive.ValueBool() {
		t.Error("unsubscribe_active was configured true; a disagreeing read must not overwrite it")
	}
	if !plan.ClickActive.ValueBool() {
		t.Error("click_active was configured true; a disagreeing read must not overwrite it")
	}
	if plan.OpenActive.ValueBool() {
		t.Error("open_active was configured false; a disagreeing read must not overwrite it")
	}
	if plan.UnsubscribeHtmlFooter.ValueString() != "<p>configured</p>" {
		t.Errorf("html footer = %v, want the configured value preserved", plan.UnsubscribeHtmlFooter)
	}
	if plan.UnsubscribeTextFooter.ValueString() != "configured" {
		t.Errorf("text footer = %v, want the configured value preserved", plan.UnsubscribeTextFooter)
	}
}

func TestFillUnknownsFromResolvesUnknownValues(t *testing.T) {
	plan := DomainTrackingModel{
		ClickActive:           types.BoolUnknown(),
		OpenActive:            types.BoolUnknown(),
		UnsubscribeActive:     types.BoolUnknown(),
		UnsubscribeHtmlFooter: types.StringUnknown(),
		UnsubscribeTextFooter: types.StringUnknown(),
	}
	actual := DomainTrackingModel{
		ClickActive:           types.BoolValue(true),
		OpenActive:            types.BoolValue(true),
		UnsubscribeActive:     types.BoolValue(true),
		UnsubscribeHtmlFooter: types.StringValue("<p>from api</p>"),
		UnsubscribeTextFooter: types.StringNull(),
	}

	fillUnknownsFrom(&plan, &actual)

	if !plan.ClickActive.ValueBool() || !plan.OpenActive.ValueBool() || !plan.UnsubscribeActive.ValueBool() {
		t.Error("unknown booleans should be resolved from the API read")
	}
	if plan.UnsubscribeHtmlFooter.ValueString() != "<p>from api</p>" {
		t.Errorf("html footer = %v, want the API value", plan.UnsubscribeHtmlFooter)
	}
	if !plan.UnsubscribeTextFooter.IsNull() {
		t.Errorf("text footer = %v, want null from the API read", plan.UnsubscribeTextFooter)
	}
}
