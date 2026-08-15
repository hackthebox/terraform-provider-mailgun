// Copyright Hack The Box 2025, 2026
// SPDX-License-Identifier: MPL-2.0

// Package mgerr gives every resource one typed way to ask "was that a 404?"
package mgerr

import (
	"net/http"

	"github.com/mailgun/mailgun-go/v5"
)

// IsNotFound reports whether err represents an HTTP 404 response from
// Mailgun. It resolves through wrapped errors the same way
// mailgun.GetStatusFromErr does (errors.As against *mailgun.UnexpectedResponseError),
// so a %w-wrapped error or one nested inside *mailgun.RateLimitedError still
// matches. Transport failures, other status codes and a nil error all report
// -1 from the SDK and so are never mistaken for a 404.
func IsNotFound(err error) bool {
	return mailgun.GetStatusFromErr(err) == http.StatusNotFound
}
