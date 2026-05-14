package arize_test

import (
	"errors"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		cfg     arize.Config
		wantErr error
	}{
		{
			name:    "valid config",
			cfg:     arize.Config{APIKey: "test-key"},
			wantErr: nil,
		},
		{
			name:    "missing API key",
			cfg:     arize.Config{},
			wantErr: arize.ErrMissingAPIKey,
		},
		{
			name: "multiple endpoint overrides",
			cfg: arize.Config{
				APIKey:     "key",
				Region:     arize.RegionEUWest,
				SingleHost: "host.example.com",
			},
			wantErr: arize.ErrMultipleEndpointOverrides,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := arize.NewClient(tt.cfg)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Errorf("want error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if client == nil {
				t.Fatal("expected non-nil client")
			}
		})
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
