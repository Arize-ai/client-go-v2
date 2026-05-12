package apierrors_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
)

func TestCheckResponse_200_ReturnsNil(t *testing.T) {
	resp := &http.Response{StatusCode: 200}
	if err := apierrors.CheckResponse(resp, []byte(`{}`)); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestCheckResponse_404_ReturnsNotFoundError(t *testing.T) {
	resp := &http.Response{StatusCode: 404, Header: http.Header{}}
	body := []byte(`{"title":"not found","status":404}`)
	err := apierrors.CheckResponse(resp, body)
	var nfe *apierrors.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected *NotFoundError, got %T", err)
	}
	if nfe.StatusCode != 404 {
		t.Errorf("expected StatusCode 404, got %d", nfe.StatusCode)
	}
}

func TestCheckResponse_401_ReturnsUnauthorizedError(t *testing.T) {
	resp := &http.Response{StatusCode: 401, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"unauthorized"}`))
	var ue *apierrors.UnauthorizedError
	if !errors.As(err, &ue) {
		t.Errorf("expected *UnauthorizedError, got %T", err)
	}
}

func TestCheckResponse_403_ReturnsForbiddenError(t *testing.T) {
	resp := &http.Response{StatusCode: 403, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"forbidden"}`))
	var fe *apierrors.ForbiddenError
	if !errors.As(err, &fe) {
		t.Errorf("expected *ForbiddenError, got %T", err)
	}
}

func TestCheckResponse_409_ReturnsConflictError(t *testing.T) {
	resp := &http.Response{StatusCode: 409, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"conflict"}`))
	var ce *apierrors.ConflictError
	if !errors.As(err, &ce) {
		t.Errorf("expected *ConflictError, got %T", err)
	}
}

func TestCheckResponse_429_ReturnsRateLimitError(t *testing.T) {
	resp := &http.Response{StatusCode: 429, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"rate limited"}`))
	var rle *apierrors.RateLimitError
	if !errors.As(err, &rle) {
		t.Errorf("expected *RateLimitError, got %T", err)
	}
}

func TestCheckResponse_500_ReturnsServerError(t *testing.T) {
	resp := &http.Response{StatusCode: 500, Header: http.Header{}}
	err := apierrors.CheckResponse(resp, []byte(`{"title":"internal error"}`))
	var se *apierrors.ServerError
	if !errors.As(err, &se) {
		t.Errorf("expected *ServerError, got %T", err)
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

func TestCheckResponse_NonJSONBody_FallsBackToStatusText(t *testing.T) {
	resp := &http.Response{StatusCode: 502}
	body := []byte(`<html><body>Bad Gateway</body></html>`)
	err := apierrors.CheckResponse(resp, body)
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
}

func TestCheckResponse_JSONWithoutTitle_FallsBackToStatusText(t *testing.T) {
	resp := &http.Response{StatusCode: 400}
	body := []byte(`{"detail":"missing field"}`)
	err := apierrors.CheckResponse(resp, body)
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
}

func TestCheckResponse_LongBody_PreservedInBody(t *testing.T) {
	resp := &http.Response{StatusCode: 500}
	body := make([]byte, 500)
	for i := range body {
		body[i] = 'a'
	}
	err := apierrors.CheckResponse(resp, body)
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
}
