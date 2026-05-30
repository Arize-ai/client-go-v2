// Package main demonstrates how to use the evaluators subclient of the Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/evaluators
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/evaluators"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit these to match your account before running create/get/delete.
	// Space accepts either a space name or ID; aiIntegration is the AI
	// integration the template evaluator invokes.
	const (
		space         = "U3BhY2U6MTox"
		aiIntegration = "YWlfaW50ZWdyYXRpb246MTox"
		evaluatorName = "example-evaluator"
	)

	listEvaluators(ctx, client)

	ev := createEvaluator(ctx, client, evaluatorName, space, aiIntegration)
	getEvaluator(ctx, client, evaluatorName, space)
	listVersions(ctx, client, evaluatorName, space)
	addVersion(ctx, client, evaluatorName, space, aiIntegration)
	renameEvaluator(ctx, client, evaluatorName, space)
	deleteEvaluator(ctx, client, ev.Id)
}

func listEvaluators(ctx context.Context, client *arize.Client) {
	resp, err := client.Evaluators.List(ctx, evaluators.ListRequest{Limit: 25})
	if err != nil {
		log.Fatalf("list evaluators: %v", err)
	}
	for _, e := range resp.Evaluators {
		fmt.Printf("  %s\t%s\t%s\n", e.Id, e.Name, e.Type)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// createEvaluator creates a template (LLM-based) evaluator with its initial
// version. Space accepts either a space name or ID; it is resolved internally
// to a space ID. The evaluator's type is derived from Version (Template -> template).
func createEvaluator(ctx context.Context, client *arize.Client, name, space, aiIntegration string) *evaluators.EvaluatorWithVersion {
	ev, err := client.Evaluators.Create(ctx, evaluators.CreateRequest{
		Name:        name,
		Space:       space,
		Description: "scores answer relevance",
		Version: evaluators.VersionConfig{
			CommitMessage: "initial version",
			Template: &evaluators.TemplateConfig{
				Name:     "relevance",
				Template: "Is the answer relevant to the question?\n{{input}}",
				LlmConfig: evaluators.EvaluatorLlmConfig{
					AiIntegrationId: aiIntegration,
					ModelName:       "gpt-4o",
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create evaluator: %v", err)
	}
	fmt.Printf("created evaluator %s (%s)\n", ev.Name, ev.Id)
	return ev
}

// getEvaluator accepts an evaluator name or ID. Space is required when the
// evaluator is a name. The returned version is a oneOf — read the active
// variant with evaluators.AsTemplate / evaluators.AsCode.
func getEvaluator(ctx context.Context, client *arize.Client, evaluator, space string) {
	ev, err := client.Evaluators.Get(ctx, evaluators.GetRequest{Evaluator: evaluator, Space: space})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("evaluator %q not found in space %q\n", evaluator, space)
			return
		}
		log.Fatalf("get evaluator: %v", err)
	}
	fmt.Printf("found evaluator %s (%s), type %s\n", ev.Name, ev.Id, ev.Type)
	if tmpl, ok := evaluators.AsTemplate(ev.Version); ok {
		fmt.Printf("  latest template version %s: %q\n", tmpl.Id, tmpl.TemplateConfig.Template)
	}
}

// listVersions lists an evaluator's versions, resolved by name or ID.
func listVersions(ctx context.Context, client *arize.Client, evaluator, space string) {
	resp, err := client.Evaluators.ListVersions(ctx, evaluators.ListVersionsRequest{
		Evaluator: evaluator,
		Space:     space,
		Limit:     25,
	})
	if err != nil {
		log.Fatalf("list versions: %v", err)
	}
	fmt.Printf("evaluator %q has %d version(s) on this page\n", evaluator, len(resp.EvaluatorVersions))
}

// addVersion appends a new template version to an existing evaluator, resolved
// by name. The new version's kind must match the parent evaluator's type. It
// then fetches the new version by its ID.
func addVersion(ctx context.Context, client *arize.Client, evaluator, space, aiIntegration string) {
	v, err := client.Evaluators.CreateVersion(ctx, evaluators.CreateVersionRequest{
		Evaluator: evaluator,
		Space:     space,
		Version: evaluators.VersionConfig{
			CommitMessage: "tighten the rubric",
			Template: &evaluators.TemplateConfig{
				Name:     "relevance",
				Template: "Rate 0-1 how relevant the answer is.\n{{input}}",
				LlmConfig: evaluators.EvaluatorLlmConfig{
					AiIntegrationId: aiIntegration,
					ModelName:       "gpt-4o",
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("create version: %v", err)
	}
	tmpl, ok := evaluators.AsTemplate(*v)
	if !ok {
		log.Fatal("expected a template version")
	}
	fmt.Printf("created version %s\n", tmpl.Id)

	got, err := client.Evaluators.GetVersion(ctx, evaluators.GetVersionRequest{VersionID: tmpl.Id})
	if err != nil {
		log.Fatalf("get version: %v", err)
	}
	if v, ok := evaluators.AsTemplate(*got); ok {
		fmt.Printf("fetched version %s: %q\n", v.Id, v.TemplateConfig.Template)
	}
}

// renameEvaluator updates the evaluator's name. Only non-nil patch fields are
// sent; Space is required when the evaluator is a name.
func renameEvaluator(ctx context.Context, client *arize.Client, evaluator, space string) {
	newName := evaluator + "-renamed"
	ev, err := client.Evaluators.Update(ctx, evaluators.UpdateRequest{
		Evaluator: evaluator,
		Space:     space,
		Name:      &newName,
	})
	if err != nil {
		log.Fatalf("rename evaluator: %v", err)
	}
	fmt.Printf("renamed evaluator to %s (%s)\n", ev.Name, ev.Id)
}

func deleteEvaluator(ctx context.Context, client *arize.Client, evaluatorID string) {
	if err := client.Evaluators.Delete(ctx, evaluators.DeleteRequest{Evaluator: evaluatorID}); err != nil {
		log.Fatalf("delete evaluator: %v", err)
	}
	fmt.Printf("deleted evaluator %s\n", evaluatorID)
}
