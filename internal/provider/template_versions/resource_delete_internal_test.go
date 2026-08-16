// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package template_versions

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func deleteTemplateVersion(t *testing.T, client *mailgun.Client, tag string) *resource.DeleteResponse {
	t.Helper()

	resourceSchema := TemplateVersionResourceSchema()
	state := tfsdk.State{
		Raw: templateVersionStateObject(t, map[string]tftypes.Value{
			"domain":        tftypes.NewValue(tftypes.String, "example.com"),
			"template_name": tftypes.NewValue(tftypes.String, "my-template"),
			"tag":           tftypes.NewValue(tftypes.String, tag),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.DeleteResponse{State: state}
	(&templateVersionResource{client: client}).Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

func TestDelete_GenuineNotFoundIsIgnored(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Template version not found"}`))
	})

	resp := deleteTemplateVersion(t, client, "gone-tag")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for a genuine 404 during delete: %v", resp.Diagnostics)
	}
}

// A 500 whose body happens to contain "not found"/"404" must raise a
// diagnostic instead of being silently ignored as "already deleted".
func TestDelete_500MentioningNotFoundRaisesDiagnostic(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream dependency not found: 404 from internal service"}`))
	})

	resp := deleteTemplateVersion(t, client, "v1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response mentioning 'not found'/'404'")
	}
}

// The "cannot delete an active version" exception is a genuine 200/4xx
// business rule, distinct from "not found", and must survive the switch to
// mgerr.IsNotFound with its own dedicated diagnostic.
func TestDelete_ActiveVersionRaisesDedicatedDiagnostic(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"deleting active version is not allowed"}`))
	})

	resp := deleteTemplateVersion(t, client, "v1")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when deleting the active version")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Cannot Delete Active Template Version" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Cannot Delete Active Template Version")
	}
}
