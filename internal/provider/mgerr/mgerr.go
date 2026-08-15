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

// statusError carries a caller-supplied message alongside the status the SDK
// expects to find via errors.As. Its own Error() never delegates to the
// wrapped *mailgun.UnexpectedResponseError: that type's Error() is an alias
// for the deprecated String(), which dumps Method/URL/ExpectedOneOf with
// %#v and would leak Go syntax into a terraform apply diagnostic.
type statusError struct {
	msg     string
	wrapped *mailgun.UnexpectedResponseError
}

func (e *statusError) Error() string { return e.msg }
func (e *statusError) Unwrap() error { return e.wrapped }

// StatusError builds an error whose message is exactly msg, but which
// mailgun.GetStatusFromErr (and therefore IsNotFound) resolves to status.
// It exists for hand-rolled API clients that don't go through the SDK's own
// request path and so never produce a real *mailgun.UnexpectedResponseError.
// Expected, Method, URL and Data are intentionally left zero: nothing reads
// them, since msg already carries the readable message.
func StatusError(msg string, status int) error {
	return &statusError{
		msg:     msg,
		wrapped: &mailgun.UnexpectedResponseError{Actual: status},
	}
}
