// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package smtp_credentials

import (
	"context"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/mailgun/mailgun-go/v5"
)

func createCredential(t *testing.T, client *mailgun.Client, planOverrides, configOverrides map[string]tftypes.Value) *resource.CreateResponse {
	t.Helper()

	resourceSchema := SmtpCredentialResourceSchema()
	plan := credentialObject(t, planOverrides)
	config := credentialObject(t, configOverrides)

	resp := &resource.CreateResponse{State: tfsdk.State{Raw: plan, Schema: resourceSchema}}
	(&SmtpCredentialResource{client: client}).Create(context.Background(), resource.CreateRequest{
		Plan:   tfsdk.Plan{Raw: plan, Schema: resourceSchema},
		Config: tfsdk.Config{Raw: config, Schema: resourceSchema},
	}, resp)
	return resp
}

func createdCredentialListResponse() []byte {
	return []byte(`{"total_count":1,"items":[{"login":"user@example.com","created_at":"Mon, 02 Jan 2006 15:04:05 UTC"}]}`)
}

func TestCreate_Success_LegacyPassword(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(createdCredentialListResponse())
		}
	})

	resp := createCredential(t, client,
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
		},
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "legacy-secret"),
		},
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestCreate_Success_WriteOnlyPassword(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(createdCredentialListResponse())
		}
	})

	resp := createCredential(t, client,
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
		map[string]tftypes.Value{
			"domain":              tftypes.NewValue(tftypes.String, "example.com"),
			"login":               tftypes.NewValue(tftypes.String, "user"),
			"password_wo":         tftypes.NewValue(tftypes.String, "wo-secret"),
			"password_wo_version": tftypes.NewValue(tftypes.Number, int64(1)),
		},
	)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func TestCreate_MissingDomain(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
		map[string]tftypes.Value{
			"login":    tftypes.NewValue(tftypes.String, "user"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing domain")
	}
}

func TestCreate_MissingLogin(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
		map[string]tftypes.Value{
			"domain":   tftypes.NewValue(tftypes.String, "example.com"),
			"password": tftypes.NewValue(tftypes.String, "secret"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a missing login")
	}
}

func TestCreate_MissingPassword(t *testing.T) {
	resp := createCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}),
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
		map[string]tftypes.Value{
			"domain": tftypes.NewValue(tftypes.String, "example.com"),
			"login":  tftypes.NewValue(tftypes.String, "user"),
		},
	)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when neither password source is set")
	}
}

func TestCreate_UnconfiguredClient(t *testing.T) {
	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, nil, overrides, overrides)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the client is not configured")
	}
}

func TestCreate_CreateCredentialAPIError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, client, overrides, overrides)

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when CreateCredential fails")
	}
}

// The created_at lookup is best-effort: if the listing never catches up
// within the retry budget, Create must still succeed with created_at left
// null for Read to fill in later, not fail the apply.
func TestCreate_LookupNeverCatchesUp_LeavesCreatedAtNull(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
		}
	})

	overrides := map[string]tftypes.Value{
		"domain":   tftypes.NewValue(tftypes.String, "example.com"),
		"login":    tftypes.NewValue(tftypes.String, "user"),
		"password": tftypes.NewValue(tftypes.String, "secret"),
	}

	resp := createCredential(t, client, overrides, overrides)

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}

func importCredential(t *testing.T, client *mailgun.Client, id string) *resource.ImportStateResponse {
	t.Helper()

	resp := &resource.ImportStateResponse{
		State: tfsdk.State{Schema: SmtpCredentialResourceSchema()},
	}
	(&SmtpCredentialResource{client: client}).ImportState(context.Background(), resource.ImportStateRequest{ID: id}, resp)
	return resp
}

func TestImportState_InvalidID(t *testing.T) {
	resp := importCredential(t, testClient(t, func(http.ResponseWriter, *http.Request) {}), "not-a-valid-id")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic for a malformed import ID")
	}
}

func TestImportState_UnconfiguredClient(t *testing.T) {
	resp := importCredential(t, nil, "example.com/user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the client is not configured")
	}
}

func TestImportState_ListFailureProducesError(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"message":"boom"}`))
	})

	resp := importCredential(t, client, "example.com/user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the credential listing fails")
	}
}

func TestImportState_CredentialNotFound(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"total_count":0,"items":[]}`))
	})

	resp := importCredential(t, client, "example.com/gone-user")

	if !resp.Diagnostics.HasError() {
		t.Fatal("expected a diagnostic when the credential does not exist")
	}
}

func TestImportState_Success(t *testing.T) {
	client := testClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(createdCredentialListResponse())
	})

	resp := importCredential(t, client, "example.com/user")

	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}
}
