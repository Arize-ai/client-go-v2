// Package main demonstrates how to use the experiments subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/experiments
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/experiments"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/delete.
	// Space and dataset each accept either a name or an ID.
	const (
		space          = "U3BhY2U6MTox"
		datasetName    = "example-dataset"
		experimentName = "example-experiment"
	)

	listExperiments(ctx, client, datasetName, space)

	exp := createExperiment(ctx, client, experimentName, datasetName, space)
	getExperiment(ctx, client, experimentName, datasetName, space)
	appendRuns(ctx, client, exp.Id)
	listRuns(ctx, client, experimentName, datasetName, space)
	deleteExperiment(ctx, client, exp.Id)
}

// listExperiments lists experiments, optionally filtered to a single dataset.
// Dataset accepts a name or ID; Space is only used to resolve Dataset when it
// is a name.
func listExperiments(ctx context.Context, client *arize.Client, dataset, space string) {
	resp, err := client.Experiments.List(ctx, experiments.ListRequest{
		Dataset: dataset,
		Space:   space,
		Limit:   25,
	})
	if err != nil {
		log.Fatalf("list experiments: %v", err)
	}
	for _, e := range resp.Experiments {
		fmt.Printf("  %s\t%s\n", e.Id, e.Name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createExperiment creates an experiment from a small set of run records. Each
// run is an arbitrary map of user-named columns; TaskFields tells the SDK which
// columns hold the dataset example_id and the task output, and EvaluatorColumns
// maps an evaluator's result columns into the wire format. Dataset accepts a
// name or ID; Space is required when Dataset is a name.
func createExperiment(ctx context.Context, client *arize.Client, name, dataset, space string) *experiments.Experiment {
	exp, err := client.Experiments.Create(ctx, experiments.CreateRequest{
		Name:    name,
		Dataset: dataset,
		Space:   space,
		Runs: []map[string]any{
			{
				"example_id":  "example-1",
				"answer":      "An AI observability platform.",
				"quality":     0.9,
				"verdict":     "good",
				"explanation": "matches the reference answer",
			},
			{
				"example_id":  "example-2",
				"answer":      "A unit of work in a trace.",
				"quality":     0.7,
				"verdict":     "ok",
				"explanation": "mostly correct",
			},
		},
		TaskFields: experiments.TaskFields{ExampleID: "example_id", Output: "answer"},
		EvaluatorColumns: map[string]experiments.EvaluatorFields{
			"correctness": {
				Score:       "quality",
				Label:       "verdict",
				Explanation: "explanation",
			},
		},
	})
	if err != nil {
		log.Fatalf("create experiment: %v", err)
	}
	fmt.Printf("created experiment %s (%s)\n", exp.Name, exp.Id)
	return exp
}

// getExperiment accepts an experiment name or ID. Dataset is required when
// experiment is a name; Space is required when Dataset is also a name.
func getExperiment(ctx context.Context, client *arize.Client, experiment, dataset, space string) {
	exp, err := client.Experiments.Get(ctx, experiments.GetRequest{
		Experiment: experiment,
		Dataset:    dataset,
		Space:      space,
	})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("experiment %q not found\n", experiment)
			return
		}
		log.Fatalf("get experiment: %v", err)
	}
	fmt.Printf("found experiment %s (%s)\n", exp.Name, exp.Id)
}

// appendRuns appends new runs to an existing experiment by ID.
func appendRuns(ctx context.Context, client *arize.Client, experimentID string) {
	result, err := client.Experiments.AppendRuns(ctx, experiments.AppendRunsRequest{
		ExperimentID: experimentID,
		ExperimentRuns: []experiments.ExperimentRunCreate{
			{ExampleId: "example-1", Output: "An AI observability platform."},
			{ExampleId: "example-2", Output: "A unit of work in a trace."},
		},
	})
	if err != nil {
		log.Fatalf("append runs: %v", err)
	}
	fmt.Printf("appended %d run(s) to experiment %s\n", len(result.RunIds), result.Id)
}

// listRuns lists the runs recorded for an experiment.
func listRuns(ctx context.Context, client *arize.Client, experiment, dataset, space string) {
	resp, err := client.Experiments.ListRuns(ctx, experiments.ListRunsRequest{
		Experiment: experiment,
		Dataset:    dataset,
		Space:      space,
		Limit:      50,
	})
	if err != nil {
		log.Fatalf("list runs: %v", err)
	}
	fmt.Printf("experiment %q has %d run(s) on this page\n", experiment, len(resp.ExperimentRuns))
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages available)")
	}
}

func deleteExperiment(ctx context.Context, client *arize.Client, experimentID string) {
	if err := client.Experiments.Delete(ctx, experiments.DeleteRequest{Experiment: experimentID}); err != nil {
		log.Fatalf("delete experiment: %v", err)
	}
	fmt.Printf("deleted experiment %s\n", experimentID)
}
