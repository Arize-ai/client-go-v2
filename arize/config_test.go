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

func TestConfig_DefaultsApplied(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "test-key"})
	if cfg.APIHost != "api.arize.com" {
		t.Errorf("expected api.arize.com, got %s", cfg.APIHost)
	}
	if cfg.APIScheme != "https" {
		t.Errorf("expected https, got %s", cfg.APIScheme)
	}
	if cfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify false (TLS verification enabled by default)")
	}
}

func TestConfig_EnvVarOverride(t *testing.T) {
	t.Setenv("ARIZE_API_KEY", "env-key")
	t.Setenv("ARIZE_API_HOST", "custom.arize.com")

	cfg := mustResolve(t, arize.Config{})
	if cfg.APIKey != "env-key" {
		t.Errorf("expected env-key, got %s", cfg.APIKey)
	}
	if cfg.APIHost != "custom.arize.com" {
		t.Errorf("expected custom.arize.com, got %s", cfg.APIHost)
	}
}

func TestConfig_ExplicitArgWinsOverEnv(t *testing.T) {
	t.Setenv("ARIZE_API_KEY", "env-key")

	cfg := mustResolve(t, arize.Config{APIKey: "explicit-key"})
	if cfg.APIKey != "explicit-key" {
		t.Errorf("expected explicit-key, got %s", cfg.APIKey)
	}
}

func TestConfig_Validate_MissingAPIKey(t *testing.T) {
	cfg := mustResolve(t, arize.Config{})
	if err := cfg.Validate(); !errors.Is(err, arize.ErrMissingAPIKey) {
		t.Errorf("expected ErrMissingAPIKey, got %v", err)
	}
}

func TestConfig_Validate_MultipleEndpointOverrides(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		Region:     arize.RegionEUWest,
		SingleHost: "custom.host.example.com",
	})
	if err := cfg.Validate(); !errors.Is(err, arize.ErrMultipleEndpointOverrides) {
		t.Errorf("expected ErrMultipleEndpointOverrides, got %v", err)
	}
}

func TestConfig_Validate_InvalidSinglePort(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		SingleHost: "host.example.com",
		SinglePort: 99999,
	})
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid port")
	}
}

func TestConfig_APIURL(t *testing.T) {
	cfg := arize.Config{APIKey: "key", APIHost: "api.arize.com", APIScheme: "https"}
	if url := cfg.APIURL(); url != "https://api.arize.com" {
		t.Errorf("unexpected APIURL: %s", url)
	}
}

func TestConfig_Resolve_BaseDomain(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		BaseDomain: "example.com",
	})
	if cfg.APIHost != "api.example.com" {
		t.Errorf("APIHost: want api.example.com, got %s", cfg.APIHost)
	}
	if cfg.OTLPHost != "otlp.example.com" {
		t.Errorf("OTLPHost: want otlp.example.com, got %s", cfg.OTLPHost)
	}
	if cfg.FlightHost != "flight.example.com" {
		t.Errorf("FlightHost: want flight.example.com, got %s", cfg.FlightHost)
	}
	if cfg.APIURL() != "https://api.example.com" {
		t.Errorf("APIURL: want https://api.example.com, got %s", cfg.APIURL())
	}
}

func TestConfig_Resolve_SingleHost(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		SingleHost: "single.host.example.com",
	})
	if cfg.APIHost != "single.host.example.com" || cfg.OTLPHost != "single.host.example.com" || cfg.FlightHost != "single.host.example.com" {
		t.Errorf("expected all hosts to be single.host.example.com, got api=%s otlp=%s flight=%s", cfg.APIHost, cfg.OTLPHost, cfg.FlightHost)
	}
	if cfg.APIURL() != "https://single.host.example.com" {
		t.Errorf("APIURL: want https://single.host.example.com, got %s", cfg.APIURL())
	}
}

