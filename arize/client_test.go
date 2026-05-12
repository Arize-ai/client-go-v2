package arize_test

import (
	"errors"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
)

func TestNewClient_ValidConfig(t *testing.T) {
	client, err := arize.NewClient(arize.Config{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_MissingAPIKey_ReturnsError(t *testing.T) {
	_, err := arize.NewClient(arize.Config{})
	if !errors.Is(err, arize.ErrMissingAPIKey) {
		t.Errorf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestNewClient_MultipleOverrides_ReturnsError(t *testing.T) {
	_, err := arize.NewClient(arize.Config{
		APIKey:     "key",
		Region:     arize.RegionEUWest,
		SingleHost: "host.example.com",
	})
	if !errors.Is(err, arize.ErrMultipleEndpointOverrides) {
		t.Errorf("expected ErrMultipleEndpointOverrides, got %v", err)
	}
}

func TestClient_Config_ReturnsResolved(t *testing.T) {
	client, err := arize.NewClient(arize.Config{
		APIKey: "test-key",
		Region: arize.RegionEUWest,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cfg := client.Config()
	if cfg.APIKey != "test-key" {
		t.Errorf("APIKey: want test-key, got %s", cfg.APIKey)
	}
	if cfg.APIHost != "api.eu-west-1a.arize.com" {
		t.Errorf("APIHost should reflect region resolution, got %s", cfg.APIHost)
	}
}
