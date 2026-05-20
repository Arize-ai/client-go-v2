// Package main demonstrates how to use the spans subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/spans
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/spans"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Either a project name (with a space name/ID) or a base64 project ID works.
	const (
		project = "my-project"
		space   = "my-space"
	)

	listSpans(ctx, client, project, space)
	deleteSpans(ctx, client, project, space, []string{"span-1", "span-2"})
}

// listSpans flattens the body half (project, time range, filter) and the
// query-params half (limit, cursor) of the underlying POST into a single
// ListRequest. spans.List uses POST because the filter DSL can be too large
// for a query string.
func listSpans(ctx context.Context, client *arize.Client, project, space string) {
	resp, err := client.Spans.List(ctx, spans.ListRequest{
		Project: project,
		Space:   space,
		End:     time.Now(),
		Filter:  "status_code = 'ERROR'",
		Limit:   50,
	})
	if err != nil {
		log.Fatalf("list spans: %v", err)
	}
	for _, s := range resp.Spans {
		fmt.Printf("  %s\t%s\n", s.Context.SpanId, s.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// deleteSpans removes a batch of spans by ID. The server may return a partial
// success (HTTP 200 with a SpanDeletePartial listing the IDs actually
// deleted); a fully-successful delete returns HTTP 204 and a nil partial.
func deleteSpans(ctx context.Context, client *arize.Client, project, space string, spanIDs []string) {
	partial, err := client.Spans.Delete(ctx, spans.DeleteRequest{
		Project: project,
		Space:   space,
		SpanIDs: spanIDs,
	})
	if err != nil {
		log.Fatalf("delete spans: %v", err)
	}
	if partial != nil {
		fmt.Printf("partial delete — server deleted %d of %d spans: %v\n",
			len(partial.DeletedSpanIds), len(spanIDs), partial.DeletedSpanIds)
		return
	}
	fmt.Printf("deleted all %d spans\n", len(spanIDs))
}
