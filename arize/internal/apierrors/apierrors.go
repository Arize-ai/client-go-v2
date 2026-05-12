// Package apierrors provides shared HTTP error types and response checking
// for all Arize subclient packages. It lives in internal/ to avoid import
// cycles between the root arize package and its subclient children.
package apierrors

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// APIError is the base type for all Arize HTTP API errors.
type APIError struct {
	StatusCode int
	Title      string
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("arize API error %d: %s", e.StatusCode, e.Title)
}

// Typed error subtypes — one per HTTP status class.

type BadRequestError struct{ APIError }
type UnauthorizedError struct{ APIError }
type ForbiddenError struct{ APIError }
type NotFoundError struct{ APIError }
type ConflictError struct{ APIError }
type RateLimitError struct{ APIError }
type ServerError struct{ APIError }

// Unwrap methods allow errors.As to traverse from a subtype up to *APIError.

func (e *BadRequestError) Unwrap() error   { return &e.APIError }
func (e *UnauthorizedError) Unwrap() error { return &e.APIError }
func (e *ForbiddenError) Unwrap() error    { return &e.APIError }
func (e *NotFoundError) Unwrap() error     { return &e.APIError }
func (e *ConflictError) Unwrap() error     { return &e.APIError }
func (e *RateLimitError) Unwrap() error    { return &e.APIError }
func (e *ServerError) Unwrap() error       { return &e.APIError }

// CheckResponse converts a non-2xx HTTP response into a typed error.
// Returns nil for status codes below 400.
func CheckResponse(resp *http.Response, body []byte) error {
	if resp.StatusCode < 400 {
		return nil
	}

	base := APIError{
		StatusCode: resp.StatusCode,
		Body:       string(body),
	}

	var problem struct {
		Title string `json:"title"`
	}
	if err := json.Unmarshal(body, &problem); err == nil && problem.Title != "" {
		base.Title = problem.Title
	} else {
		base.Title = http.StatusText(resp.StatusCode)
	}

	switch resp.StatusCode {
	case 400:
		return &BadRequestError{base}
	case 401:
		return &UnauthorizedError{base}
	case 403:
		return &ForbiddenError{base}
	case 404:
		return &NotFoundError{base}
	case 409:
		return &ConflictError{base}
	case 429:
		return &RateLimitError{base}
	default:
		if resp.StatusCode >= 500 {
			return &ServerError{base}
		}
		return &base
	}
}
