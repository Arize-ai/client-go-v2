package arize

import "github.com/Arize-ai/client-go-v2/arize/internal/sdkconfig"

// Config holds all configuration for the Arize client. The canonical type
// lives in arize/internal/sdkconfig so subclients can import it without
// creating a cycle through the public arize package.
type Config = sdkconfig.Config
