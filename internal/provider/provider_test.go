// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func configureWithAPIKey(t *testing.T, apiKey tftypes.Value) *provider.ConfigureResponse {
	t.Helper()

	ctx := context.Background()
	p := New("test")()

	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, provider.SchemaRequest{}, schemaResp)

	objectType := tftypes.Object{AttributeTypes: map[string]tftypes.Type{
		"api_key":  tftypes.String,
		"region":   tftypes.String,
		"endpoint": tftypes.String,
	}}

	config := tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw: tftypes.NewValue(objectType, map[string]tftypes.Value{
			"api_key":  apiKey,
			"region":   tftypes.NewValue(tftypes.String, nil),
			"endpoint": tftypes.NewValue(tftypes.String, nil),
		}),
	}

	resp := &provider.ConfigureResponse{}
	p.Configure(ctx, provider.ConfigureRequest{Config: config}, resp)

	return resp
}

func configuredAPIKey(t *testing.T, resp *provider.ConfigureResponse) string {
	t.Helper()

	client, ok := resp.ResourceData.(*mailgun.Client)
	if !ok {
		t.Fatalf("expected a *mailgun.Client, got %T", resp.ResourceData)
	}

	return client.APIKey()
}

func TestConfigureUsesAPIKeyFromConfiguration(t *testing.T) {
	t.Setenv("MAILGUN_API_KEY", "key-from-environment")

	resp := configureWithAPIKey(t, tftypes.NewValue(tftypes.String, "key-from-config"))

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := configuredAPIKey(t, resp); got != "key-from-config" {
		t.Errorf("expected the configured key to win, got %q", got)
	}
}

func TestConfigureFallsBackToEnvironmentAPIKey(t *testing.T) {
	t.Setenv("MAILGUN_API_KEY", "key-from-environment")

	resp := configureWithAPIKey(t, tftypes.NewValue(tftypes.String, nil))

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if got := configuredAPIKey(t, resp); got != "key-from-environment" {
		t.Errorf("expected the environment key, got %q", got)
	}
}

func TestConfigureErrorsWhenAPIKeyIsMissingEverywhere(t *testing.T) {
	t.Setenv("MAILGUN_API_KEY", "")

	resp := configureWithAPIKey(t, tftypes.NewValue(tftypes.String, nil))

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error when neither configuration nor environment supplies a key")
	}
	if resp.ResourceData != nil {
		t.Errorf("expected no client, got %T", resp.ResourceData)
	}
}

func TestConfigureErrorsOnUnknownAPIKey(t *testing.T) {
	t.Setenv("MAILGUN_API_KEY", "key-from-environment")

	resp := configureWithAPIKey(t, tftypes.NewValue(tftypes.String, tftypes.UnknownValue))

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected an error rather than a silent fall back to the environment key")
	}
	if resp.ResourceData != nil {
		t.Errorf("expected no client, got %T", resp.ResourceData)
	}
}

func TestSchemaMakesAPIKeyOptional(t *testing.T) {
	resp := &provider.SchemaResponse{}
	New("test")().Schema(context.Background(), provider.SchemaRequest{}, resp)

	apiKey, ok := resp.Schema.Attributes["api_key"]
	if !ok {
		t.Fatal("expected an api_key attribute")
	}
	if apiKey.IsRequired() {
		t.Error("api_key must not be required, it falls back to MAILGUN_API_KEY")
	}
	if !apiKey.IsOptional() {
		t.Error("api_key must be optional so the provider block can omit it")
	}
	if !apiKey.IsSensitive() {
		t.Error("api_key must stay sensitive")
	}
}
