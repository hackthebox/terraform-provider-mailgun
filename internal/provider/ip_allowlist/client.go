// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package ip_allowlist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/mailgun/mailgun-go/v5"

	"github.com/hackthebox/terraform-provider-mailgun/internal/provider/mgerr"
)

// IPAllowlistEntry represents a single IP allowlist entry from the Mailgun API.
type IPAllowlistEntry struct {
	IPAddress   string `json:"ip_address"`
	Description string `json:"description"`
}

// IPAllowlistResponse represents the response from the IP allowlist API.
type IPAllowlistResponse struct {
	Addresses []IPAllowlistEntry `json:"addresses"`
	Message   string             `json:"message,omitempty"`
}

// IPAllowlistClient provides methods to interact with the Mailgun IP allowlist API.
type IPAllowlistClient struct {
	mg *mailgun.Client
}

// NewIPAllowlistClient creates a new IP allowlist client from an existing Mailgun client.
func NewIPAllowlistClient(mg *mailgun.Client) *IPAllowlistClient {
	return &IPAllowlistClient{mg: mg}
}

// getBaseURL returns the API base URL for v2 endpoints.
func (c *IPAllowlistClient) getBaseURL() string {
	// The SDK's APIBase returns something like "https://api.mailgun.net/v3"
	// We need to use v2 for the ip_whitelist endpoint
	base := c.mg.APIBase()
	// Replace /v3 with /v2, or append /v2 if no version present
	if strings.HasSuffix(base, "/v3") {
		return strings.TrimSuffix(base, "/v3") + "/v2"
	}
	if strings.HasSuffix(base, "/v4") {
		return strings.TrimSuffix(base, "/v4") + "/v2"
	}
	return base + "/v2"
}

// ListIPAllowlist retrieves all IP allowlist entries.
func (c *IPAllowlistClient) ListIPAllowlist(ctx context.Context) ([]IPAllowlistEntry, error) {
	url := c.getBaseURL() + "/ip_whitelist"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.mg.APIKey())

	resp, err := c.mg.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp IPAllowlistResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result IPAllowlistResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Addresses, nil
}

// GetIPAllowlistEntry retrieves a specific IP allowlist entry by address.
// There is no per-entry endpoint: it lists all entries (a single 200
// response) and scans for a match. found is false on a scan miss; err is
// only non-nil for a genuine failure of the underlying list request, and
// must never be treated as "not found" by the caller.
func (c *IPAllowlistClient) GetIPAllowlistEntry(ctx context.Context, address string) (*IPAllowlistEntry, bool, error) {
	entries, err := c.ListIPAllowlist(ctx)
	if err != nil {
		return nil, false, err
	}

	for _, entry := range entries {
		if entry.IPAddress == address {
			return &entry, true, nil
		}
	}

	return nil, false, nil
}

// CreateIPAllowlistEntry adds a new IP allowlist entry.
func (c *IPAllowlistClient) CreateIPAllowlistEntry(ctx context.Context, address, description string) error {
	apiURL := c.getBaseURL() + "/ip_whitelist"

	form := url.Values{}
	form.Set("address", address)
	if description != "" {
		form.Set("description", description)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.mg.APIKey())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.mg.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp IPAllowlistResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// UpdateIPAllowlistEntry updates an existing IP allowlist entry's description.
func (c *IPAllowlistClient) UpdateIPAllowlistEntry(ctx context.Context, address, description string) error {
	apiURL := c.getBaseURL() + "/ip_whitelist"

	form := url.Values{}
	form.Set("address", address)
	form.Set("description", description)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, apiURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.mg.APIKey())
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.mg.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp IPAllowlistResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, errResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteIPAllowlistEntry removes an IP allowlist entry. Its error, unlike
// the other methods on this client, is wrapped in mgerr.StatusError: the
// caller ignores a genuine 404 as "already gone" and must be able to tell
// that apart from a 500 whose body happens to mention "not found" or "404".
func (c *IPAllowlistClient) DeleteIPAllowlistEntry(ctx context.Context, address string) error {
	apiURL := c.getBaseURL() + "/ip_whitelist?address=" + url.QueryEscape(address)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, apiURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.mg.APIKey())

	resp, err := c.mg.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp IPAllowlistResponse
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Message != "" {
			return mgerr.StatusError(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, errResp.Message), resp.StatusCode)
		}
		return mgerr.StatusError(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)), resp.StatusCode)
	}

	return nil
}
