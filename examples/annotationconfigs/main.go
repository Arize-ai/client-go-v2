// Package main demonstrates how to use the annotationconfigs subclient of the
// Arize Go SDK v2.
//
// Run with: ARIZE_API_KEY=<key> go run ./examples/annotationconfigs
package main

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Arize-ai/client-go-v2/arize"
	"github.com/Arize-ai/client-go-v2/arize/annotationconfigs"
)

func main() {
	client, err := arize.NewClient(arize.Config{})
	if err != nil {
		log.Fatalf("create client: %v", err)
	}
	ctx := context.Background()

	// Edit to match your account before running create/get/delete.
	// Space accepts either a space name or ID.
	const space = "U3BhY2U6MTox"

	listAnnotationConfigs(ctx, client, space)

	categorical := createCategorical(ctx, client, space, "example-quality")
	continuous := createContinuous(ctx, client, space, "example-score")
	freeform := createFreeform(ctx, client, space, "example-notes")

	getAnnotationConfig(ctx, client, "example-quality", space)

	updateCategorical(ctx, client, idOf(categorical))
	updateContinuous(ctx, client, idOf(continuous))
	updateFreeform(ctx, client, idOf(freeform))

	deleteAnnotationConfig(ctx, client, idOf(categorical))
	deleteAnnotationConfig(ctx, client, idOf(continuous))
	deleteAnnotationConfig(ctx, client, idOf(freeform))
}

// listAnnotationConfigs lists configs in a space. AnnotationConfig is a
// discriminated union — call ValueByDiscriminator and type-switch over the
// variant to read variant-specific fields (see identify).
func listAnnotationConfigs(ctx context.Context, client *arize.Client, space string) {
	resp, err := client.AnnotationConfigs.List(ctx, annotationconfigs.ListRequest{
		Space: space,
		Limit: 25,
	})
	if err != nil {
		log.Fatalf("list annotation configs: %v", err)
	}
	for _, ac := range resp.AnnotationConfigs {
		id, kind, name := identify(&ac)
		fmt.Printf("  %s\t%s\t%s\n", id, kind, name)
	}
	if resp.Pagination.HasMore {
		fmt.Println("  (more pages — pass NextCursor as Cursor in the next ListRequest)")
	}
}

// getAnnotationConfig accepts an annotation config name or ID. Space is
// required when the config is identified by name.
func getAnnotationConfig(ctx context.Context, client *arize.Client, config, space string) {
	ac, err := client.AnnotationConfigs.Get(ctx, annotationconfigs.GetRequest{
		AnnotationConfig: config,
		Space:            space,
	})
	if err != nil {
		var nfe *arize.NotFoundError
		if errors.As(err, &nfe) {
			fmt.Printf("annotation config %q not found in space %q\n", config, space)
			return
		}
		log.Fatalf("get annotation config: %v", err)
	}
	id, _, name := identify(ac)
	fmt.Printf("found annotation config %s (%s)\n", name, id)
}

// createCategorical creates a categorical annotation config with a fixed set
// of allowed labels. OptimizationDirection tells the platform which direction
// is "better"; leave it zero to use the server default.
func createCategorical(ctx context.Context, client *arize.Client, space, name string) *annotationconfigs.AnnotationConfig {
	score1, score0 := 1.0, 0.0
	ac, err := client.AnnotationConfigs.CreateCategorical(ctx, annotationconfigs.CreateCategoricalRequest{
		Space:                 space,
		Name:                  name,
		OptimizationDirection: annotationconfigs.OptimizationDirectionMaximize,
		Values: []annotationconfigs.CategoricalAnnotationValue{
			{Label: "good", Score: &score1},
			{Label: "bad", Score: &score0},
		},
	})
	if err != nil {
		log.Fatalf("create categorical annotation config: %v", err)
	}
	id := idOf(ac)
	fmt.Printf("created categorical annotation config %s (%s)\n", name, id)
	return ac
}

// createContinuous creates a continuous annotation config bounded by
// MinimumScore and MaximumScore. Both are required for continuous configs.
func createContinuous(ctx context.Context, client *arize.Client, space, name string) *annotationconfigs.AnnotationConfig {
	ac, err := client.AnnotationConfigs.CreateContinuous(ctx, annotationconfigs.CreateContinuousRequest{
		Space:                 space,
		Name:                  name,
		OptimizationDirection: annotationconfigs.OptimizationDirectionMaximize,
		MinimumScore:          0.0,
		MaximumScore:          1.0,
	})
	if err != nil {
		log.Fatalf("create continuous annotation config: %v", err)
	}
	id := idOf(ac)
	fmt.Printf("created continuous annotation config %s (%s)\n", name, id)
	return ac
}

