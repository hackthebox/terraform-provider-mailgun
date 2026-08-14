// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5/mtypes"
)

func TestFindEventuallyReturnsFirstSuccess(t *testing.T) {
	calls := 0
	want := &mtypes.Credential{Login: "user@example.com"}

	got, err := findEventually(context.Background(), 4, 0, func() (*mtypes.Credential, error) {
		calls++
		return want, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("credential = %v, want %v", got, want)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 (no retry once the listing succeeds)", calls)
	}
}

func TestFindEventuallyRetriesUntilTheListingCatchesUp(t *testing.T) {
	calls := 0
	want := &mtypes.Credential{Login: "user@example.com"}

	got, err := findEventually(context.Background(), 4, 0, func() (*mtypes.Credential, error) {
		calls++
		if calls < 3 {
			return nil, errors.New("credential not found")
		}
		return want, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("credential = %v, want %v", got, want)
	}
	if calls != 3 {
		t.Errorf("calls = %d, want 3", calls)
	}
}

func TestFindEventuallyGivesUpAfterAttempts(t *testing.T) {
	calls := 0
	wantErr := errors.New("credential not found")

	_, err := findEventually(context.Background(), 2, 0, func() (*mtypes.Credential, error) {
		calls++
		return nil, wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("err = %v, want %v", err, wantErr)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2 (the attempt budget, not the package default)", calls)
	}
}

func TestFindEventuallyAbortsOnCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	calls := 0
	_, err := findEventually(ctx, 4, time.Minute, func() (*mtypes.Credential, error) {
		calls++
		return nil, errors.New("credential not found")
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1 before the context cancels the wait", calls)
	}
}

func TestFindEventuallyDefaultsToFourAttempts(t *testing.T) {
	calls := 0

	_, err := findEventually(context.Background(), createdAtLookupAttempts, 0, func() (*mtypes.Credential, error) {
		calls++
		return nil, errors.New("credential not found")
	})

	if err == nil {
		t.Fatal("expected the lookup to give up")
	}
	if calls != 4 {
		t.Errorf("calls = %d, want 4: the package default budget", calls)
	}
}

func TestCreatedAtLookupActuallyBacksOff(t *testing.T) {
	if createdAtLookupDelay <= 0 {
		t.Fatalf("createdAtLookupDelay = %v; a zero delay retries the listing with no time to catch up", createdAtLookupDelay)
	}
	if createdAtLookupDelay > time.Second {
		t.Errorf("createdAtLookupDelay = %v; too long to sit in a create", createdAtLookupDelay)
	}
}

func TestCreatedAtLookupBudgetCoversItsAttempts(t *testing.T) {
	if createdAtLookupBudget <= 0 {
		t.Fatalf("createdAtLookupBudget = %v; a non-positive budget expires before the first lookup runs", createdAtLookupBudget)
	}

	// The budget has to outlast the waits it wraps, or later attempts never happen.
	if minimum := time.Duration(createdAtLookupAttempts-1) * createdAtLookupDelay; createdAtLookupBudget <= minimum {
		t.Errorf("createdAtLookupBudget = %v, must exceed the %v spent waiting between attempts", createdAtLookupBudget, minimum)
	}
}

func TestPasswordForCreate(t *testing.T) {
	tests := []struct {
		name       string
		passwordWO types.String
		legacy     types.String
		wantPass   string
		wantOK     bool
	}{
		{"write-only preferred", types.StringValue("wo-secret"), types.StringValue("legacy"), "wo-secret", true},
		{"legacy when no wo", types.StringNull(), types.StringValue("legacy"), "legacy", true},
		{"neither set", types.StringNull(), types.StringNull(), "", false},
		{"legacy unknown ignored", types.StringNull(), types.StringUnknown(), "", false},
		{"wo unknown falls back to legacy", types.StringUnknown(), types.StringValue("legacy"), "legacy", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPass, gotOK := passwordForCreate(tt.passwordWO, tt.legacy)
			if gotPass != tt.wantPass || gotOK != tt.wantOK {
				t.Errorf("passwordForCreate() = (%q, %v), want (%q, %v)", gotPass, gotOK, tt.wantPass, tt.wantOK)
			}
		})
	}
}

func TestWriteOnlyRotationRequested(t *testing.T) {
	tests := []struct {
		name  string
		plan  types.Int64
		state types.Int64
		want  bool
	}{
		{"version bumped", types.Int64Value(2), types.Int64Value(1), true},
		{"version unchanged", types.Int64Value(1), types.Int64Value(1), false},
		{"first set from null state", types.Int64Value(1), types.Int64Null(), true},
		{"no version in plan", types.Int64Null(), types.Int64Null(), false},
		{"version dropped from config", types.Int64Null(), types.Int64Value(1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := writeOnlyRotationRequested(tt.plan, tt.state); got != tt.want {
				t.Errorf("writeOnlyRotationRequested() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolveUpdatePassword(t *testing.T) {
	tests := []struct {
		name         string
		passwordWO   types.String
		planPW       types.String
		planVersion  types.Int64
		stateVersion types.Int64
		wantPW       string
		wantRotate   bool
		wantErr      string
	}{
		{
			name:         "write-only with version bump rotates",
			passwordWO:   types.StringValue("wo-secret"),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Value(2),
			stateVersion: types.Int64Value(1),
			wantPW:       "wo-secret",
			wantRotate:   true,
			wantErr:      "",
		},
		{
			name:         "write-only without version bump skips rotation",
			passwordWO:   types.StringValue("wo-secret"),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Value(1),
			stateVersion: types.Int64Value(1),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "",
		},
		{
			name:         "legacy null preserves imported state",
			passwordWO:   types.StringNull(),
			planPW:       types.StringNull(),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "",
		},
		{
			name:         "legacy empty string is an error",
			passwordWO:   types.StringNull(),
			planPW:       types.StringValue(""),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "",
			wantRotate:   false,
			wantErr:      "Invalid Password",
		},
		{
			name:         "legacy non-empty rotates",
			passwordWO:   types.StringNull(),
			planPW:       types.StringValue("newpass"),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "newpass",
			wantRotate:   true,
			wantErr:      "",
		},
		{
			name:         "write-only unknown falls through to legacy",
			passwordWO:   types.StringUnknown(),
			planPW:       types.StringValue("legacy"),
			planVersion:  types.Int64Null(),
			stateVersion: types.Int64Null(),
			wantPW:       "legacy",
			wantRotate:   true,
			wantErr:      "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPW, gotRotate, gotErr := resolveUpdatePassword(tt.passwordWO, tt.planPW, tt.planVersion, tt.stateVersion)
			if gotPW != tt.wantPW || gotRotate != tt.wantRotate || gotErr != tt.wantErr {
				t.Errorf("resolveUpdatePassword() = (%q, %v, %q), want (%q, %v, %q)",
					gotPW, gotRotate, gotErr, tt.wantPW, tt.wantRotate, tt.wantErr)
			}
		})
	}
}

// credentialObject builds a resource-shaped tftypes value, defaulting every
// attribute to null and applying the given overrides.
func credentialObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := SmtpCredentialResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("schema terraform type is not an object")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for name, attrType := range objType.AttributeTypes {
		if override, found := overrides[name]; found {
			attrs[name] = override
			continue
		}
		attrs[name] = tftypes.NewValue(attrType, nil)
	}

	return tftypes.NewValue(objType, attrs)
}

func planPassword(t *testing.T, plan tfsdk.Plan) types.String {
	t.Helper()

	var got types.String
	if diags := plan.GetAttribute(context.Background(), path.Root("password"), &got); diags.HasError() {
		t.Fatalf("reading planned password: %v", diags)
	}
	return got
}

func TestModifyPlanPinsLegacyPasswordWhenWriteOnlyConfigured(t *testing.T) {
	resourceSchema := SmtpCredentialResourceSchema()

	config := credentialObject(t, map[string]tftypes.Value{
		"password_wo":         tftypes.NewValue(tftypes.String, "wo-secret"),
		"password_wo_version": tftypes.NewValue(tftypes.Number, int64(1)),
	})
	// An Optional+Computed password carries its prior state value into the plan.
	plan := credentialObject(t, map[string]tftypes.Value{
		"password":            tftypes.NewValue(tftypes.String, "legacy-secret"),
		"password_wo_version": tftypes.NewValue(tftypes.Number, int64(1)),
	})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan, Schema: resourceSchema}}
	(&SmtpCredentialResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: config, Schema: resourceSchema},
		Plan:   tfsdk.Plan{Raw: plan, Schema: resourceSchema},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := planPassword(t, resp.Plan); !got.IsNull() {
		t.Errorf("planned password = %v, want null", got)
	}
}

func TestModifyPlanLeavesLegacyPasswordWhenWriteOnlyAbsent(t *testing.T) {
	resourceSchema := SmtpCredentialResourceSchema()

	values := credentialObject(t, map[string]tftypes.Value{
		"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
	})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: values, Schema: resourceSchema}}
	(&SmtpCredentialResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: values, Schema: resourceSchema},
		Plan:   tfsdk.Plan{Raw: values, Schema: resourceSchema},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := planPassword(t, resp.Plan); got.ValueString() != "legacy-secret" {
		t.Errorf("planned password = %v, want %q", got, "legacy-secret")
	}
}

func TestModifyPlanIgnoresDestroy(t *testing.T) {
	resourceSchema := SmtpCredentialResourceSchema()

	objType := SmtpCredentialResourceSchema().Type().TerraformType(context.Background())
	nullPlan := tftypes.NewValue(objType, nil)

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: nullPlan, Schema: resourceSchema}}
	(&SmtpCredentialResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: nullPlan, Schema: resourceSchema},
		Plan:   tfsdk.Plan{Raw: nullPlan, Schema: resourceSchema},
	}, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.Plan.Raw.IsNull() {
		t.Error("destroy plan should be left untouched")
	}
}

