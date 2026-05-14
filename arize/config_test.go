package arize_test

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize"
)

func mustResolve(t *testing.T, cfg arize.Config) arize.Config {
	t.Helper()
	resolved, err := cfg.Resolve()
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	return resolved
}

// A. TestConfig_Validate merges 7 Validate_* tests.
func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name            string
		cfg             arize.Config
		wantErrIs       error
		wantErrContains []string
	}{
		{
			name:      "missing API key",
			cfg:       arize.Config{},
			wantErrIs: arize.ErrMissingAPIKey,
		},
		{
			name:      "region + single host",
			cfg:       arize.Config{APIKey: "key", Region: arize.RegionEUWest, SingleHost: "custom.host.example.com"},
			wantErrIs: arize.ErrMultipleEndpointOverrides,
		},
		{
			name:            "invalid single port",
			cfg:             arize.Config{APIKey: "key", SingleHost: "host.example.com", SinglePort: 99999},
			wantErrContains: []string{"port"},
		},
		{
			name:      "single port alone triggers override",
			cfg:       arize.Config{APIKey: "key", Region: arize.RegionEUWest, SinglePort: 8443},
			wantErrIs: arize.ErrMultipleEndpointOverrides,
		},
		{
			name:            "max HTTP payload size below 1",
			cfg:             arize.Config{APIKey: "k", MaxHTTPPayloadSizeMB: 0.5},
			wantErrContains: []string{"max_http_payload_size_mb"},
		},
		{
			name:            "max past years below 1",
			cfg:             arize.Config{APIKey: "k", MaxPastYears: -1},
			wantErrContains: []string{"max_past_years"},
		},
		{
			name:            "multiple overrides lists all conflicting fields",
			cfg:             arize.Config{APIKey: "k", Region: arize.RegionEUWest, SingleHost: "host.example.com", BaseDomain: "example.com"},
			wantErrIs:       arize.ErrMultipleEndpointOverrides,
			wantErrContains: []string{"Region=", "SingleHost=", "BaseDomain="},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustResolve(t, tt.cfg)
			err := cfg.Validate()
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("errors.Is(%v): want %v, got %v", tt.wantErrIs, tt.wantErrIs, err)
					return
				}
			}
			if len(tt.wantErrContains) > 0 {
				if err == nil {
					t.Errorf("expected non-nil error containing %v, got nil", tt.wantErrContains)
					return
				}
				for _, sub := range tt.wantErrContains {
					if !strings.Contains(err.Error(), sub) {
						t.Errorf("error %q should contain %q", err.Error(), sub)
					}
				}
			}
		})
	}
}

// B. TestConfig_Resolve_InvalidEnv merges 4 Invalid*Env tests.
func TestConfig_Resolve_InvalidEnv(t *testing.T) {
	tests := []struct {
		name   string
		envVar string
		value  string
	}{
		{"ARIZE_SINGLE_PORT invalid", "ARIZE_SINGLE_PORT", "not-a-number"},
		{"ARIZE_FLIGHT_PORT invalid", "ARIZE_FLIGHT_PORT", "abc"},
		{"ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB invalid", "ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB", "not-a-float"},
		{"ARIZE_MAX_PAST_YEARS invalid", "ARIZE_MAX_PAST_YEARS", "five"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(tt.envVar, tt.value)
			_, err := arize.Config{APIKey: "k"}.Resolve()
			if err == nil {
				t.Fatalf("expected error for malformed %s", tt.envVar)
			}
			if !strings.Contains(err.Error(), tt.envVar) {
				t.Errorf("error should mention env var name %q, got: %v", tt.envVar, err)
			}
		})
	}
}

