// Package resolve implements name-or-ID lookup for Arize resources.
//
// Each FindXID helper accepts a string that is either a base64-encoded resource
// ID or a human-readable name, plus parent context fields (e.g. space) needed
// to disambiguate names. If the input looks like an ID, it is returned as-is.
// Otherwise the helper paginates the appropriate List endpoint with a name +
// parent filter until it finds an exact name match.
//
// Helpers return ResourceNotFoundError when a name cannot be resolved.
package resolve

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// listPageSize is the per-page limit used while resolving names.
const listPageSize = 100

// ResourceNotFoundError is returned when a resource name cannot be resolved
// to an ID. Its zero-allocation Error() format includes the resource type,
// the name that was searched for, optional list of names that were observed,
// and an optional hint about how to recover.
type ResourceNotFoundError struct {
	ResourceType string
	Name         string
	Available    []string
	Hint         string
}

// AmbiguousNameError is returned when a resource name matches more than one
// resource (e.g. a space name shared across different organizations). The
// caller should pass a resource ID instead of the name to disambiguate.
type AmbiguousNameError struct {
	ResourceType string
	Name         string
	MatchingIDs  []string
}

func (e *AmbiguousNameError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Multiple %ss named %q found. ", e.ResourceType, e.Name)
	fmt.Fprintf(&b, "Use a %s ID to disambiguate. ", e.ResourceType)
	fmt.Fprintf(&b, "Matching IDs: %s", strings.Join(e.MatchingIDs, ", "))
	return b.String()
}

func (e *ResourceNotFoundError) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s %q not found", e.ResourceType, e.Name)
	if len(e.Available) > 0 {
		fmt.Fprintf(&b, ". Available %ss: %s", e.ResourceType, strings.Join(e.Available, ", "))
	}
	if e.Hint != "" {
		fmt.Fprintf(&b, ". %s", e.Hint)
	}
	return b.String()
}

// IsResourceID returns true if value looks like a base64-encoded Arize global
// resource ID (decodes to a string containing a colon). Names that happen to
// be valid base64 of the right shape are extremely unlikely to also decode to
// a string containing a colon.
func IsResourceID(value string) bool {
	if value == "" {
		return false
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return false
	}
	return strings.Contains(string(decoded), ":")
}

// applySpaceFilter wires a space (passed as either an ID or a name) into the
// space_id / space_name query params used by every space-scoped list call.
type spaceFilter struct {
	spaceID   *string
	spaceName *string
}

func resolveSpaceFilter(space string) spaceFilter {
	if space == "" {
		return spaceFilter{}
	}
	if IsResourceID(space) {
		return spaceFilter{spaceID: &space}
	}
	return spaceFilter{spaceName: &space}
}

// requireParent constructs a ResourceNotFoundError when a name lookup needs a
// parent context (typically space) that the caller didn't provide.
func requireParent(resourceType, name, parent string) error {
	return &ResourceNotFoundError{
		ResourceType: resourceType,
		Name:         name,
		Hint: fmt.Sprintf(
			"Provide '%s' so the %s name can be resolved, or provide the %s ID instead of the name.",
			parent, resourceType, resourceType,
		),
	}
}

// ===========================================================================
// FindSpaceID
// ===========================================================================

