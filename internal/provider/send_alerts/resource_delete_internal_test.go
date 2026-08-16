// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func deleteSendAlert(t *testing.T, client *mailgun.Client, name string) *resource.DeleteResponse {
	t.Helper()

	resourceSchema := SendAlertResourceSchema()
	state := tfsdk.State{
		Raw: sendAlertStateObject(t, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, name),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.DeleteResponse{State: state}
	(&sendAlertResource{client: client}).Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

// A genuine 404 on delete means the alert is already gone: Mailgun's own
// Delete must not hard-fail the destroy, matching every other resource's
// Delete guard (`mgerr.IsNotFound(err)`).
func TestDelete_GenuineNotFoundIsIgnored(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"alert not found"}`))
	})

	resp := deleteSendAlert(t, client, "gone-alert")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics for a genuine 404 during delete: %v", resp.Diagnostics)
	}
}

// A 500 must still raise a diagnostic; only a genuine 404 is swallowed.
func TestDelete_500RaisesDiagnostic(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := deleteSendAlert(t, client, "my-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response during delete")
	}
}