// C. TestConfig_String_Masking merges 3 String_* tests.
func TestConfig_String_Masking(t *testing.T) {
	tests := []struct {
		name           string
		apiKey         string
		format         string
		wantContains   string
		wantNotContain string
	}{
		{"%v masks API key", "sk-test-deadbeef", "%v", "sk-tes***", "sk-test-deadbeef"},
		{"%+v also masks", "sk-test-deadbeef", "%+v", "", "sk-test-deadbeef"},
		{"empty API key not masked", "", "%v", "", "***"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := fmt.Sprintf(tt.format, arize.Config{APIKey: tt.apiKey})
			if tt.wantContains != "" && !strings.Contains(out, tt.wantContains) {
				t.Errorf("output should contain %q, got: %s", tt.wantContains, out)
			}
			if strings.Contains(out, tt.wantNotContain) {
				t.Errorf("output should NOT contain %q, got: %s", tt.wantNotContain, out)
			}
		})
	}
}

// D. TestConfig_Resolve_Precedence merges EnvVarOverride + ExplicitArgWinsOverEnv.
func TestConfig_Resolve_Precedence(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		explicit arize.Config
		wantKey  string
		wantHost string
	}{
		{
			name:     "env vars override zero-value config",
			env:      map[string]string{"ARIZE_API_KEY": "env-key", "ARIZE_API_HOST": "custom.arize.com"},
			explicit: arize.Config{},
			wantKey:  "env-key",
			wantHost: "custom.arize.com",
		},
		{
			name:     "explicit arg wins over env",
			env:      map[string]string{"ARIZE_API_KEY": "env-key"},
			explicit: arize.Config{APIKey: "explicit-key"},
			wantKey:  "explicit-key",
			wantHost: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			cfg := mustResolve(t, tt.explicit)
			if cfg.APIKey != tt.wantKey {
				t.Errorf("APIKey: want %q, got %q", tt.wantKey, cfg.APIKey)
			}
			if tt.wantHost != "" && cfg.APIHost != tt.wantHost {
				t.Errorf("APIHost: want %q, got %q", tt.wantHost, cfg.APIHost)
			}
		})
	}
}

// E. TestConfig_Headers merges 2 Headers_* tests.
func TestConfig_Headers(t *testing.T) {
	tests := []struct {
		name         string
		cfg          arize.Config
		key          string
		want         string // exact value; empty means "must be non-empty"
		mustNotExist string // if non-empty: assert this key does NOT exist
	}{
		{"authorization is raw key", arize.Config{APIKey: "my-key"}, "authorization", "my-key", ""},
		{"no canonical Authorization key", arize.Config{APIKey: "my-key"}, "", "", "Authorization"},
		{"sdk-language", arize.Config{APIKey: "k"}, "sdk-language", "go", ""},
		{"language-version", arize.Config{APIKey: "k"}, "language-version", runtime.Version(), ""},
		{"sdk-package-name", arize.Config{APIKey: "k"}, "sdk-package-name", "arize", ""},
		{"sdk-version non-empty", arize.Config{APIKey: "k"}, "sdk-version", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headers := tt.cfg.Headers()
			if tt.mustNotExist != "" {
				if _, exists := headers[tt.mustNotExist]; exists {
					t.Errorf("header %q should not exist in headers map", tt.mustNotExist)
				}
			}
			if tt.key != "" {
				if tt.want != "" {
					if got := headers[tt.key]; got != tt.want {
						t.Errorf("header %q: want %q, got %q", tt.key, tt.want, got)
					}
				} else {
					if headers[tt.key] == "" {
						t.Errorf("header %q must be non-empty", tt.key)
					}
				}
			}
		})
	}
}

// F. TestConfig_Resolve_RequestVerifyEnv merges TruthyValues + FalsyValues + Unset.
func TestConfig_Resolve_RequestVerifyEnv(t *testing.T) {
	tests := []struct {
		name                   string
		envValue               string
		envUnset               bool
		wantInsecureSkipVerify bool
	}{
		{"1 → verify=true", "1", false, false},
		{"true → verify=true", "true", false, false},
		{"TRUE → verify=true", "TRUE", false, false},
		{"True → verify=true", "True", false, false},
		{"yes → verify=true", "yes", false, false},
		{"YES → verify=true", "YES", false, false},
		{"on → verify=true", "on", false, false},
		{"ON → verify=true", "ON", false, false},
		{"0 → verify=false", "0", false, true},
		{"false → verify=false", "false", false, true},
		{"FALSE → verify=false", "FALSE", false, true},
		{"False → verify=false", "False", false, true},
		{"no → verify=false", "no", false, true},
		{"NO → verify=false", "NO", false, true},
		{"off → verify=false", "off", false, true},
		{"random → verify=false", "random", false, true},
		{"unset → default false", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.envUnset {
				t.Setenv("ARIZE_REQUEST_VERIFY", tt.envValue)
			}
			cfg := mustResolve(t, arize.Config{APIKey: "k"})
			if cfg.InsecureSkipVerify != tt.wantInsecureSkipVerify {
				t.Errorf("InsecureSkipVerify: want %v, got %v", tt.wantInsecureSkipVerify, cfg.InsecureSkipVerify)
			}
		})
	}
}

