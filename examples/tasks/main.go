// Package main demonstrates how to use the tasks subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/tasks
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/tasks"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/trigger/delete.
	// Space, project, and dataset accept either a name or an ID; evaluator and
	// aiIntegration are IDs.
	const (
		space         = "U3BhY2U6MTox"
		project       = "example-project"
		dataset       = "example-dataset"
		evaluator     = "ZXZhbHVhdG9yOjE6MQ=="
		aiIntegration = "YWlfaW50ZWdyYXRpb246MTox"
		evalTaskName  = "example-eval-task"
		expTaskName   = "example-experiment-task"
	)

	listTasks(ctx, client, space)

	evalTask := createEvaluationTask(ctx, client, evalTaskName, project, space, evaluator)
	getTask(ctx, client, evalTaskName, space)
	updateTask(ctx, client, evalTaskName, space)
	run := triggerEvaluationRun(ctx, client, evalTaskName, space)
	waitForRun(ctx, client, run.Id)
	listRuns(ctx, client, evalTaskName, space)
	deleteTask(ctx, client, evalTask.Id)

	expTask := createRunExperimentTask(ctx, client, expTaskName, dataset, space, aiIntegration)
	triggerExperimentRun(ctx, client, expTaskName, space)
	deleteTask(ctx, client, expTask.Id)
}

func listTasks(ctx context.Context, client *arize.Client, space string) {
	resp, err := client.Tasks.List(ctx, tasks.ListRequest{Space: space, Limit: 25})
	if err != nil {
		log.Fatalf("list tasks: %v", err)
	}
	for _, tk := range resp.Tasks {
		fmt.Printf("  %s\t%s\t%s\n", tk.Id, tk.Name, tk.Type)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createEvaluationTask creates a project-based template_evaluation task that
// runs the given evaluator over incoming spans. Project accepts either a
// project name or ID; Space is required when it is a name.
func createEvaluationTask(ctx context.Context, client *arize.Client, name, project, space, evaluator string) *tasks.Task {
	tk, err := client.Tasks.CreateEvaluationTask(ctx, tasks.CreateEvaluationTaskRequest{
		Name:    name,
		Type:    tasks.TaskTypeTemplateEvaluation,
		Project: project,
		Space:   space,
		Evaluators: []tasks.EvaluatorInput{{
			EvaluatorID: evaluator,
		}},
		SamplingRate: 0.5,
		QueryFilter:  "attributes.openinference.span.kind = 'LLM'",
	})
	if err != nil {
		log.Fatalf("create evaluation task: %v", err)
	}
	fmt.Printf("created task %s (%s)\n", tk.Name, tk.Id)
	return tk
}

// createRunExperimentTask creates a run_experiment task whose runs execute an
// experiment over the dataset. RunConfiguration is a oneOf — populate exactly
// one variant via FromLlmGenerationRunConfig / FromTemplateEvaluationRunConfig.
func createRunExperimentTask(ctx context.Context, client *arize.Client, name, dataset, space, aiIntegration string) *tasks.Task {
	var rc tasks.RunConfiguration
	if err := rc.FromTemplateEvaluationRunConfig(tasks.TemplateEvaluationRunConfig{
		AiIntegrationId:    aiIntegration,
		Template:           "Is the answer relevant to the question?\n{{input}}",
		ProvideExplanation: true,
	}); err != nil {
		log.Fatalf("build run configuration: %v", err)
	}
	tk, err := client.Tasks.CreateRunExperimentTask(ctx, tasks.CreateRunExperimentTaskRequest{
		Name:             name,
		Dataset:          dataset,
		Space:            space,
		RunConfiguration: rc,
	})
	if err != nil {
		log.Fatalf("create run_experiment task: %v", err)
	}
	fmt.Printf("created task %s (%s)\n", tk.Name, tk.Id)
	return tk
}

// getTask accepts a task name or ID. Space is required when the task is a
// name.
func getTask(ctx context.Context, client *arize.Client, task, space string) {
	tk, err := client.Tasks.Get(ctx, tasks.GetRequest{Task: task, Space: space})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("task %q not found in space %q\n", task, space)
			return
		}
		log.Fatalf("get task: %v", err)
	}
	fmt.Printf("found task %s (%s), type %s, %d evaluator(s)\n", tk.Name, tk.Id, tk.Type, len(tk.Evaluators))
}

