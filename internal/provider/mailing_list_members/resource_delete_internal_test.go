// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package mailing_list_members

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func deleteMember(t *testing.T, client *mailgun.Client, memberAddress string) *resource.DeleteResponse {
	t.Helper()

	resourceSchema := MailingListMemberResourceSchema()
	state := tfsdk.State{
		Raw: memberStateObject(t, map[string]tftypes.Value{
			"list_address":   tftypes.NewValue(tftypes.String, "list@example.com"),
			"member_address": tftypes.NewValue(tftypes.String, memberAddress),
		}),
		Schema: resourceSchema,
	}

	resp := &resource.DeleteResponse{State: state}
	(&mailingListMemberResource{client: client}).Delete(context.Background(), resource.DeleteRequest{State: state}, resp)
	return resp
}

func TestDelete_GenuineNotFoundIsIgnored(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"Member not found"}`))
	})

	resp := deleteMember(t, client, "gone@example.com")

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

	resp := deleteMember(t, client, "member@example.com")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a 500 response mentioning 'not found'/'404'")
	}
}
