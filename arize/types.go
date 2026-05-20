package arize

import "github.com/Arize-ai/client-go-v2/arize/internal/generated"

// PaginationMetadata is the cursor-pagination shape returned by list endpoints.
// Advance to the next page by passing NextCursor as the Cursor field in ListRequest.
type PaginationMetadata = generated.PaginationMetadata
