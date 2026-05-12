package arize_test

import (
	"errors"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
)

func TestAPIError_ErrorString(t *testing.T) {
	e := &arize.NotFoundError{APIError: arize.APIError{StatusCode: 404, Title: "not found"}}
	if e.Error() == "" {
		t.Error("expected non-empty error string")
	}
}

func TestAPIError_IsWrappable(t *testing.T) {
	// Construct via the typed error struct directly — the public package no
	// longer exposes CheckResponse; the wrapping behavior is exercised through
	// the apierrors package tests.
	err := &arize.NotFoundError{APIError: arize.APIError{StatusCode: 404, Title: "not found"}}

	var apiErr *arize.NotFoundError
	if !errors.As(err, &apiErr) {
		t.Error("errors.As should unwrap to *NotFoundError")
	}
	var baseErr *arize.APIError
	if !errors.As(err, &baseErr) {
		t.Error("errors.As should also unwrap to *APIError base")
	}
}