func TestConfig_Resolve_SinglePort_OnlyAffectsFlight(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		SingleHost: "single.host.example.com",
		SinglePort: 8443,
	})
	if cfg.FlightPort != 8443 {
		t.Errorf("FlightPort: want 8443, got %d", cfg.FlightPort)
	}
	// SinglePort must not change the REST URL — matches Python's behavior.
	if cfg.APIURL() != "https://single.host.example.com" {
		t.Errorf("APIURL: want https://single.host.example.com (port not appended), got %s", cfg.APIURL())
	}
}

func TestConfig_Resolve_Region(t *testing.T) {
	tests := []struct {
		region     arize.Region
		wantAPI    string
		wantOTLP   string
		wantFlight string
	}{
		{arize.RegionUSCentral, "api.us-central-1a.arize.com", "otlp.us-central-1a.arize.com", "flight.us-central-1a.arize.com"},
		{arize.RegionEUWest, "api.eu-west-1a.arize.com", "otlp.eu-west-1a.arize.com", "flight.eu-west-1a.arize.com"},
		{arize.RegionCACentral, "api.ca-central-1a.arize.com", "otlp.ca-central-1a.arize.com", "flight.ca-central-1a.arize.com"},
		{arize.RegionUSEast, "api.us-east-1b.arize.com", "otlp.us-east-1b.arize.com", "flight.us-east-1b.arize.com"},
	}
	for _, tt := range tests {
		t.Run(string(tt.region), func(t *testing.T) {
			cfg := mustResolve(t, arize.Config{APIKey: "key", Region: tt.region})
			if cfg.APIHost != tt.wantAPI {
				t.Errorf("APIHost: want %s, got %s", tt.wantAPI, cfg.APIHost)
			}
			if cfg.OTLPHost != tt.wantOTLP {
				t.Errorf("OTLPHost: want %s, got %s", tt.wantOTLP, cfg.OTLPHost)
			}
			if cfg.FlightHost != tt.wantFlight {
				t.Errorf("FlightHost: want %s, got %s", tt.wantFlight, cfg.FlightHost)
			}
			if cfg.FlightPort != 443 {
				t.Errorf("FlightPort: want 443, got %d", cfg.FlightPort)
			}
		})
	}
}

func TestConfig_Resolve_FlightDefaults(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "key"})
	if cfg.FlightHost != "flight.arize.com" {
		t.Errorf("FlightHost default: got %s", cfg.FlightHost)
	}
	if cfg.FlightPort != 443 {
		t.Errorf("FlightPort default: got %d", cfg.FlightPort)
	}
	if cfg.FlightScheme != "grpc+tls" {
		t.Errorf("FlightScheme default: got %s", cfg.FlightScheme)
	}
	if cfg.OTLPHost != "otlp.arize.com" {
		t.Errorf("OTLPHost default: got %s", cfg.OTLPHost)
	}
}

func TestConfig_Validate_SinglePortAlone_TriggersOverride(t *testing.T) {
	// SinglePort with Region should fail mutual-exclusion (Python treats
	// SingleHost and SinglePort as one bucket).
	cfg := mustResolve(t, arize.Config{
		APIKey:     "key",
		Region:     arize.RegionEUWest,
		SinglePort: 8443,
	})
	if err := cfg.Validate(); !errors.Is(err, arize.ErrMultipleEndpointOverrides) {
		t.Errorf("expected ErrMultipleEndpointOverrides, got %v", err)
	}
}

func TestConfig_Resolve_InvalidSinglePortEnv(t *testing.T) {
	t.Setenv("ARIZE_SINGLE_PORT", "not-a-number")

	_, err := arize.Config{APIKey: "key"}.Resolve()
	if err == nil {
		t.Fatal("expected error for malformed ARIZE_SINGLE_PORT")
	}
	if !strings.Contains(err.Error(), "ARIZE_SINGLE_PORT") {
		t.Errorf("error should mention env var name, got: %v", err)
	}
}

func TestConfig_Resolve_InvalidFlightPortEnv(t *testing.T) {
	t.Setenv("ARIZE_FLIGHT_PORT", "abc")

	_, err := arize.Config{APIKey: "key"}.Resolve()
	if err == nil {
		t.Fatal("expected error for malformed ARIZE_FLIGHT_PORT")
	}
	if !strings.Contains(err.Error(), "ARIZE_FLIGHT_PORT") {
		t.Errorf("error should mention env var name, got: %v", err)
	}
}

