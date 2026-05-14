package apierrors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
)

func TestCheckResponse_StatusCodeMapping(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       []byte
		wantNil    bool
		checkErr   func(t *testing.T, err error)
	}{
		{
			name:       "200 returns nil",
			statusCode: 200,
			body:       []byte(`{}`),
			wantNil:    true,
		},
		{
			name:       "401 returns UnauthorizedError",
			statusCode: 401,
			body:       []byte(`{"title":"unauthorized"}`),
			checkErr: func(t *testing.T, err error) {
				var ue *apierrors.UnauthorizedError
				if !errors.As(err, &ue) {
					t.Errorf("expected *UnauthorizedError, got %T", err)
				}
			},
		},
		{
			name:       "403 returns ForbiddenError",
			statusCode: 403,
			body:       []byte(`{"title":"forbidden"}`),
			checkErr: func(t *testing.T, err error) {
				var fe *apierrors.ForbiddenError
				if !errors.As(err, &fe) {
					t.Errorf("expected *ForbiddenError, got %T", err)
				}
			},
		},
		{
			name:       "404 returns NotFoundError",
			statusCode: 404,
			body:       []byte(`{"title":"not found","status":404}`),
			checkErr: func(t *testing.T, err error) {
				var nfe *apierrors.NotFoundError
				if !errors.As(err, &nfe) {
					t.Errorf("expected *NotFoundError, got %T", err)
					return
				}
				if nfe.StatusCode != 404 {
					t.Errorf("expected StatusCode 404, got %d", nfe.StatusCode)
				}
			},
		},
		{
			name:       "409 returns ConflictError",
			statusCode: 409,
			body:       []byte(`{"title":"conflict"}`),
			checkErr: func(t *testing.T, err error) {
				var ce *apierrors.ConflictError
				if !errors.As(err, &ce) {
					t.Errorf("expected *ConflictError, got %T", err)
				}
			},
		},
		{
			name:       "429 returns RateLimitError",
			statusCode: 429,
			body:       []byte(`{"title":"rate limited"}`),
			checkErr: func(t *testing.T, err error) {
				var rle *apierrors.RateLimitError
				if !errors.As(err, &rle) {
					t.Errorf("expected *RateLimitError, got %T", err)
				}
			},
		},
		{
			name:       "500 returns ServerError",
			statusCode: 500,
			body:       []byte(`{"title":"internal error"}`),
			checkErr: func(t *testing.T, err error) {
				var se *apierrors.ServerError
				if !errors.As(err, &se) {
					t.Errorf("expected *ServerError, got %T", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := &http.Response{StatusCode: tt.statusCode, Header: http.Header{}}
			err := apierrors.CheckResponse(resp, tt.body)
			if tt.wantNil {
				if err != nil {
					t.Errorf("expected nil, got %v", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
			tt.checkErr(t, err)
		})
	}
}

func TestCheckResponse_TitleAndBodyFallback(t *testing.T) {
	tests := []struct {
		name  string
		resp  *http.Response
		body  []byte
		check func(t *testing.T, err error)
	}{
		{
			name: "non-JSON body falls back to status text",
			resp: &http.Response{StatusCode: 502},
			body: []byte(`<html><body>Bad Gateway</body></html>`),
			check: func(t *testing.T, err error) {
				body := []byte(`<html><body>Bad Gateway</body></html>`)
				var se *apierrors.ServerError
				if !errors.As(err, &se) {
					t.Fatalf("expected *ServerError, got %T", err)
				}
				if se.Title != "Bad Gateway" {
					t.Errorf("expected title %q, got %q", "Bad Gateway", se.Title)
				}
				if se.Body != string(body) {
					t.Errorf("expected full body preserved in Body, got %q", se.Body)
				}
			},
		},
		{
			name: "JSON without title falls back to status text",
			resp: &http.Response{StatusCode: 400},
			body: []byte(`{"detail":"missing field"}`),
			check: func(t *testing.T, err error) {
				body := []byte(`{"detail":"missing field"}`)
				var be *apierrors.BadRequestError
				if !errors.As(err, &be) {
					t.Fatalf("expected *BadRequestError, got %T", err)
				}
				if be.Title != "Bad Request" {
					t.Errorf("expected title %q, got %q", "Bad Request", be.Title)
				}
				if be.Body != string(body) {
					t.Errorf("expected full body preserved in Body, got %q", be.Body)
				}
			},
		},
		{
			name: "long body is preserved in Body",
			resp: &http.Response{StatusCode: 500},
			body: func() []byte {
				b := make([]byte, 500)
				for i := range b {
					b[i] = 'a'
				}
				return b
			}(),
			check: func(t *testing.T, err error) {
				body := make([]byte, 500)
				for i := range body {
					body[i] = 'a'
				}
				var se *apierrors.ServerError
				if !errors.As(err, &se) {
					t.Fatalf("expected *ServerError, got %T", err)
				}
				if se.Title != "Internal Server Error" {
					t.Errorf("expected title %q, got %q", "Internal Server Error", se.Title)
				}
				if se.Body != string(body) {
					t.Errorf("expected full body preserved in Body, got len=%d", len(se.Body))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := apierrors.CheckResponse(tt.resp, tt.body)
			tt.check(t, err)
		})
	}
}

func TestCheckResponse_ParsesTitle(t *testing.T) {
	resp := &http.Response{StatusCode: 400, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"invalid input"}`))
	var be *apierrors.BadRequestError
	if !errors.As(err, &be) {
		t.Fatalf("expected *BadRequestError, got %T", err)
	}
	if be.Title != "invalid input" {
		t.Errorf("expected title 'invalid input', got %q", be.Title)
	}
}
