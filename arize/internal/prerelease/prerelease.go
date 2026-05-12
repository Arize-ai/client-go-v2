// Package prerelease emits a one-time warning the first time a caller uses
// an alpha/beta endpoint.
package prerelease

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"

	"github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"
)

// Stage indicates the release stage of an endpoint.
type Stage string

const (
	Alpha Stage = "alpha"
	Beta  Stage = "beta"
)

var warned sync.Map

// Warn emits a one-time slog.Warn for the given key. Subsequent calls with
// the same key are no-ops.
func Warn(key string, stage Stage) {
	if _, loaded := warned.LoadOrStore(key, struct{}{}); loaded {
		return
	}
	slog.Warn(formatMessage(key, stage))
}

func formatMessage(key string, stage Stage) string {
	article := "a"
	if stage == Alpha {
		article = "an"
	}
	return fmt.Sprintf(
		"[%s] %s is %s %s API in Arize SDK v%s and may change without notice. "+
			"If you experience unexpected failures, please upgrade to the most recent version of the package.",
		strings.ToUpper(string(stage)), key, article, string(stage), sdkconfig.SDKVersion,
	)
}