// FindSpaceID resolves a space ID or name to a space ID.
//
// Returns AmbiguousNameError if multiple spaces share the same name (e.g.
// across different organizations). Pass a space ID to disambiguate.
func FindSpaceID(ctx context.Context, gen *generated.ClientWithResponses, space string) (string, error) {
	if IsResourceID(space) {
		return space, nil
	}
	var matches []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.SpacesListParams{Name: &space, Limit: &limit}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.SpacesListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, s := range resp.JSON200.Spaces {
			if s.Name == space {
				matches = append(matches, s.Id)
			}
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	if len(matches) > 1 {
		return "", &AmbiguousNameError{ResourceType: "space", Name: space, MatchingIDs: matches}
	}
	if len(matches) == 1 {
		return matches[0], nil
	}
	return "", &ResourceNotFoundError{ResourceType: "space", Name: space}
}

// ===========================================================================
// FindOrganizationID
// ===========================================================================

// FindOrganizationID resolves an organization ID or name to an ID.
func FindOrganizationID(ctx context.Context, gen *generated.ClientWithResponses, organization string) (string, error) {
	if IsResourceID(organization) {
		return organization, nil
	}
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.OrganizationsListParams{Name: &organization, Limit: &limit}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.OrganizationsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, o := range resp.JSON200.Organizations {
			if o.Name == organization {
				return o.Id, nil
			}
			available = append(available, o.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "organization", Name: organization, Available: available}
}

// ===========================================================================
// FindRoleID
// ===========================================================================

// FindRoleID resolves a role ID or name to an ID.
func FindRoleID(ctx context.Context, gen *generated.ClientWithResponses, role string) (string, error) {
	if IsResourceID(role) {
		return role, nil
	}
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.RolesListParams{Limit: &limit}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.RolesListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, r := range resp.JSON200.Roles {
			if r.Name == role {
				return r.Id, nil
			}
			available = append(available, r.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "role", Name: role, Available: available}
}

// ===========================================================================
// FindProjectID
// ===========================================================================

// FindProjectID resolves a project ID or name to an ID. Space (ID or name) is
// required when project is a name.
func FindProjectID(ctx context.Context, gen *generated.ClientWithResponses, project, space string) (string, error) {
	if IsResourceID(project) {
		return project, nil
	}
	if space == "" {
		return "", requireParent("project", project, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.ProjectsListParams{Name: &project, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.ProjectsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, pr := range resp.JSON200.Projects {
			if pr.Name == project {
				return pr.Id, nil
			}
			available = append(available, pr.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "project", Name: project, Available: available}
}

// ===========================================================================
// FindDatasetID
// ===========================================================================

// FindDatasetID resolves a dataset ID or name to an ID. Space (ID or name) is
// required when dataset is a name.
func FindDatasetID(ctx context.Context, gen *generated.ClientWithResponses, dataset, space string) (string, error) {
	if IsResourceID(dataset) {
		return dataset, nil
	}
	if space == "" {
		return "", requireParent("dataset", dataset, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.DatasetsListParams{Name: &dataset, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.DatasetsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, d := range resp.JSON200.Datasets {
			if d.Name == dataset {
				return d.Id, nil
			}
			available = append(available, d.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "dataset", Name: dataset, Available: available}
}

// ===========================================================================
// FindExperimentID
// ===========================================================================

// FindExperimentID resolves an experiment ID or name to an ID. Dataset (ID or
// name) is required when experiment is a name; space is required when dataset
// is itself passed as a name.
func FindExperimentID(ctx context.Context, gen *generated.ClientWithResponses, experiment, dataset, space string) (string, error) {
	if IsResourceID(experiment) {
		return experiment, nil
	}
	if dataset == "" {
		return "", requireParent("experiment", experiment, "dataset")
	}
	if !IsResourceID(dataset) && space == "" {
		return "", &ResourceNotFoundError{
			ResourceType: "experiment",
			Name:         experiment,
			Hint: "Provide 'space' so the dataset name can be resolved, " +
				"which is needed to resolve the experiment name. Alternatively, " +
				"provide the experiment ID, or the dataset ID instead of the name.",
		}
	}
	datasetID, err := FindDatasetID(ctx, gen, dataset, space)
	if err != nil {
		return "", err
	}
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.ExperimentsListParams{DatasetId: &datasetID, Limit: &limit}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.ExperimentsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, e := range resp.JSON200.Experiments {
			if e.Name == experiment {
				return e.Id, nil
			}
			available = append(available, e.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "experiment", Name: experiment, Available: available}
}

// ===========================================================================
// FindPromptID
// ===========================================================================

// FindPromptID resolves a prompt ID or name to an ID. Space (ID or name) is
// required when prompt is a name.
func FindPromptID(ctx context.Context, gen *generated.ClientWithResponses, prompt, space string) (string, error) {
	if IsResourceID(prompt) {
		return prompt, nil
	}
	if space == "" {
		return "", requireParent("prompt", prompt, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.PromptsListParams{Name: &prompt, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.PromptsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, pr := range resp.JSON200.Prompts {
			if pr.Name == prompt {
				return pr.Id, nil
			}
			available = append(available, pr.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "prompt", Name: prompt, Available: available}
}

// ===========================================================================
// FindEvaluatorID
// ===========================================================================

// FindEvaluatorID resolves an evaluator ID or name to an ID. Space (ID or
// name) is required when evaluator is a name.
func FindEvaluatorID(ctx context.Context, gen *generated.ClientWithResponses, evaluator, space string) (string, error) {
	if IsResourceID(evaluator) {
		return evaluator, nil
	}
	if space == "" {
		return "", requireParent("evaluator", evaluator, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.EvaluatorsListParams{Name: &evaluator, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.EvaluatorsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, ev := range resp.JSON200.Evaluators {
			if ev.Name == evaluator {
				return ev.Id, nil
			}
			available = append(available, ev.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "evaluator", Name: evaluator, Available: available}
}

// ===========================================================================
// FindAnnotationConfigID
// ===========================================================================

// FindAnnotationConfigID resolves an annotation config ID or name to an ID.
// Space (ID or name) is required when annotationConfig is a name. Annotation
// configs are returned by the API as a oneof discriminated union; this helper
// re-marshals each entry to extract its base Id/Name fields.
func FindAnnotationConfigID(ctx context.Context, gen *generated.ClientWithResponses, annotationConfig, space string) (string, error) {
	if IsResourceID(annotationConfig) {
		return annotationConfig, nil
	}
	if space == "" {
		return "", requireParent("annotation config", annotationConfig, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.AnnotationConfigsListParams{Name: &annotationConfig, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.AnnotationConfigsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, cfg := range resp.JSON200.AnnotationConfigs {
			id, name, ok := annotationConfigIDAndName(cfg)
			if !ok {
				continue
			}
			if name == annotationConfig {
				return id, nil
			}
			available = append(available, name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "annotation config", Name: annotationConfig, Available: available}
}

// annotationConfigIDAndName extracts (id, name) from a oneof AnnotationConfig
// by re-marshalling to JSON and unmarshalling to AnnotationConfigBase.
func annotationConfigIDAndName(cfg generated.AnnotationConfig) (string, string, bool) {
	b, err := json.Marshal(cfg)
	if err != nil {
		return "", "", false
	}
	var base generated.AnnotationConfigBase
	if err := json.Unmarshal(b, &base); err != nil {
		return "", "", false
	}
	return base.Id, base.Name, true
}

// ===========================================================================
// FindAnnotationQueueID
// ===========================================================================

// FindAnnotationQueueID resolves an annotation queue ID or name to an ID.
// Space (ID or name) is required when annotationQueue is a name.
func FindAnnotationQueueID(ctx context.Context, gen *generated.ClientWithResponses, annotationQueue, space string) (string, error) {
	if IsResourceID(annotationQueue) {
		return annotationQueue, nil
	}
	if space == "" {
		return "", requireParent("annotation queue", annotationQueue, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.AnnotationQueuesListParams{Name: &annotationQueue, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.AnnotationQueuesListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, q := range resp.JSON200.AnnotationQueues {
			if q.Name == annotationQueue {
				return q.Id, nil
			}
			available = append(available, q.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "annotation queue", Name: annotationQueue, Available: available}
}

// ===========================================================================
// FindAIIntegrationID
// ===========================================================================

// FindAIIntegrationID resolves an AI integration ID or name to an ID. Space
// (ID or name) is required when integration is a name.
func FindAIIntegrationID(ctx context.Context, gen *generated.ClientWithResponses, integration, space string) (string, error) {
	if IsResourceID(integration) {
		return integration, nil
	}
	if space == "" {
		return "", requireParent("AI integration", integration, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.AiIntegrationsListParams{Name: &integration, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.AiIntegrationsListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, ai := range resp.JSON200.AiIntegrations {
			if ai.Name == integration {
				return ai.Id, nil
			}
			available = append(available, ai.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "AI integration", Name: integration, Available: available}
}

// ===========================================================================
// FindTaskID
// ===========================================================================

// FindTaskID resolves a task ID or name to an ID. Space (ID or name) is
// required when task is a name.
func FindTaskID(ctx context.Context, gen *generated.ClientWithResponses, task, space string) (string, error) {
	if IsResourceID(task) {
		return task, nil
	}
	if space == "" {
		return "", requireParent("task", task, "space")
	}
	sf := resolveSpaceFilter(space)
	var available []string
	limit := listPageSize
	var cursor string
	for {
		p := &generated.TasksListParams{Name: &task, Limit: &limit, SpaceId: sf.spaceID, SpaceName: sf.spaceName}
		if cursor != "" {
			p.Cursor = &cursor
		}
		resp, err := gen.TasksListWithResponse(ctx, p)
		if err != nil {
			return "", err
		}
		if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
			return "", err
		}
		for _, t := range resp.JSON200.Tasks {
			if t.Name == task {
				return t.Id, nil
			}
			available = append(available, t.Name)
		}
		if !resp.JSON200.Pagination.HasMore || resp.JSON200.Pagination.NextCursor == nil {
			break
		}
		cursor = *resp.JSON200.Pagination.NextCursor
	}
	return "", &ResourceNotFoundError{ResourceType: "task", Name: task, Available: available}
}
