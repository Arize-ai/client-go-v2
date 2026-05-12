package resolve_test

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// b64 produces the encoded form of "Type:1:" + suffix — the shape of real
// Arize global IDs.
func b64(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Resource:1:" + suffix))
}

func TestIsResourceID(t *testing.T) {
	tests := []struct {
		in   string
		want bool
	}{
		{"", false},
		{"my-dataset", false},
		{"production", false},
		{b64("abc"), true},
		{b64("xyz123"), true},
		{"!!!notbase64!!!", false},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			if got := resolve.IsResourceID(tt.in); got != tt.want {
				t.Errorf("IsResourceID(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestResourceNotFoundError_Format(t *testing.T) {
	e := &resolve.ResourceNotFoundError{
		ResourceType: "dataset",
		Name:         "missing",
		Available:    []string{"a", "b", "c"},
	}
	msg := e.Error()
	for _, want := range []string{"dataset", `"missing"`, "Available", "a, b, c"} {
		if !strings.Contains(msg, want) {
			t.Errorf("Error() %q missing %q", msg, want)
		}
	}
}

func TestResourceNotFoundError_WithHint(t *testing.T) {
	e := &resolve.ResourceNotFoundError{
		ResourceType: "dataset",
		Name:         "missing",
		Hint:         "Provide 'space'.",
	}
	if !strings.Contains(e.Error(), "Provide 'space'.") {
		t.Errorf("hint missing from %q", e.Error())
	}
}

// newTestGen wraps an httptest server in a generated client, returning both.
func newTestGen(t *testing.T, h http.HandlerFunc) *generated.ClientWithResponses {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	gen, err := generated.NewClientWithResponses("http://" + srv.Listener.Addr().String())
	if err != nil {
		t.Fatalf("NewClientWithResponses: %v", err)
	}
	return gen
}

func TestFindDatasetID_PassesThroughForResourceID(t *testing.T) {
	id := b64("ds-1")
	called := false
	gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(500)
	})
	got, err := resolve.FindDatasetID(context.Background(), gen, id, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != id {
		t.Errorf("want %q, got %q", id, got)
	}
	if called {
		t.Error("server should not have been hit when input is an ID")
	}
}

func TestFindDatasetID_ResolvesByName(t *testing.T) {
	dsID := b64("ds-real")
	gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("name"); got != "my-dataset" {
			t.Errorf("name query: want my-dataset, got %q", got)
		}
		if got := r.URL.Query().Get("space_name"); got != "demo" {
			t.Errorf("space_name query: want demo, got %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"datasets":[{"id":"` + dsID + `","name":"my-dataset","space_id":"sp","created_at":"2026-01-01T00:00:00Z"}],"pagination":{"has_more":false}}`))
	})
	got, err := resolve.FindDatasetID(context.Background(), gen, "my-dataset", "demo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != dsID {
		t.Errorf("want %q, got %q", dsID, got)
	}
}

func TestFindDatasetID_NameWithoutSpace_Errors(t *testing.T) {
	gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called when space is missing")
		w.WriteHeader(500)
	})
	_, err := resolve.FindDatasetID(context.Background(), gen, "my-dataset", "")
	var rnfe *resolve.ResourceNotFoundError
	if !errors.As(err, &rnfe) {
		t.Fatalf("want *ResourceNotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(rnfe.Hint, "space") {
		t.Errorf("hint should mention space, got %q", rnfe.Hint)
	}
}

func TestFindDatasetID_NotFound(t *testing.T) {
	gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"datasets":[{"id":"` + b64("ds-other") + `","name":"other","space_id":"sp","created_at":"2026-01-01T00:00:00Z"}],"pagination":{"has_more":false}}`))
	})
	_, err := resolve.FindDatasetID(context.Background(), gen, "missing", "demo")
	var rnfe *resolve.ResourceNotFoundError
	if !errors.As(err, &rnfe) {
		t.Fatalf("want *ResourceNotFoundError, got %T: %v", err, err)
	}
	if rnfe.Name != "missing" {
		t.Errorf("Name: want missing, got %q", rnfe.Name)
	}
	if len(rnfe.Available) != 1 || rnfe.Available[0] != "other" {
		t.Errorf("Available: want [other], got %v", rnfe.Available)
	}
}

func TestFindExperimentID_DatasetNameWithoutSpace_Errors(t *testing.T) {
	gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not be called")
		w.WriteHeader(500)
	})
	_, err := resolve.FindExperimentID(context.Background(), gen, "exp", "ds-by-name", "")
	var rnfe *resolve.ResourceNotFoundError
	if !errors.As(err, &rnfe) {
		t.Fatalf("want *ResourceNotFoundError, got %T: %v", err, err)
	}
	if !strings.Contains(rnfe.Hint, "space") {
		t.Errorf("hint should mention space, got %q", rnfe.Hint)
	}
}
