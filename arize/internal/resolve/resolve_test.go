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
// Arize unique identifiers.
func b64(suffix string) string {
	return base64.StdEncoding.EncodeToString([]byte("Resource:1:" + suffix))
}

func TestIsResourceID(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{name: "empty string", in: "", want: false},
		{name: "plain name", in: "my-dataset", want: false},
		{name: "production name", in: "production", want: false},
		{name: "valid base64 id abc", in: b64("abc"), want: true},
		{name: "valid base64 id xyz123", in: b64("xyz123"), want: true},
		{name: "invalid base64", in: "!!!notbase64!!!", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

	tests := []struct {
		name string
		want string
	}{
		{"includes resource type", "dataset"},
		{"includes quoted name", `"missing"`},
		{"includes Available header", "Available"},
		{"includes items", "a, b, c"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(msg, tt.want) {
				t.Errorf("Error() %q missing %q", msg, tt.want)
			}
		})
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

func TestFind_NameWithoutSpace_Errors(t *testing.T) {
	tests := []struct {
		name             string
		invoke           func(ctx context.Context, gen *generated.ClientWithResponses) (string, error)
		wantHintContains string
	}{
		{
			name: "FindDatasetID requires space name",
			invoke: func(ctx context.Context, gen *generated.ClientWithResponses) (string, error) {
				return resolve.FindDatasetID(ctx, gen, "my-dataset", "")
			},
			wantHintContains: "space",
		},
		{
			name: "FindExperimentID requires space name via dataset",
			invoke: func(ctx context.Context, gen *generated.ClientWithResponses) (string, error) {
				return resolve.FindExperimentID(ctx, gen, "exp", "ds-by-name", "")
			},
			wantHintContains: "space",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
				t.Error("server should not be called when space is missing")
				w.WriteHeader(500)
			})
			_, err := tt.invoke(context.Background(), gen)
			var rnfe *resolve.ResourceNotFoundError
			if !errors.As(err, &rnfe) {
				t.Fatalf("want *ResourceNotFoundError, got %T: %v", err, err)
			}
			if !strings.Contains(rnfe.Hint, tt.wantHintContains) {
				t.Errorf("hint should contain %q, got %q", tt.wantHintContains, rnfe.Hint)
			}
		})
	}
}

func TestFindDatasetID(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T) *generated.ClientWithResponses
		input   string
		space   string
		wantID  string
		wantErr func(t *testing.T, err error)
	}{
		{
			name:  "passes through resource ID",
			input: b64("ds-1"),
			space: "",
			setup: func(t *testing.T) *generated.ClientWithResponses {
				t.Helper()
				called := false
				gen := newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
					called = true
					w.WriteHeader(500)
				})
				t.Cleanup(func() {
					if called {
						t.Error("server should not have been hit when input is an ID")
					}
				})
				return gen
			},
			wantID:  b64("ds-1"),
			wantErr: nil,
		},
		{
			name:  "resolves by name",
			input: "my-dataset",
			space: "demo",
			setup: func(t *testing.T) *generated.ClientWithResponses {
				t.Helper()
				dsID := b64("ds-real")
				return newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
					if got := r.URL.Query().Get("name"); got != "my-dataset" {
						t.Errorf("name query: want my-dataset, got %q", got)
					}
					if got := r.URL.Query().Get("space_name"); got != "demo" {
						t.Errorf("space_name query: want demo, got %q", got)
					}
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"datasets":[{"id":"` + dsID + `","name":"my-dataset","space_id":"sp","created_at":"2026-01-01T00:00:00Z"}],"pagination":{"has_more":false}}`))
				})
			},
			wantID:  b64("ds-real"),
			wantErr: nil,
		},
		{
			name:  "not found returns error with available names",
			input: "missing",
			space: "demo",
			setup: func(t *testing.T) *generated.ClientWithResponses {
				t.Helper()
				return newTestGen(t, func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					_, _ = w.Write([]byte(`{"datasets":[{"id":"` + b64("ds-other") + `","name":"other","space_id":"sp","created_at":"2026-01-01T00:00:00Z"}],"pagination":{"has_more":false}}`))
				})
			},
			wantID: "",
			wantErr: func(t *testing.T, err error) {
				t.Helper()
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gen := tt.setup(t)
			got, err := resolve.FindDatasetID(context.Background(), gen, tt.input, tt.space)
			if tt.wantErr != nil {
				tt.wantErr(t, err)
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantID {
				t.Errorf("want %q, got %q", tt.wantID, got)
			}
		})
	}
}
