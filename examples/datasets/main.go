// Package main demonstrates how to use the datasets subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/datasets
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/datasets"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/delete.
	// Space accepts either a space name or ID.
	const (
		space       = "U3BhY2U6MTox"
		datasetName = "example-dataset"
	)

	listDatasets(ctx, client)

	ds := createDataset(ctx, client, datasetName, space)
	getDataset(ctx, client, datasetName, space)
	listExamples(ctx, client, datasetName, space)
	appendExamples(ctx, client, datasetName, space)
	annotateExamples(ctx, client, datasetName, space)
	deleteExamples(ctx, client, datasetName, space, "<dataset-version-id>", []string{"<example-id>"})
	renameDataset(ctx, client, datasetName, space)
	deleteDataset(ctx, client, ds.Id)
}

func listDatasets(ctx context.Context, client *arize.Client) {
	resp, err := client.Datasets.List(ctx, datasets.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list datasets: %v", err)
	}
	for _, d := range resp.Datasets {
		fmt.Printf("  %s\t%s\n", d.Id, d.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getDataset accepts a dataset name or ID. Space is required when dataset is
// a name.
func getDataset(ctx context.Context, client *arize.Client, dataset, space string) {
	ds, err := client.Datasets.Get(ctx, datasets.GetRequest{Dataset: dataset, Space: space})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("dataset %q not found in space %q\n", dataset, space)
			return
		}
		log.Fatalf("get dataset: %v", err)
	}
	fmt.Printf("found dataset %s (%s)\n", ds.Name, ds.Id)
}

// createDataset creates a dataset with two initial examples. Space accepts
// either a space name or ID; it is resolved internally to a space ID. Each
// example is an arbitrary set of user-defined fields.
func createDataset(ctx context.Context, client *arize.Client, name, space string) *datasets.Dataset {
	ds, err := client.Datasets.Create(ctx, datasets.CreateRequest{
		Name:  name,
		Space: space,
		Examples: []datasets.CreateDatasetExampleInput{
			{"input": "What is Arize?", "output": "An AI observability platform."},
			{"input": "What is a span?", "output": "A unit of work in a trace."},
		},
	})
	if err != nil {
		log.Fatalf("create dataset: %v", err)
	}
	fmt.Printf("created dataset %s (%s)\n", ds.Name, ds.Id)
	return ds
}

// listExamples accepts a dataset name or ID. Space is required when dataset is
// a name.
func listExamples(ctx context.Context, client *arize.Client, dataset, space string) {
	resp, err := client.Datasets.ListExamples(ctx, datasets.ListExamplesRequest{
		Dataset: dataset,
		Space:   space,
		Limit:   50,
	})
	if err != nil {
		log.Fatalf("list examples: %v", err)
	}
	fmt.Printf("dataset %q has %d example(s) on this page\n", dataset, len(resp.Examples))
}

// appendExamples appends an example to the dataset's latest version and prints
// the server-assigned IDs of the inserted examples.
func appendExamples(ctx context.Context, client *arize.Client, dataset, space string) {
	resp, err := client.Datasets.AppendExamples(ctx, datasets.AppendExamplesRequest{
		Dataset: dataset,
		Space:   space,
		Examples: []datasets.CreateDatasetExampleInput{
			{"input": "What is an evaluator?", "output": "A function that scores spans."},
		},
	})
	if err != nil {
		log.Fatalf("append examples: %v", err)
	}
	fmt.Printf("appended %d example(s) to dataset %q (version %s)\n",
		len(resp.ExampleIds), dataset, resp.DatasetVersionId)
}

// deleteExamples removes a batch of examples from a specific dataset version.
// The delete is partial-tolerant: the result reports which IDs were deleted and
// which were not, and a false Completed means the full request should be
// retried (the operation is idempotent).
func deleteExamples(ctx context.Context, client *arize.Client, dataset, space, datasetVersionID string, exampleIDs []string) {
	resp, err := client.Datasets.DeleteExamples(ctx, datasets.DeleteExamplesRequest{
		Dataset:          dataset,
		Space:            space,
		DatasetVersionID: datasetVersionID,
		ExampleIDs:       exampleIDs,
	})
	if err != nil {
		log.Fatalf("delete examples: %v", err)
	}
	fmt.Printf("deleted %d example(s) from dataset %q (completed=%t, not deleted: %d)\n",
		len(resp.DeletedExampleIds), dataset, resp.Completed, len(resp.NotDeletedExampleIds))
}

// annotateExamples writes a human annotation to the dataset's first example.
// Annotations are upserted by config name, so re-running this overwrites the
// previous value rather than creating duplicates.
func annotateExamples(ctx context.Context, client *arize.Client, dataset, space string) {
	resp, err := client.Datasets.ListExamples(ctx, datasets.ListExamplesRequest{
		Dataset: dataset,
		Space:   space,
		Limit:   1,
	})
	if err != nil {
		log.Fatalf("list examples for annotation: %v", err)
	}
	if len(resp.Examples) == 0 || resp.Examples[0].Id == nil {
		fmt.Println("no examples to annotate")
		return
	}
	score := 0.9
	err = client.Datasets.AnnotateExamples(ctx, datasets.AnnotateExamplesRequest{
		Dataset: dataset,
		Space:   space,
		Annotations: []datasets.AnnotateRecordInput{
			{
				RecordId: *resp.Examples[0].Id,
				Values: []datasets.AnnotationInput{
					{Name: "quality", Score: &score},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("annotate examples: %v", err)
	}
	fmt.Printf("annotated example %s in dataset %q\n", *resp.Examples[0].Id, dataset)
}

// renameDataset renames the dataset. Space is required when dataset is a name.
func renameDataset(ctx context.Context, client *arize.Client, dataset, space string) {
	ds, err := client.Datasets.Update(ctx, datasets.UpdateRequest{
		Dataset: dataset,
		Space:   space,
		Name:    dataset + "-renamed",
	})
	if err != nil {
		log.Fatalf("rename dataset: %v", err)
	}
	fmt.Printf("renamed dataset to %s (%s)\n", ds.Name, ds.Id)
}

func deleteDataset(ctx context.Context, client *arize.Client, datasetID string) {
	if err := client.Datasets.Delete(ctx, datasets.DeleteRequest{Dataset: datasetID}); err != nil {
		log.Fatalf("delete dataset: %v", err)
	}
	fmt.Printf("deleted dataset %s\n", datasetID)
}
