package prerelease

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"
)

func resetForTest() {
	warned.Range(func(k, _ any) bool {
		warned.Delete(k)
		return true
	})
}

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })
	return &buf
}

func TestWarn(t *testing.T) {
	tests := []struct {
		name          string
		calls         []struct{ key string; stage Stage }
		wantContains  []string
		wantNotChange bool
	}{
		{
			name: "once on first call",
			calls: []struct{ key string; stage Stage }{
				{key: "datasets.list", stage: Beta},
			},
			wantContains: []string{
				"datasets.list",
				"[BETA]",
				"v" + sdkconfig.SDKVersion,
			},
			wantNotChange: false,
		},
		{
			name: "suppresses repeated calls",
			calls: []struct{ key string; stage Stage }{
				{key: "datasets.list", stage: Beta},
				{key: "datasets.list", stage: Beta},
			},
			wantNotChange: true,
			wantContains:  nil,
		},
		{
			name: "distinct keys warn independently",
			calls: []struct{ key string; stage Stage }{
				{key: "datasets.list", stage: Beta},
				{key: "datasets.create", stage: Alpha},
			},
			wantContains: []string{
				"datasets.list",
				"datasets.create",
				"[ALPHA]",
				"[BETA]",
			},
			wantNotChange: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetForTest()
			buf := captureLogs(t)
			if tt.wantNotChange {
				Warn(tt.calls[0].key, tt.calls[0].stage)
				first := buf.String()
				Warn(tt.calls[1].key, tt.calls[1].stage)
				if buf.String() != first {
					t.Errorf("expected no additional log on repeat, got: %s", buf.String())
				}
				return
			}
			for _, c := range tt.calls {
				Warn(c.key, c.stage)
			}
			out := buf.String()
			for _, want := range tt.wantContains {
				if !strings.Contains(out, want) {
					t.Errorf("log missing %q, got: %s", want, out)
				}
			}
		})
	}
}

func TestFormatMessage_Articles(t *testing.T) {
	tests := []struct {
		name         string
		stage        Stage
		wantContains string
	}{
		{"Alpha uses 'an'", Alpha, "an alpha"},
		{"Beta uses 'a'", Beta, "a beta"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if msg := formatMessage("k", tt.stage); !strings.Contains(msg, tt.wantContains) {
				t.Errorf("want %q in message, got: %s", tt.wantContains, msg)
			}
		})
	}
}
