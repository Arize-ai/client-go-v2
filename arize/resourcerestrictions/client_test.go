package resourcerestrictions_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/resourcerestrictions"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *arize.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client, err := arize.NewClient(arize.Config{
		APIKey:    "test-key",
		APIHost:   srv.Listener.Addr().String(),
		APIScheme: "http",
	})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	return srv, client
}

func TestResourceRestrictions_Create(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(generated.ResourceRestrictionResponse{
			ResourceRestriction: generated.ResourceRestriction{
				ResourceId:   "proj-1",
				ResourceType: "PROJECT",
				CreatedAt:    time.Now(),
			},
		})
	})
	resp, err := client.ResourceRestrictions.Create(context.Background(), resourcerestrictions.CreateResourceRestrictionRequestBody{
		ResourceId: "proj-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ResourceRestriction.ResourceId != "proj-1" {
		t.Errorf("unexpected resource_id: %s", resp.ResourceRestriction.ResourceId)
	}
}

func TestResourceRestrictions_Delete_NotFound(t *testing.T) {
	_, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		json.NewEncoder(w).Encode(map[string]interface{}{"title": "not found", "status": 404})
	})
	err := client.ResourceRestrictions.Delete(context.Background(), "nonexistent")
	var nfe *arize.NotFoundError
	if !errors.As(err, &nfe) {
		t.Errorf("expected *NotFoundError, got %T: %v", err, err)
	}
}
