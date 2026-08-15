// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

package mgerr

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/mailgun/mailgun-go/v5"
)

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "404 UnexpectedResponseError",
			err:  &mailgun.UnexpectedResponseError{Actual: 404},
			want: true,
		},
		{
			name: "404 wrapped via %w",
			err:  fmt.Errorf("reading domain: %w", &mailgun.UnexpectedResponseError{Actual: 404}),
			want: true,
		},
		{
			name: "429 RateLimitedError",
			err:  &mailgun.RateLimitedError{Err: &mailgun.UnexpectedResponseError{Actual: 429}},
			want: false,
		},
		{
			name: "500",
			err:  &mailgun.UnexpectedResponseError{Actual: 500},
			want: false,
		},
		{
			name: "transport failure",
			err:  &url.Error{Op: "Get", URL: "https://api.mailgun.net/v3/domains/example.com", Err: errors.New("connection refused")},
			want: false,
		},
		{
			name: "bare error mentioning not found",
			err:  errors.New("not found"),
			want: false,
		},
		{
			name: "nil",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNotFound(tt.err); got != tt.want {
				t.Errorf("IsNotFound(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