// --- Python v8 parity tests ---

func TestConfig_Headers_AuthorizationIsRawKey(t *testing.T) {
	cfg := arize.Config{APIKey: "my-key"}
	headers := cfg.Headers()
	if got := headers["authorization"]; got != "my-key" {
		t.Errorf("authorization header: want raw key %q (matching Python v8, no Bearer prefix), got %q", "my-key", got)
	}
	if _, hasBearer := headers["Authorization"]; hasBearer {
		t.Error("unexpected canonical Authorization key; Python v8 sends lowercase 'authorization'")
	}
}

func TestConfig_Headers_IncludesAllSDKHeaders(t *testing.T) {
	cfg := arize.Config{APIKey: "k"}
	headers := cfg.Headers()
	want := map[string]string{
		"sdk-language":     "go",
		"language-version": runtime.Version(),
		"sdk-package-name": "arize",
	}
	for k, v := range want {
		if got := headers[k]; got != v {
			t.Errorf("header %q: want %q, got %q", k, v, got)
		}
	}
	if headers["sdk-version"] == "" {
		t.Error("sdk-version header must be non-empty")
	}
}

func TestConfig_Resolve_RequestVerifyEnv_TruthyValues(t *testing.T) {
	for _, v := range []string{"1", "true", "TRUE", "True", "yes", "YES", "on", "ON"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("ARIZE_REQUEST_VERIFY", v)
			cfg := mustResolve(t, arize.Config{APIKey: "k"})
			if cfg.InsecureSkipVerify {
				t.Errorf("ARIZE_REQUEST_VERIFY=%q should mean verify=true → InsecureSkipVerify=false", v)
			}
		})
	}
}

func TestConfig_Resolve_RequestVerifyEnv_FalsyValues(t *testing.T) {
	for _, v := range []string{"0", "false", "FALSE", "False", "no", "NO", "off", "random"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("ARIZE_REQUEST_VERIFY", v)
			cfg := mustResolve(t, arize.Config{APIKey: "k"})
			if !cfg.InsecureSkipVerify {
				t.Errorf("ARIZE_REQUEST_VERIFY=%q should mean verify=false → InsecureSkipVerify=true", v)
			}
		})
	}
}

func TestConfig_Resolve_RequestVerifyEnv_Unset(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "k"})
	if cfg.InsecureSkipVerify {
		t.Error("unset ARIZE_REQUEST_VERIFY should keep zero-value (InsecureSkipVerify=false, TLS verified)")
	}
}

func TestConfig_Resolve_DefaultsForNewFields(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "k"})
	if cfg.MaxHTTPPayloadSizeMB != 8 {
		t.Errorf("MaxHTTPPayloadSizeMB: want 8, got %v", cfg.MaxHTTPPayloadSizeMB)
	}
	if cfg.ArizeDirectory != "~/.arize" {
		t.Errorf("ArizeDirectory: want ~/.arize, got %q", cfg.ArizeDirectory)
	}
	if cfg.DisableCaching {
		t.Error("DisableCaching default: want false (caching enabled, matching Python's enable_caching=True)")
	}
	if cfg.MaxPastYears != 5 {
		t.Errorf("MaxPastYears: want 5, got %d", cfg.MaxPastYears)
	}
}

func TestConfig_Resolve_EnvOverridesForNewFields(t *testing.T) {
	t.Setenv("ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB", "32.5")
	t.Setenv("ARIZE_DIRECTORY", "/var/arize")
	t.Setenv("ARIZE_ENABLE_CACHING", "false")
	t.Setenv("ARIZE_MAX_PAST_YEARS", "10")

	cfg := mustResolve(t, arize.Config{APIKey: "k"})
	if cfg.MaxHTTPPayloadSizeMB != 32.5 {
		t.Errorf("MaxHTTPPayloadSizeMB: want 32.5, got %v", cfg.MaxHTTPPayloadSizeMB)
	}
	if cfg.ArizeDirectory != "/var/arize" {
		t.Errorf("ArizeDirectory: want /var/arize, got %q", cfg.ArizeDirectory)
	}
	if !cfg.DisableCaching {
		t.Error("ARIZE_ENABLE_CACHING=false should flip DisableCaching to true")
	}
	if cfg.MaxPastYears != 10 {
		t.Errorf("MaxPastYears: want 10, got %d", cfg.MaxPastYears)
	}
}

