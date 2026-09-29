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
	annotateSpans(ctx, client, project, space)
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
		ExcludedColumns: []string{
			"attributes.embedding.vectors",
		},
		Limit: 50,
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

// deleteSpans removes a batch of spans by ID. The response always contains
// Completed, DeletedSpanIds, and NotDeletedSpanIds. When Completed is false,
// the server could not fully process all data — retry the original full
// request (the delete is idempotent).
func deleteSpans(ctx context.Context, client *arize.Client, project, space string, spanIDs []string) {
	result, err := client.Spans.Delete(ctx, spans.DeleteRequest{
		Project: project,
		Space:   space,
		SpanIDs: spanIDs,
	})
	if err != nil {
		log.Fatalf("delete spans: %v", err)
	}
	fmt.Printf("deleted %d span(s): %v\n", len(result.DeletedSpanIds), result.DeletedSpanIds)
	if len(result.NotDeletedSpanIds) > 0 {
		fmt.Printf("not deleted %d span(s): %v\n", len(result.NotDeletedSpanIds), result.NotDeletedSpanIds)
	}
	if !result.Completed {
		fmt.Println("server could not fully process all data — retry the original full request")
	}
}

// annotateSpans writes human annotations to a batch of records. Granularity
// selects what each RecordId identifies: a span (GranularitySPAN, the
// default), a trace's root span (GranularityTRACE), or a session
// (GranularitySESSION, written to the root span of the session's earliest
// trace). Up to 1000 records may be annotated per request for
// GranularitySPAN/GranularityTRACE; up to 100 for GranularitySESSION.
func annotateSpans(ctx context.Context, client *arize.Client, project, space string) {
	label := "correct"
	if err := client.Spans.Annotate(ctx, spans.AnnotateRequest{
		Project: project,
		Space:   space,
		Annotations: []spans.AnnotateRecordInput{
			{
				RecordId: "span-1",
				Values: []spans.AnnotationInput{
					{Name: "Correctness", Label: &label},
				},
			},
		},
	}); err != nil {
		log.Fatalf("annotate spans: %v", err)
	}
	fmt.Println("annotated span-1")

	// Session-granularity example. Start/End are optional — when omitted,
	// they default to a 7-day window (vs. 31 days for SPAN/TRACE) ending
	// now. Pass them explicitly when annotating an older session.
	sessionLabel := "good"
	if err := client.Spans.Annotate(ctx, spans.AnnotateRequest{
		Project:     project,
		Space:       space,
		Granularity: spans.GranularitySESSION,
		Start:       time.Now().Add(-7 * 24 * time.Hour),
		End:         time.Now(),
		Annotations: []spans.AnnotateRecordInput{
			{
				RecordId: "your-session-id",
				Values: []spans.AnnotationInput{
					{Name: "quality", Label: &sessionLabel},
				},
			},
		},
	}); err != nil {
		log.Fatalf("annotate session: %v", err)
	}
	fmt.Println("annotated session your-session-id")
}
