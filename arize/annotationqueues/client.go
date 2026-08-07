package annotationqueues

import (
	"bytes"
	"context"
	"encoding/json"
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
func (c *Client) List(ctx context.Context, req ListRequest) (*ListAnnotationQueues, error) {
	prerelease.Warn("annotationqueues.list", prerelease.Beta)
	params := generated.ListAnnotationQueuesParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrIfSet(req.Limit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.ListAnnotationQueuesWithResponse(ctx, &params)
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
	resp, err := c.gen.GetAnnotationQueueWithResponse(ctx, id)
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
	body := generated.CreateAnnotationQueueJSONRequestBody{
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
	resp, err := c.gen.CreateAnnotationQueueWithResponse(ctx, body)
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
	body := map[string]any{}
	if req.Name != nil {
		body["name"] = *req.Name
	}
	if req.Instructions != nil {
		if *req.Instructions == "" {
			body["instructions"] = nil
		} else {
			body["instructions"] = *req.Instructions
		}
	}
	if req.AnnotatorEmails != nil {
		body["annotator_emails"] = *req.AnnotatorEmails
	}
	if req.AnnotationConfigIDs != nil {
		body["annotation_config_ids"] = *req.AnnotationConfigIDs
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("annotationqueues: marshal update body: %w", err)
	}
	resp, err := c.gen.UpdateAnnotationQueueWithBodyWithResponse(ctx, id, "application/json", bytes.NewReader(raw))
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
	resp, err := c.gen.DeleteAnnotationQueueWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListRecords returns a paginated list of records for an annotation queue,
// resolving the queue by name or ID.
func (c *Client) ListRecords(ctx context.Context, req ListRecordsRequest) (*ListAnnotationQueueRecords, error) {
	prerelease.Warn("annotationqueues.list_records", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	params := &generated.ListAnnotationQueueRecordsParams{
		Cursor: optfields.PtrIfSet(req.Cursor),
		Limit:  optfields.PtrIfSet(req.Limit),
	}
	resp, err := c.gen.ListAnnotationQueueRecordsWithResponse(ctx, id, params)
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
func (c *Client) AddRecords(ctx context.Context, req AddRecordsRequest) (*CreateAnnotationQueueRecord, error) {
	prerelease.Warn("annotationqueues.add_records", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.CreateAnnotationQueueRecordWithResponse(ctx, id, generated.CreateAnnotationQueueRecordJSONRequestBody{
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
	resp, err := c.gen.DeleteAnnotationQueueRecordWithResponse(ctx, id, generated.DeleteAnnotationQueueRecordJSONRequestBody{
		RecordIds: req.RecordIDs,
	})
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Annotate submits annotations for a record in an annotation queue, resolving
// the queue by name or ID. RecordID is a pure ID with no name resolution.
func (c *Client) Annotate(ctx context.Context, req AnnotateRequest) (*AnnotateAnnotationQueueRecord, error) {
	prerelease.Warn("annotationqueues.annotate", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AnnotateAnnotationQueueRecordWithResponse(ctx, id, req.RecordID, generated.AnnotateAnnotationQueueRecordJSONRequestBody{
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
func (c *Client) Assign(ctx context.Context, req AssignRequest) (*AssignAnnotationQueueRecord, error) {
	prerelease.Warn("annotationqueues.assign", prerelease.Beta)
	id, err := resolve.FindAnnotationQueueID(ctx, c.gen, req.AnnotationQueue, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.AssignAnnotationQueueRecordWithResponse(ctx, id, req.RecordID, generated.AssignAnnotationQueueRecordJSONRequestBody{
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
