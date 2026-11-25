// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_sending_keys

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func keyObject(t *testing.T, deletionDelay int64) tftypes.Value {
	t.Helper()
	return readStateObject(t, map[string]tftypes.Value{
		"id":             tftypes.NewValue(tftypes.String, "key-1"),
		"domain":         tftypes.NewValue(tftypes.String, "example.com"),
		"secret":         tftypes.NewValue(tftypes.String, "s3cret"),
		"deletion_delay": tftypes.NewValue(tftypes.Number, deletionDelay),
	})
}

func deleteKey(t *testing.T, r *domainSendingKeyResource, deletionDelay int64) *resource.DeleteResponse {
	t.Helper()
	state := tfsdk.State{Raw: keyObject(t, deletionDelay), Schema: DomainSendingKeyResourceSchema()}
	resp := &resource.DeleteResponse{State: state}
	r.Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

func TestDelete_WaitsDeletionDelayBeforeDeletingKey(t *testing.T) {
	var events []string
	client := testClient(t, func(w http.ResponseWriter, req *http.Request) {
		events = append(events, req.Method+" "+req.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":"deleted"}`))
	})
	r := &domainSendingKeyResource{
		client: client,
		wait: func(_ context.Context, d time.Duration) error {
			events = append(events, "wait "+d.String())
			return nil
		},
	}

	resp := deleteKey(t, r, 120)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	want := []string{"wait 2m0s", "DELETE /v1/keys/key-1"}
	if !slices.Equal(events, want) {
		t.Errorf("events = %v, want %v", events, want)
	}
}

// An interrupted wait (Ctrl-C, a cancelled CI run) must not delete the key and
// must fail the delete, so Terraform keeps the object and retries on the next apply.
func TestDelete_InterruptedWaitKeepsKey(t *testing.T) {
	deleted := false
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		deleted = true
	})
	r := &domainSendingKeyResource{
		client: client,
		wait: func(context.Context, time.Duration) error {
			return context.Canceled
		},
	}

	resp := deleteKey(t, r, 120)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the wait is interrupted")
	}
	if got := resp.Diagnostics[0].Detail(); !strings.Contains(got, "key-1") || !strings.Contains(got, context.Canceled.Error()) {
		t.Errorf("diagnostic detail = %q, want the key id and the cause", got)
	}
	if deleted {
		t.Error("the key must not be deleted when the wait is interrupted")
	}
}

func TestNewDomainSendingKeyResource_WaitsWithWaitFor(t *testing.T) {
	r, ok := NewDomainSendingKeyResource().(*domainSendingKeyResource)
	if !ok || r.wait == nil {
		t.Fatal("the resource must be built with a wait function")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := r.wait(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("wait on a cancelled context = %v, want context.Canceled", err)
	}
}

func TestWaitFor_ReturnsNilAfterTheDelay(t *testing.T) {
	if err := waitFor(context.Background(), time.Millisecond); err != nil {
		t.Errorf("waitFor = %v, want nil", err)
	}
}

func TestWaitFor_ReturnsContextErrorWhenCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := waitFor(ctx, time.Hour); !errors.Is(err, context.Canceled) {
		t.Errorf("waitFor = %v, want context.Canceled", err)
	}
}

func TestUpdate_StoresPlannedDeletionDelay(t *testing.T) {
	resourceSchema := DomainSendingKeyResourceSchema()
	req := resource.UpdateRequest{
		Plan:  tfsdk.Plan{Raw: keyObject(t, 120), Schema: resourceSchema},
		State: tfsdk.State{Raw: keyObject(t, 0), Schema: resourceSchema},
	}
	resp := &resource.UpdateResponse{State: tfsdk.State{Raw: keyObject(t, 0), Schema: resourceSchema}}

	(&domainSendingKeyResource{}).Update(context.Background(), req, resp)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
	var got DomainSendingKeyModel
	resp.State.Get(context.Background(), &got)
	if got.DeletionDelay.ValueInt64() != 120 {
		t.Errorf("deletion_delay = %d, want 120", got.DeletionDelay.ValueInt64())
	}
	if got.Secret.ValueString() != "s3cret" {
		t.Errorf("secret = %q, want it carried over from the plan", got.Secret.ValueString())
	}
}