func TestConfig_Resolve_InvalidMaxHTTPPayloadSizeMBEnv(t *testing.T) {
	t.Setenv("ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB", "not-a-float")
	_, err := arize.Config{APIKey: "k"}.Resolve()
	if err == nil {
		t.Fatal("expected error for malformed ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB")
	}
	if !strings.Contains(err.Error(), "ARIZE_MAX_HTTP_PAYLOAD_SIZE_MB") {
		t.Errorf("error should mention env var name, got: %v", err)
	}
}

func TestConfig_Resolve_InvalidMaxPastYearsEnv(t *testing.T) {
	t.Setenv("ARIZE_MAX_PAST_YEARS", "five")
	_, err := arize.Config{APIKey: "k"}.Resolve()
	if err == nil {
		t.Fatal("expected error for malformed ARIZE_MAX_PAST_YEARS")
	}
	if !strings.Contains(err.Error(), "ARIZE_MAX_PAST_YEARS") {
		t.Errorf("error should mention env var name, got: %v", err)
	}
}

func TestConfig_Validate_MaxHTTPPayloadSizeMBBelowOne(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "k", MaxHTTPPayloadSizeMB: 0.5})
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "max_http_payload_size_mb") {
		t.Errorf("expected validation error for MaxHTTPPayloadSizeMB < 1, got: %v", err)
	}
}

func TestConfig_Validate_MaxPastYearsBelowOne(t *testing.T) {
	cfg := mustResolve(t, arize.Config{APIKey: "k", MaxPastYears: -1})
	if err := cfg.Validate(); err == nil || !strings.Contains(err.Error(), "max_past_years") {
		t.Errorf("expected validation error for MaxPastYears < 1, got: %v", err)
	}
}

func TestConfig_Validate_MultipleOverridesErrorListsConflicts(t *testing.T) {
	cfg := mustResolve(t, arize.Config{
		APIKey:     "k",
		Region:     arize.RegionEUWest,
		SingleHost: "host.example.com",
		BaseDomain: "example.com",
	})
	err := cfg.Validate()
	if !errors.Is(err, arize.ErrMultipleEndpointOverrides) {
		t.Fatalf("expected ErrMultipleEndpointOverrides via errors.Is, got %v", err)
	}
	msg := err.Error()
	for _, want := range []string{"Region=", "SingleHost=", "BaseDomain="} {
		if !strings.Contains(msg, want) {
			t.Errorf("multi-override error should name conflicting field %q, got: %s", want, msg)
		}
	}
}

func TestConfig_String_MasksAPIKey(t *testing.T) {
	cfg := arize.Config{APIKey: "sk-test-deadbeef"}
	out := fmt.Sprintf("%v", cfg)
	if strings.Contains(out, "sk-test-deadbeef") {
		t.Errorf("String() must not leak full APIKey, got: %s", out)
	}
	if !strings.Contains(out, "sk-tes***") {
		t.Errorf("String() should show first 6 chars + ***, got: %s", out)
	}
}

func TestConfig_String_NoSecretWith_PercentPlusV(t *testing.T) {
	cfg := arize.Config{APIKey: "sk-test-deadbeef"}
	out := fmt.Sprintf("%+v", cfg)
	if strings.Contains(out, "sk-test-deadbeef") {
		t.Errorf("%%+v must not leak full APIKey (Stringer takes precedence), got: %s", out)
	}
}

func TestConfig_String_EmptyAPIKey(t *testing.T) {
	cfg := arize.Config{}
	out := fmt.Sprintf("%v", cfg)
	if strings.Contains(out, "***") {
		t.Errorf("empty APIKey should not be masked as ***, got: %s", out)
	}
}
