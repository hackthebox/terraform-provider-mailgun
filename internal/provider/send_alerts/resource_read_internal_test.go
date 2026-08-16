// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func testClient(t *testing.T, handler http.HandlerFunc) *mailgun.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	return mg
}

// sendAlertStateObject builds a resource-shaped tftypes.Value, defaulting
// every attribute to null and applying the given overrides.
func sendAlertStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := SendAlertResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
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

func readSendAlert(t *testing.T, client *mailgun.Client, name string) *resource.ReadResponse {
	t.Helper()

	resourceSchema := SendAlertResourceSchema()
	state := tfsdk.State{
		Raw: sendAlertStateObject(t, map[string]tftypes.Value{
			"name": tftypes.NewValue(tftypes.String, name),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&sendAlertResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func TestRead_RemovesResourceOnGenuine404(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"alert not found"}`))
	})

	resp := readSendAlert(t, client, "gone-alert")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state on a genuine 404")
	}
}

func TestRead_500ProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readSendAlert(t, client, "my-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a server error must not remove the resource from state")
	}
}

func updateSendAlert(t *testing.T, client *mailgun.Client, name string) *resource.UpdateResponse {
	t.Helper()

	resourceSchema := SendAlertResourceSchema()
	val := sendAlertStateObject(t, map[string]tftypes.Value{
		"name":       tftypes.NewValue(tftypes.String, name),
		"metric":     tftypes.NewValue(tftypes.String, "delivered_rate"),
		"comparator": tftypes.NewValue(tftypes.String, ">"),
		"limit":      tftypes.NewValue(tftypes.String, "0.05"),
		"dimension":  tftypes.NewValue(tftypes.String, "domain"),
	})
	plan := tfsdk.Plan{Raw: val, Schema: resourceSchema}
	state := tfsdk.State{Raw: val, Schema: resourceSchema}

	resp := &resource.UpdateResponse{State: state}
	(&sendAlertResource{client: client}).Update(context.Background(), resource.UpdateRequest{Plan: plan, State: state}, resp)
	return resp
}

// The Update flow first PUTs the change, then GETs the alert back to
// populate computed fields. A 404 on that read-back must raise a
// diagnostic, not silently map the plan as partial state (GetSendAlert no
// longer has a (nil, nil) "not found" outcome to fall back on).
func TestUpdate_ReadBack404RaisesDiagnostic(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"message":"alert not found"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	})

	resp := updateSendAlert(t, client, "my-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the post-update read-back 404s")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Error Reading Updated Mailgun Send Alert" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Error Reading Updated Mailgun Send Alert")
	}
}

func importSendAlert(t *testing.T, client *mailgun.Client, name string) *resource.ImportStateResponse {
	t.Helper()

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: SendAlertResourceSchema()},
	}
	(&sendAlertResource{client: client}).ImportState(context.Background(), resource.ImportStateRequest{ID: name}, resp)
	return resp
}

func TestImportState_NotFoundKeepsSendAlertNotFoundSummary(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"alert not found"}`))
	})

	resp := importSendAlert(t, client, "gone-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing alert")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Send Alert Not Found" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Send Alert Not Found")
	}
}

func TestImportState_500ProducesGenericErrorSummary(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := importSendAlert(t, client, "my-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Error Importing Mailgun Send Alert" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Error Importing Mailgun Send Alert")
	}
}

func readSendAlertDataSource(t *testing.T, client *mailgun.Client, name string) *datasource.ReadResponse {
	t.Helper()

	dsSchema := SendAlertDataSourceSchema()
	objType, ok := dsSchema.Type().TerraformType(context.Background()).(tftypes.Object)
	if !ok {
		t.Fatal("schema terraform type is not an object")
	}

	attrs := make(map[string]tftypes.Value, len(objType.AttributeTypes))
	for attrName, attrType := range objType.AttributeTypes {
		if attrName == "name" {
			attrs[attrName] = tftypes.NewValue(attrType, name)
			continue
		}
		attrs[attrName] = tftypes.NewValue(attrType, nil)
	}

	config := tfsdk.Config{
		Raw:    tftypes.NewValue(objType, attrs),
		Schema: dsSchema,
	}

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: dsSchema}}
	(&sendAlertDataSource{client: client}).Read(context.Background(), datasource.ReadRequest{Config: config}, resp)
	return resp
}

func TestDataSourceRead_NotFoundKeepsSendAlertNotFoundSummary(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"alert not found"}`))
	})

	resp := readSendAlertDataSource(t, client, "gone-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing alert")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Send Alert Not Found" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Send Alert Not Found")
	}
}

func TestDataSourceRead_500ProducesGenericErrorSummary(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readSendAlertDataSource(t, client, "my-alert")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
	if got := resp.Diagnostics[0].Summary(); got != "Error Reading Mailgun Send Alert" {
		t.Errorf("diagnostic summary = %q, want %q", got, "Error Reading Mailgun Send Alert")
	}
}
