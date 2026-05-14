package arize_test

import (
	"errors"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
)

func TestAPIError_TypedErrors(t *testing.T) {
	tests := []struct {
		name  string
		err   error
		check func(t *testing.T, err error)
	}{
		{
			name: "NotFoundError 404",
			err:  &arize.NotFoundError{APIError: arize.APIError{StatusCode: 404, Title: "not found"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.NotFoundError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *NotFoundError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "BadRequestError 400",
			err:  &arize.BadRequestError{APIError: arize.APIError{StatusCode: 400, Title: "bad request"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.BadRequestError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *BadRequestError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "UnauthorizedError 401",
			err:  &arize.UnauthorizedError{APIError: arize.APIError{StatusCode: 401, Title: "unauthorized"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.UnauthorizedError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *UnauthorizedError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "ForbiddenError 403",
			err:  &arize.ForbiddenError{APIError: arize.APIError{StatusCode: 403, Title: "forbidden"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.ForbiddenError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *ForbiddenError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "ConflictError 409",
			err:  &arize.ConflictError{APIError: arize.APIError{StatusCode: 409, Title: "conflict"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.ConflictError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *ConflictError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "RateLimitError 429",
			err:  &arize.RateLimitError{APIError: arize.APIError{StatusCode: 429, Title: "rate limit"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.RateLimitError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *RateLimitError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
		{
			name: "ServerError 500",
			err:  &arize.ServerError{APIError: arize.APIError{StatusCode: 500, Title: "server error"}},
			check: func(t *testing.T, err error) {
				t.Helper()
				var ptr *arize.ServerError
				if !errors.As(err, &ptr) {
					t.Errorf("errors.As should unwrap to *ServerError")
				}
				if !errors.As(err, new(*arize.APIError)) {
					t.Errorf("errors.As should unwrap to *APIError")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() == "" {
				t.Errorf("expected non-empty error string")
			}
			tt.check(t, tt.err)
		})
	}
}