// createFreeform creates a freeform annotation config. Freeform configs take
// no extra fields beyond Name and Space.
func createFreeform(ctx context.Context, client *arize.Client, space, name string) *annotationconfigs.AnnotationConfig {
	ac, err := client.AnnotationConfigs.CreateFreeform(ctx, annotationconfigs.CreateFreeformRequest{
		Space: space,
		Name:  name,
	})
	if err != nil {
		log.Fatalf("create freeform annotation config: %v", err)
	}
	id := idOf(ac)
	fmt.Printf("created freeform annotation config %s (%s)\n", name, id)
	return ac
}

// updateCategorical patches an existing categorical annotation config.
// Leave a patch field nil to preserve its current value. Here we rename the
// config and replace its label set.
func updateCategorical(ctx context.Context, client *arize.Client, configID string) {
	name := "example-quality-v2"
	score1, score0 := 1.0, 0.0
	ac, err := client.AnnotationConfigs.UpdateCategorical(ctx, annotationconfigs.UpdateCategoricalRequest{
		AnnotationConfig: configID,
		Name:             &name,
		Values: &[]annotationconfigs.CategoricalAnnotationValue{
			{Label: "great", Score: &score1},
			{Label: "poor", Score: &score0},
		},
	})
	if err != nil {
		log.Fatalf("update annotation config: %v", err)
	}
	id, _, updatedName := identify(ac)
	fmt.Printf("updated categorical annotation config %s (%s)\n", updatedName, id)
}

// updateContinuous patches an existing continuous annotation config. Leave a
// patch field nil to preserve its current value. Here we widen the score
// range's upper bound.
func updateContinuous(ctx context.Context, client *arize.Client, configID string) {
	maxScore := 10.0
	ac, err := client.AnnotationConfigs.UpdateContinuous(ctx, annotationconfigs.UpdateContinuousRequest{
		AnnotationConfig: configID,
		MaximumScore:     &maxScore,
	})
	if err != nil {
		log.Fatalf("update continuous annotation config: %v", err)
	}
	id, _, updatedName := identify(ac)
	fmt.Printf("updated continuous annotation config %s (%s)\n", updatedName, id)
}

// updateFreeform patches an existing freeform annotation config. Freeform
// configs only support renaming; leave Name nil to preserve the current value.
func updateFreeform(ctx context.Context, client *arize.Client, configID string) {
	name := "example-notes-v2"
	ac, err := client.AnnotationConfigs.UpdateFreeform(ctx, annotationconfigs.UpdateFreeformRequest{
		AnnotationConfig: configID,
		Name:             &name,
	})
	if err != nil {
		log.Fatalf("update freeform annotation config: %v", err)
	}
	id, _, updatedName := identify(ac)
	fmt.Printf("updated freeform annotation config %s (%s)\n", updatedName, id)
}

func deleteAnnotationConfig(ctx context.Context, client *arize.Client, configID string) {
	err := client.AnnotationConfigs.Delete(ctx, annotationconfigs.DeleteRequest{
		AnnotationConfig: configID,
	})
	if err != nil {
		log.Fatalf("delete annotation config: %v", err)
	}
	fmt.Printf("deleted annotation config %s\n", configID)
}

// identify unwraps the discriminated union and returns the common id/name
// fields along with a human-readable variant label. ValueByDiscriminator reads
// the discriminator and returns the matching concrete variant; type-switch over
// it to cover every variant.
func identify(ac *annotationconfigs.AnnotationConfig) (id, kind, name string) {
	v, err := ac.ValueByDiscriminator()
	if err != nil {
		return "", "", ""
	}
	switch cfg := v.(type) {
	case annotationconfigs.CategoricalAnnotationConfig:
		return cfg.Id, string(annotationconfigs.AnnotationConfigTypeCategorical), cfg.Name
	case annotationconfigs.ContinuousAnnotationConfig:
		return cfg.Id, string(annotationconfigs.AnnotationConfigTypeContinuous), cfg.Name
	case annotationconfigs.FreeformAnnotationConfig:
		return cfg.Id, string(annotationconfigs.AnnotationConfigTypeFreeform), cfg.Name
	}
	return "", "", ""
}

func idOf(ac *annotationconfigs.AnnotationConfig) string {
	id, _, _ := identify(ac)
	return id
}
