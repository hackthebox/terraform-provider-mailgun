// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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

// webhookStateObject builds a resource-shaped tftypes.Value, defaulting
// every attribute to null and applying the given overrides.
func webhookStateObject(t *testing.T, overrides map[string]tftypes.Value) tftypes.Value {
	t.Helper()

	objType, ok := WebhookResourceSchema().Type().TerraformType(context.Background()).(tftypes.Object)
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

func readWebhook(t *testing.T, client *mailgun.Client) *resource.ReadResponse {
	t.Helper()

	resourceSchema := WebhookResourceSchema()
	state := tfsdk.State{
		Raw: webhookStateObject(t, map[string]tftypes.Value{
			"domain":       tftypes.NewValue(tftypes.String, "example.com"),
			"webhook_type": tftypes.NewValue(tftypes.String, "opened"),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.ReadResponse{State: state}
	(&webhookResource{client: client}).Read(context.Background(), resource.ReadRequest{State: state}, resp)
	return resp
}

func TestRead_RemovesResourceOnGenuine404(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Webhook not found"}`))
	})

	resp := readWebhook(t, client)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state on a genuine 404")
	}
}

// The SDK reports an empty webhook object (200 OK, no url/urls) as the bare
// error `webhook '%s' returned no urls`, not as a typed 404. See
// mailgun-go/v5@v5.19.1/webhooks.go GetWebhook.
func TestRead_RemovesResourceWhenSDKReportsNoURLsSentinel(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"webhook":{}}`))
	})

	resp := readWebhook(t, client)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("expected the resource to be removed from state when the SDK reports no urls")
	}
}

// A 500 whose body happens to mention "not found" must not be mistaken for a
// 404: the string-matching this replaces would remove a live resource here.
func TestRead_500MentioningNotFoundIsNotTreatedAsMissing(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"upstream dependency not found: 404 from internal service"}`))
	})

	resp := readWebhook(t, client)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response, even one mentioning 'not found'")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a 500 must not remove the resource from state just because its body mentions 'not found'")
	}
}

func TestRead_500ProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := readWebhook(t, client)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a server error must not remove the resource from state")
	}
}

func TestRead_RateLimitedProducesErrorAndLeavesStateIntact(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"message":"Too Many Requests"}`))
	})

	resp := readWebhook(t, client)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 429 response")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a rate-limited error must not remove the resource from state")
	}
}

func TestRead_TransportErrorProducesErrorAndLeavesStateIntact(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	mg := mailgun.NewMailgun("test-key")
	if err := mg.SetAPIBase(srv.URL); err != nil {
		t.Fatalf("SetAPIBase: %v", err)
	}
	srv.Close() // connections to this address now fail outright

	resp := readWebhook(t, mg)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a transport failure")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a transport failure must not remove the resource from state")
	}
}
