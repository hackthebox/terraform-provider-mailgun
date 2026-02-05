// Copyright (c) Hack The Box
// SPDX-License-Identifier: MPL-2.0

package send_alerts

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/mailgun/mailgun-go/v5"
)

const (
	sendAlertsEndpoint = "/v1/thresholds/alerts/send"
)

// SendAlertsAPIClient wraps the Mailgun client to provide Send Alerts API functionality.
type SendAlertsAPIClient struct {
	client *mailgun.Client
}

// NewSendAlertsAPIClient creates a new Send Alerts API client.
func NewSendAlertsAPIClient(client *mailgun.Client) *SendAlertsAPIClient {
	return &SendAlertsAPIClient{client: client}
}

// ListSendAlerts retrieves all send alerts for the account.
func (c *SendAlertsAPIClient) ListSendAlerts(ctx context.Context) (*SendAlertsListAPIResponse, error) {
	url := c.client.APIBase() + sendAlertsEndpoint

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.client.APIKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var msgResp MessageResponse
		if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, msgResp.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SendAlertsListAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// GetSendAlert retrieves a single send alert by name.
func (c *SendAlertsAPIClient) GetSendAlert(ctx context.Context, name string) (*SendAlertAPIResponse, error) {
	url := c.client.APIBase() + sendAlertsEndpoint + "/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.client.APIKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // Alert not found
	}

	if resp.StatusCode != http.StatusOK {
		var msgResp MessageResponse
		if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, msgResp.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SendAlertAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// CreateSendAlert creates a new send alert.
func (c *SendAlertsAPIClient) CreateSendAlert(ctx context.Context, alert SendAlertAPIRequest) (*SendAlertAPIResponse, error) {
	url := c.client.APIBase() + sendAlertsEndpoint

	jsonBody, err := json.Marshal(alert)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.client.APIKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTPClient().Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var msgResp MessageResponse
		if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, msgResp.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var result SendAlertAPIResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &result, nil
}

// UpdateSendAlert updates an existing send alert.
func (c *SendAlertsAPIClient) UpdateSendAlert(ctx context.Context, name string, alert SendAlertAPIRequest) error {
	url := c.client.APIBase() + sendAlertsEndpoint + "/" + name

	jsonBody, err := json.Marshal(alert)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.client.APIKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var msgResp MessageResponse
		if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, msgResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// DeleteSendAlert deletes a send alert by name.
func (c *SendAlertsAPIClient) DeleteSendAlert(ctx context.Context, name string) error {
	url := c.client.APIBase() + sendAlertsEndpoint + "/" + name

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.SetBasicAuth("api", c.client.APIKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.HTTPClient().Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		var msgResp MessageResponse
		if json.Unmarshal(body, &msgResp) == nil && msgResp.Message != "" {
			return fmt.Errorf("API error (status %d): %s", resp.StatusCode, msgResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}
