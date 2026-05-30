package annotationqueues

import (
	"context"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Annotation Queues API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of annotation queues. req.Space, when
// non-empty, accepts a space name or ID and restricts results to that space.
func (c *Client) List(ctx context.Context, req ListRequest) (*AnnotationQueueList, error) {
	prerelease.Warn("annotationqueues.list", prerelease.Beta)
	params := generated.AnnotationQueuesListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.AnnotationQueuesListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single annotation queue, resolving by name or ID.
func (c *Client) Get(ctx context.Context, req GetRequest) (*AnnotationQueue, error) {
	prerelease.Warn("annotationqueues.get", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationQueuesGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new annotation queue, resolving the parent space by name or ID.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*AnnotationQueue, error) {
	prerelease.Warn("annotationqueues.create", prerelease.Beta)
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.AnnotationQueuesCreateJSONRequestBody{
		SpaceId:             spaceID,
		Name:                req.Name,
		AnnotationConfigIds: req.AnnotationConfigIDs,
		AnnotatorEmails:     req.AnnotatorEmails,
		AssignmentMethod:    optfields.PtrIfSet(req.AssignmentMethod),
		Instructions:        optfields.PtrIfSet(req.Instructions),
	}
	if len(req.RecordSources) > 0 {
		body.RecordSources = &req.RecordSources
	}
	resp, err := c.gen.AnnotationQueuesCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update updates an existing annotation queue, resolving by name or ID.
func (c *Client) Update(ctx context.Context, req UpdateRequest) (*AnnotationQueue, error) {
	prerelease.Warn("annotationqueues.update", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationQueuesUpdateWithResponse(ctx, id, generated.AnnotationQueuesUpdateJSONRequestBody{
		Name:                req.Name,
		Instructions:        req.Instructions,
		AnnotatorEmails:     req.AnnotatorEmails,
		AnnotationConfigIds: req.AnnotationConfigIDs,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete removes an annotation queue, resolving by name or ID.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("annotationqueues.delete", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.AnnotationQueuesDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListRecords returns a paginated list of records for an annotation queue,
// resolving the queue by name or ID.
func (c *Client) ListRecords(ctx context.Context, req ListRecordsRequest) (*AnnotationQueueRecordList, error) {
	prerelease.Warn("annotationqueues.list_records", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	params := &generated.AnnotationQueueRecordsListParams{
		Cursor: optfields.PtrIfSet(req.Cursor),
		Limit:  optfields.PtrIfSet(req.Limit),
	}
	resp, err := c.gen.AnnotationQueueRecordsListWithResponse(ctx, id, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// AddRecords adds records to an annotation queue (resolved by name or ID) and
// returns the created records.
func (c *Client) AddRecords(ctx context.Context, req AddRecordsRequest) (*AnnotationQueueRecordCreate, error) {
	prerelease.Warn("annotationqueues.add_records", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationQueuesRecordsCreateWithResponse(ctx, id, generated.AnnotationQueuesRecordsCreateJSONRequestBody{
		RecordSources: req.RecordSources,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	switch {
	case resp.JSON201 != nil:
		return resp.JSON201, nil
	case resp.JSON200 != nil:
		return resp.JSON200, nil
	default:
		return nil, fmt.Errorf("annotationqueues: AddRecords: unexpected empty body (status %d)", resp.StatusCode())
	}
}

// DeleteRecords removes records from an annotation queue, resolving the queue
// by name or ID.
func (c *Client) DeleteRecords(ctx context.Context, req DeleteRecordsRequest) error {
	prerelease.Warn("annotationqueues.delete_records", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.AnnotationQueuesRecordsDeleteWithResponse(ctx, id, generated.AnnotationQueuesRecordsDeleteJSONRequestBody{
		RecordIds: req.RecordIDs,
	})
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Annotate submits annotations for a record in an annotation queue, resolving
// the queue by name or ID. RecordID is a pure ID with no name resolution.
func (c *Client) Annotate(ctx context.Context, req AnnotateRequest) (*AnnotationQueueRecordAnnotateResult, error) {
	prerelease.Warn("annotationqueues.annotate", prerelease.Alpha)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationQueuesRecordsAnnotateWithResponse(ctx, id, req.RecordID, generated.AnnotationQueuesRecordsAnnotateJSONRequestBody{
		Annotations: req.Annotations,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Assign assigns users to a record in an annotation queue, resolving the queue
// by name or ID. RecordID is a pure ID with no name resolution.
func (c *Client) Assign(ctx context.Context, req AssignRequest) (*AnnotationQueueRecordAssignResult, error) {
	prerelease.Warn("annotationqueues.assign", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotationQueuesRecordsAssignWithResponse(ctx, id, req.RecordID, generated.AnnotationQueuesRecordsAssignJSONRequestBody{
		AssignedUserEmails: req.AssignedUserEmails,
	})
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}
