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

func TestWarn_OnceFirstCall(t *testing.T) {
	resetForTest()
	buf := captureLogs(t)
	Warn("datasets.list", Beta)
	if !strings.Contains(buf.String(), "datasets.list") {
		t.Errorf("expected key in log, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "[BETA]") {
		t.Errorf("expected [BETA] tag, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "v"+sdkconfig.SDKVersion) {
		t.Errorf("expected SDK version, got: %s", buf.String())
	}
}

func TestWarn_SuppressesRepeatedCalls(t *testing.T) {
	resetForTest()
	buf := captureLogs(t)
	Warn("datasets.list", Beta)
	first := buf.String()
	Warn("datasets.list", Beta)
	if buf.String() != first {
		t.Errorf("expected no additional log on repeat, got: %s", buf.String())
	}
}

func TestWarn_DistinctKeysWarnIndependently(t *testing.T) {
	resetForTest()
	buf := captureLogs(t)
	Warn("datasets.list", Beta)
	Warn("datasets.create", Alpha)
	out := buf.String()
	if !strings.Contains(out, "datasets.list") || !strings.Contains(out, "datasets.create") {
		t.Errorf("expected both keys logged, got: %s", out)
	}
	if !strings.Contains(out, "[ALPHA]") || !strings.Contains(out, "[BETA]") {
		t.Errorf("expected both stages logged, got: %s", out)
	}
}

func TestFormatMessage_Articles(t *testing.T) {
	if msg := formatMessage("k", Alpha); !strings.Contains(msg, "an alpha") {
		t.Errorf("expected 'an alpha', got: %s", msg)
	}
	if msg := formatMessage("k", Beta); !strings.Contains(msg, "a beta") {
		t.Errorf("expected 'a beta', got: %s", msg)
	}
}
