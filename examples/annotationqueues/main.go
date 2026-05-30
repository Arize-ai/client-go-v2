// Package main demonstrates how to use the annotationqueues subclient of the
// Arize Go SDK v2.
//
// Annotation queues are an Alpha feature; every call prints a one-time
// pre-release warning.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/annotationqueues
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/annotationqueues"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/update/delete.
	// Space accepts either a space name or ID; the others are strict IDs.
	const (
		space          = "U3BhY2U6MTox"
		queueName      = "example-queue"
		projectID      = "<project-id>"
		annotationCfg  = "<annotation-config-id>"
		annotatorEmail = "annotator@example.com"
	)

	listQueues(ctx, client)

	q := createQueue(ctx, client, queueName, space, annotationCfg, annotatorEmail)
	getQueue(ctx, client, queueName, space)
	addRecords(ctx, client, queueName, space, projectID)
	listRecords(ctx, client, queueName, space)
	annotateAndAssign(ctx, client, queueName, space, annotatorEmail)
	updateQueue(ctx, client, queueName, space)
	deleteQueue(ctx, client, q.Id)
}

func listQueues(ctx context.Context, client *arize.Client) {
	resp, err := client.AnnotationQueues.List(ctx, annotationqueues.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list annotation queues: %v", err)
	}
	for _, q := range resp.AnnotationQueues {
		fmt.Printf("  %s\t%s\n", q.Id, q.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createQueue creates a queue with one annotator and instructions. Space accepts
// a name or ID and is resolved internally to a space ID.
func createQueue(ctx context.Context, client *arize.Client, name, space, configID, annotatorEmail string) *annotationqueues.AnnotationQueue {
	q, err := client.AnnotationQueues.Create(ctx, annotationqueues.CreateRequest{
		Space:               space,
		Name:                name,
		Instructions:        "Rate each response for helpfulness.",
		AnnotationConfigIDs: []string{configID},
		AnnotatorEmails:     []annotationqueues.Email{annotationqueues.Email(annotatorEmail)},
		AssignmentMethod:    annotationqueues.AssignmentMethodAll,
	})
	if err != nil {
		log.Fatalf("create annotation queue: %v", err)
	}
	fmt.Printf("created annotation queue %s (%s)\n", q.Name, q.Id)
	return q
}

// getQueue accepts a queue name or ID. Space is required when the queue is a name.
func getQueue(ctx context.Context, client *arize.Client, queue, space string) {
	q, err := client.AnnotationQueues.Get(
		ctx,
		annotationqueues.GetRequest{
			AnnotationQueue: queue,
			Space:           space,
		},
	)
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("annotation queue %q not found in space %q\n", queue, space)
			return
		}
		log.Fatalf("get annotation queue: %v", err)
	}
	fmt.Printf("found annotation queue %s (%s) with %d annotator(s)\n", q.Name, q.Id, len(q.Annotators))
}

// addRecords adds spans from a project (over the last 24 hours) to the queue.
// Build a record source with NewSpanRecordSource or NewExampleRecordSource.
func addRecords(ctx context.Context, client *arize.Client, queue, space, projectID string) {
	now := time.Now()
	src, err := annotationqueues.NewSpanRecordSource(
		annotationqueues.AnnotationQueueSpanRecordInput{
			ProjectId: projectID,
			StartTime: now.Add(-24 * time.Hour),
			EndTime:   now,
		},
	)
	if err != nil {
		log.Fatalf("build span record source: %v", err)
	}
	resp, err := client.AnnotationQueues.AddRecords(
		ctx,
		annotationqueues.AddRecordsRequest{
			AnnotationQueue: queue,
			Space:           space,
			RecordSources:   []annotationqueues.AnnotationQueueRecordInput{src},
		},
	)
	if err != nil {
		log.Fatalf("add records: %v", err)
	}
	fmt.Printf("added %d record(s) to queue %q\n", len(resp.RecordSources), queue)
}

func listRecords(ctx context.Context, client *arize.Client, queue, space string) {
	resp, err := client.AnnotationQueues.ListRecords(ctx, annotationqueues.ListRecordsRequest{
		AnnotationQueue: queue,
		Space:           space,
		Limit:           50,
	})
	if err != nil {
		log.Fatalf("list records: %v", err)
	}
	fmt.Printf("queue %q has %d record(s) on this page\n", queue, len(resp.Records))
}

// annotateAndAssign submits an annotation on the queue's first record and assigns
// a user to it. RecordID is a strict ID with no name resolution, so we read it
// from ListRecords first.
func annotateAndAssign(ctx context.Context, client *arize.Client, queue, space, annotatorEmail string) {
	resp, err := client.AnnotationQueues.ListRecords(ctx, annotationqueues.ListRecordsRequest{
		AnnotationQueue: queue,
		Space:           space,
		Limit:           1,
	})
	if err != nil {
		log.Fatalf("list records for annotation: %v", err)
	}
	if len(resp.Records) == 0 {
		fmt.Println("no records to annotate yet (span ingestion may be in progress)")
		return
	}
	recordID := resp.Records[0].Id

	score := 0.9
	if _, err := client.AnnotationQueues.Annotate(ctx, annotationqueues.AnnotateRequest{
		AnnotationQueue: queue,
		Space:           space,
		RecordID:        recordID,
		Annotations:     []annotationqueues.AnnotationInput{{Name: "helpfulness", Score: &score}},
	}); err != nil {
		log.Fatalf("annotate record: %v", err)
	}
	fmt.Printf("annotated record %s in queue %q\n", recordID, queue)

	if _, err := client.AnnotationQueues.Assign(ctx, annotationqueues.AssignRequest{
		AnnotationQueue:    queue,
		Space:              space,
		RecordID:           recordID,
		AssignedUserEmails: []annotationqueues.Email{annotationqueues.Email(annotatorEmail)},
	}); err != nil {
		log.Fatalf("assign record: %v", err)
	}
	fmt.Printf("assigned %s to record %s\n", annotatorEmail, recordID)

	// Records can also be removed by ID.
	if err := client.AnnotationQueues.DeleteRecords(ctx, annotationqueues.DeleteRecordsRequest{
		AnnotationQueue: queue,
		Space:           space,
		RecordIDs:       []string{recordID},
	}); err != nil {
		log.Fatalf("delete records: %v", err)
	}
	fmt.Printf("removed record %s from queue %q\n", recordID, queue)
}

// updateQueue patches the instructions; nil fields are left unchanged.
func updateQueue(ctx context.Context, client *arize.Client, queue, space string) {
	instructions := "Updated: rate each response for helpfulness and correctness."
	q, err := client.AnnotationQueues.Update(ctx, annotationqueues.UpdateRequest{
		AnnotationQueue: queue,
		Space:           space,
		Instructions:    &instructions,
	})
	if err != nil {
		log.Fatalf("update annotation queue: %v", err)
	}
	fmt.Printf("updated annotation queue %s (%s)\n", q.Name, q.Id)
}

func deleteQueue(ctx context.Context, client *arize.Client, queueID string) {
	if err := client.AnnotationQueues.Delete(ctx, annotationqueues.DeleteRequest{AnnotationQueue: queueID}); err != nil {
		log.Fatalf("delete annotation queue: %v", err)
	}
	fmt.Printf("deleted annotation queue %s\n", queueID)
}
