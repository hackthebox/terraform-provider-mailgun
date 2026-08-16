// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func deleteWebhook(t *testing.T, client *mailgun.Client) *resource.DeleteResponse {
	t.Helper()

	resourceSchema := WebhookResourceSchema()
	state := tfsdk.State{
		Raw: webhookStateObject(t, map[string]tftypes.Value{
			"domain":       tftypes.NewValue(tftypes.String, "example.com"),
			"webhook_type": tftypes.NewValue(tftypes.String, "opened"),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.DeleteResponse{State: state}
	(&webhookResource{client: client}).Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

func TestDelete_GenuineNotFoundIsIgnored(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Webhook not found"}`))
	})

	resp := deleteWebhook(t, client)

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

	resp := deleteWebhook(t, client)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response mentioning 'not found'/'404'")
	}
}
