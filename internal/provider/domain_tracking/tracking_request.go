// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package domain_tracking

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// putTrackingSetting updates one tracking sub-resource as multipart/form-data,
// which is what Mailgun documents. mailgun-go sends urlencoded instead and flags
// it in its own source as a misuse (still so in v5.19.1); since Mailgun keeps the
// current setting for params it does not recognise, a body it cannot parse is
// accepted with 200 and changes nothing. Refs #86.
func putTrackingSetting(ctx context.Context, client *http.Client, baseURL, apiKey, domain, setting string, fields map[string]string) error {
	body, contentType := trackingFormBody(fields)

	url := fmt.Sprintf("%s/v3/domains/%s/tracking/%s", baseURL, domain, setting)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return fmt.Errorf("building %s tracking request: %w", setting, err)
	}
	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth("api", apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("updating %s tracking: %w", setting, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		detail, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("updating %s tracking: HTTP %d: %s", setting, resp.StatusCode, bytes.TrimSpace(detail))
	}

	return nil
}

// trackingFormBody discards write errors: the destination is a bytes.Buffer,
// whose writes only ever fail by panicking on overflow.
func trackingFormBody(fields map[string]string) (*bytes.Buffer, string) {
	var body bytes.Buffer

	form := multipart.NewWriter(&body)
	for name, value := range fields {
		_ = form.WriteField(name, value)
	}
	_ = form.Close()

	return &body, form.FormDataContentType()
}