// updateTask patches an evaluation task. Only non-nil patch fields are sent;
// the SDK fetches the task first to validate the fields against its type.
// Passing a pointer to an empty string clears the query filter.
func updateTask(ctx context.Context, client *arize.Client, task, space string) {
	rate := float32(0.25)
	clearFilter := ""
	tk, err := client.Tasks.Update(ctx, tasks.UpdateRequest{
		Task:         task,
		Space:        space,
		SamplingRate: &rate,
		QueryFilter:  &clearFilter,
	})
	if err != nil {
		log.Fatalf("update task: %v", err)
	}
	fmt.Printf("updated task %s (sampling rate now %v)\n", tk.Name, tk.SamplingRate)
}

// triggerEvaluationRun starts an async run of an evaluation task over the
// last hour of data. The returned run starts in pending status.
func triggerEvaluationRun(ctx context.Context, client *arize.Client, task, space string) *tasks.TaskRun {
	run, err := client.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
		Task:          task,
		Space:         space,
		DataStartTime: time.Now().Add(-time.Hour),
		DataEndTime:   time.Now(),
		MaxSpans:      1000,
	})
	if err != nil {
		log.Fatalf("trigger run: %v", err)
	}
	fmt.Printf("triggered run %s (status %s)\n", run.Id, run.Status)
	return run
}

// triggerExperimentRun starts an async run of a run_experiment task.
// ExperimentName is required and must be unique within the dataset.
func triggerExperimentRun(ctx context.Context, client *arize.Client, task, space string) {
	run, err := client.Tasks.TriggerRun(ctx, tasks.TriggerRunRequest{
		Task:           task,
		Space:          space,
		ExperimentName: fmt.Sprintf("example-run-%d", time.Now().Unix()),
		MaxExamples:    10,
	})
	if err != nil {
		log.Fatalf("trigger experiment run: %v", err)
	}
	fmt.Printf("triggered experiment run %s (status %s)\n", run.Id, run.Status)
}

// waitForRun polls the run until it reaches a terminal state (completed,
// failed, or cancelled), then prints its statistics.
func waitForRun(ctx context.Context, client *arize.Client, runID string) {
	run, err := client.Tasks.WaitForRun(ctx, tasks.WaitForRunRequest{
		RunID:        runID,
		PollInterval: 5 * time.Second,
		Timeout:      5 * time.Minute,
	})
	if err != nil {
		if errors.Is(err, tasks.ErrWaitTimeout) {
			fmt.Printf("run %s still going — cancel it or keep waiting\n", runID)
			return
		}
		log.Fatalf("wait for run: %v", err)
	}
	fmt.Printf("run %s finished with status %s: %d ok, %d errored, %d skipped\n",
		run.Id, run.Status, run.NumSuccesses, run.NumErrors, run.NumSkipped)
}

func listRuns(ctx context.Context, client *arize.Client, task, space string) {
	resp, err := client.Tasks.ListRuns(ctx, tasks.ListRunsRequest{Task: task, Space: space, Limit: 25})
	if err != nil {
		log.Fatalf("list runs: %v", err)
	}
	fmt.Printf("task %q has %d run(s) on this page\n", task, len(resp.TaskRuns))
}

func deleteTask(ctx context.Context, client *arize.Client, taskID string) {
	if err := client.Tasks.Delete(ctx, tasks.DeleteRequest{Task: taskID}); err != nil {
		log.Fatalf("delete task: %v", err)
	}
	fmt.Printf("deleted task %s\n", taskID)
}