// G. TestConfig_Resolve_Hosts merges DefaultsApplied, Resolve_BaseDomain,
// Resolve_SingleHost, Resolve_SinglePort_OnlyAffectsFlight, Resolve_FlightDefaults,
// and the 4 Resolve_Region subtests.
func TestConfig_Resolve_Hosts(t *testing.T) {
	tests := []struct {
		name             string
		cfg              arize.Config
		wantAPIHost      string
		wantOTLPHost     string
		wantFlightHost   string
		wantFlightPort   int
		wantFlightScheme string
		wantAPIURL       string
		wantAPIScheme    string
		wantISV          bool
		checkISV         bool
	}{
		{
			name:             "defaults",
			cfg:              arize.Config{APIKey: "k"},
			wantAPIHost:      "api.arize.com",
			wantOTLPHost:     "otlp.arize.com",
			wantFlightHost:   "flight.arize.com",
			wantFlightPort:   443,
			wantFlightScheme: "grpc+tls",
			wantAPIScheme:    "https",
			checkISV:         true,
			wantISV:          false,
		},
		{
			name:             "BaseDomain",
			cfg:              arize.Config{APIKey: "key", BaseDomain: "example.com"},
			wantAPIHost:      "api.example.com",
			wantOTLPHost:     "otlp.example.com",
			wantFlightHost:   "flight.example.com",
			wantFlightPort:   443,
			wantFlightScheme: "grpc+tls",
			wantAPIURL:       "https://api.example.com",
		},
		{
			name:           "SingleHost",
			cfg:            arize.Config{APIKey: "key", SingleHost: "single.host.example.com"},
			wantAPIHost:    "single.host.example.com",
			wantOTLPHost:   "single.host.example.com",
			wantFlightHost: "single.host.example.com",
			wantAPIURL:     "https://single.host.example.com",
		},
		{
			name:           "SinglePort only affects FlightPort",
			cfg:            arize.Config{APIKey: "key", SingleHost: "single.host.example.com", SinglePort: 8443},
			wantAPIHost:    "single.host.example.com",
			wantOTLPHost:   "single.host.example.com",
			wantFlightPort: 8443,
			wantAPIURL:     "https://single.host.example.com",
		},
		{
			name:           "Region us-central",
			cfg:            arize.Config{APIKey: "key", Region: arize.RegionUSCentral},
			wantAPIHost:    "api.us-central-1a.arize.com",
			wantOTLPHost:   "otlp.us-central-1a.arize.com",
			wantFlightHost: "flight.us-central-1a.arize.com",
			wantFlightPort: 443,
		},
		{
			name:           "Region eu-west",
			cfg:            arize.Config{APIKey: "key", Region: arize.RegionEUWest},
			wantAPIHost:    "api.eu-west-1a.arize.com",
			wantOTLPHost:   "otlp.eu-west-1a.arize.com",
			wantFlightHost: "flight.eu-west-1a.arize.com",
			wantFlightPort: 443,
		},
		{
			name:           "Region ca-central",
			cfg:            arize.Config{APIKey: "key", Region: arize.RegionCACentral},
			wantAPIHost:    "api.ca-central-1a.arize.com",
			wantOTLPHost:   "otlp.ca-central-1a.arize.com",
			wantFlightHost: "flight.ca-central-1a.arize.com",
			wantFlightPort: 443,
		},
		{
			name:           "Region us-east",
			cfg:            arize.Config{APIKey: "key", Region: arize.RegionUSEast},
			wantAPIHost:    "api.us-east-1b.arize.com",
			wantOTLPHost:   "otlp.us-east-1b.arize.com",
			wantFlightHost: "flight.us-east-1b.arize.com",
			wantFlightPort: 443,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := mustResolve(t, tt.cfg)
			if tt.wantAPIHost != "" && cfg.APIHost != tt.wantAPIHost {
				t.Errorf("APIHost: want %q, got %q", tt.wantAPIHost, cfg.APIHost)
			}
			if tt.wantOTLPHost != "" && cfg.OTLPHost != tt.wantOTLPHost {
				t.Errorf("OTLPHost: want %q, got %q", tt.wantOTLPHost, cfg.OTLPHost)
			}
			if tt.wantFlightHost != "" && cfg.FlightHost != tt.wantFlightHost {
				t.Errorf("FlightHost: want %q, got %q", tt.wantFlightHost, cfg.FlightHost)
			}
			if tt.wantFlightPort != 0 && cfg.FlightPort != tt.wantFlightPort {
				t.Errorf("FlightPort: want %d, got %d", tt.wantFlightPort, cfg.FlightPort)
			}
			if tt.wantFlightScheme != "" && cfg.FlightScheme != tt.wantFlightScheme {
				t.Errorf("FlightScheme: want %q, got %q", tt.wantFlightScheme, cfg.FlightScheme)
			}
			if tt.wantAPIURL != "" && cfg.APIURL() != tt.wantAPIURL {
				t.Errorf("APIURL: want %q, got %q", tt.wantAPIURL, cfg.APIURL())
			}
			if tt.wantAPIScheme != "" && cfg.APIScheme != tt.wantAPIScheme {
				t.Errorf("APIScheme: want %q, got %q", tt.wantAPIScheme, cfg.APIScheme)
			}
			if tt.checkISV && cfg.InsecureSkipVerify != tt.wantISV {
				t.Errorf("InsecureSkipVerify: want %v, got %v", tt.wantISV, cfg.InsecureSkipVerify)
			}
		})
	}
}

