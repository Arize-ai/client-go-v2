// Package main demonstrates how to use the traces subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/traces
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/traces"
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

	listTraces(ctx, client, project, space)
}

// listTraces flattens the body half (project, time range, filter) and the
// query-params half (limit, cursor) of the underlying POST into a single
// ListRequest. traces.List uses POST because the filter DSL can be too large
// for a query string. A trace is returned when any of its spans matches the
// filter; each trace carries a flat list of spans to reconstruct client-side.
func listTraces(ctx context.Context, client *arize.Client, project, space string) {
	resp, err := client.Traces.List(ctx, traces.ListRequest{
		Project: project,
		Space:   space,
		End:     time.Now(),
		Filter:  "status_code = 'ERROR'",
		Limit:   50,
	})
	if err != nil {
		log.Fatalf("list traces: %v", err)
	}
	for _, tr := range resp.Traces {
		fmt.Printf("  trace %s (root %s, %d span(s), truncated=%t)\n",
			tr.TraceId, tr.RootSpanId, len(tr.Spans), tr.SpansTruncated)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}