func TestModifyPlanStopsOnConfigReadError(t *testing.T) {
	resourceSchema := SmtpCredentialResourceSchema()

	plan := credentialObject(t, map[string]tftypes.Value{
		"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
	})

	// A schema without password_wo makes the config read fail, which must abort
	// before the plan is touched.
	strippedSchema := schema.Schema{Attributes: map[string]schema.Attribute{
		"domain": schema.StringAttribute{Required: true},
	}}
	strippedType := strippedSchema.Type().TerraformType(context.Background())
	strippedConfig := tftypes.NewValue(strippedType, map[string]tftypes.Value{
		"domain": tftypes.NewValue(tftypes.String, "example.com"),
	})

	resp := &resource.ModifyPlanResponse{Plan: tfsdk.Plan{Raw: plan, Schema: resourceSchema}}
	(&SmtpCredentialResource{}).ModifyPlan(context.Background(), resource.ModifyPlanRequest{
		Config: tfsdk.Config{Raw: strippedConfig, Schema: strippedSchema},
		Plan:   tfsdk.Plan{Raw: plan, Schema: resourceSchema},
	}, resp)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic from the failed config read")
	}
	if got := planPassword(t, resp.Plan); got.ValueString() != "legacy-secret" {
		t.Errorf("planned password = %v, want %q left untouched", got, "legacy-secret")
	}
}