// H. TestConfig_Resolve_DefaultsForNewFields converts 4 sequential checks into a table.
func TestConfig_Resolve_DefaultsForNewFields(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "k"})
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"MaxHTTPPayloadSizeMB", cfg.MaxHTTPPayloadSizeMB, float64(8)},
		{"ArizeDirectory", cfg.ArizeDirectory, "~/.arize"},
		{"DisableCaching default false", cfg.DisableCaching, false},
		{"MaxPastYears", cfg.MaxPastYears, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("want %v, got %v", tt.want, tt.got)
			}
		})
	}
}

// I. TestConfig_Resolve_EnvOverridesForNewFields converts 4 env-override checks into a table.
func TestConfig_Resolve_EnvOverridesForNewFields(t *testing.T) {
	t.Setenv("ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB", "32.5")
	t.Setenv("ARIZE_DIRECTORY", "/var/arize")
	t.Setenv("ARIZE_ENABLE_CACHING", "false")
	t.Setenv("ARIZE_MAX_PAST_YEARS", "10")
	cfg := mustResolve(t, arize.Config{APIKey: "k"})
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"MaxHTTPPayloadSizeMB from env", cfg.MaxHTTPPayloadSizeMB, 32.5},
		{"ArizeDirectory from env", cfg.ArizeDirectory, "/var/arize"},
		{"DisableCaching=true when ENABLE_CACHING=false", cfg.DisableCaching, true},
		{"MaxPastYears from env", cfg.MaxPastYears, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("want %v, got %v", tt.want, tt.got)
			}
		})
	}
}

// J. TestConfig_APIURL — kept as-is (single-case, no merge target).
func TestConfig_APIURL(t *testing.T) {
	cfg := arize.Config{APIKey: "key", APIHost: "api.arize.com", APIScheme: "https"}
	if url := cfg.APIURL(); url != "https://api.arize.com" {
		t.Errorf("unexpected APIURL: %s", url)
	}
}
