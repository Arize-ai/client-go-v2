package datasets

import (
	"context"
	"errors"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// ErrNoExamples is returned by Create when the request carries no examples.
// Empty datasets are not allowed.
var ErrNoExamples = errors.New("datasets: cannot create a dataset without examples")

// Client provides access to the Arize Datasets API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of datasets. req.Space, when non-empty,
// accepts a space name or ID and restricts results to that space.
func (c *Client) List(
	ctx context.Context,
	req ListRequest,
) (*DatasetList, error) {
	prerelease.Warn("datasets.list", prerelease.Alpha)
	params := generated.DatasetsListParams{
		Name:   optfields.PtrIfSet(req.Name),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	params.SpaceId, params.SpaceName = resolve.ResolveSpaceFilter(req.Space)
	resp, err := c.gen.DatasetsListWithResponse(ctx, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Get returns a single dataset. req.Dataset accepts a name or ID; req.Space is
// required when req.Dataset is a name.
func (c *Client) Get(
	ctx context.Context,
	req GetRequest,
) (*Dataset, error) {
	prerelease.Warn("datasets.get", prerelease.Alpha)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.DatasetsGetWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new dataset and returns it. req.Space accepts a name or ID
// and identifies the parent space. At least one example is required; an empty
// req.Examples returns ErrNoExamples.
func (c *Client) Create(
	ctx context.Context,
	req CreateRequest,
) (*Dataset, error) {
	prerelease.Warn("datasets.create", prerelease.Alpha)
	if len(req.Examples) == 0 {
		return nil, ErrNoExamples
	}
	spaceID, err := resolve.FindSpaceID(ctx, c.gen, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.DatasetsCreateJSONRequestBody{
		SpaceId:  spaceID,
		Name:     req.Name,
		Examples: req.Examples,
	}
	resp, err := c.gen.DatasetsCreateWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// Update renames a dataset and returns it. req.Dataset accepts a name or ID;
// req.Space is required when req.Dataset is a name.
func (c *Client) Update(
	ctx context.Context,
	req UpdateRequest,
) (*Dataset, error) {
	prerelease.Warn("datasets.update", prerelease.Beta)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.DatasetsUpdateJSONRequestBody{
		Name: req.Name,
	}
	resp, err := c.gen.DatasetsUpdateWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Delete irreversibly removes a dataset. req.Dataset accepts a name or ID;
// req.Space is required when req.Dataset is a name.
func (c *Client) Delete(
	ctx context.Context,
	req DeleteRequest,
) error {
	prerelease.Warn("datasets.delete", prerelease.Alpha)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return err
	}
	resp, err := c.gen.DatasetsDeleteWithResponse(ctx, id)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ListExamples returns a paginated list of examples for a dataset.
// req.Dataset accepts a name or ID; req.Space is required when req.Dataset is
// a name.
func (c *Client) ListExamples(
	ctx context.Context,
	req ListExamplesRequest,
) (*DatasetExampleList, error) {
	prerelease.Warn("datasets.list_examples", prerelease.Alpha)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.DatasetsExamplesListParams{
		Limit:            optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		DatasetVersionId: optfields.PtrIfSet(req.DatasetVersionID),
		Cursor:           optfields.PtrIfSet(req.Cursor),
	}
	resp, err := c.gen.DatasetsExamplesListWithResponse(ctx, id, &params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// AppendExamples appends new examples to a dataset version and returns the
// version they were written to along with the server-assigned example IDs.
// req.Dataset accepts a name or ID; req.Space is required when req.Dataset is
// a name.
func (c *Client) AppendExamples(
	ctx context.Context,
	req AppendExamplesRequest,
) (*DatasetExamplesInserted, error) {
	prerelease.Warn("datasets.append_examples", prerelease.Alpha)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	params := generated.DatasetsExamplesInsertParams{
		DatasetVersionId: optfields.PtrIfSet(req.DatasetVersionID),
	}
	body := generated.DatasetsExamplesInsertJSONRequestBody{
		Examples: req.Examples,
	}
	resp, err := c.gen.DatasetsExamplesInsertWithResponse(ctx, id, &params, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON201, nil
}

// DeleteExamples removes a batch of examples from a specific dataset version.
// The result reports which IDs were deleted and which were not; when Completed
// is false the delete did not fully complete and the full request should be
// retried, which is safe because the operation is idempotent. req.Dataset
// accepts a name or ID; req.Space is required when req.Dataset is a name.
func (c *Client) DeleteExamples(
	ctx context.Context,
	req DeleteExamplesRequest,
) (*DatasetExamplesDeleteResult, error) {
	prerelease.Warn("datasets.delete_examples", prerelease.Alpha)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return nil, err
	}
	body := generated.DatasetsExamplesDeleteJSONRequestBody{
		DatasetVersionId: req.DatasetVersionID,
		ExampleIds:       req.ExampleIDs,
	}
	resp, err := c.gen.DatasetsExamplesDeleteWithResponse(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// AnnotateExamples writes human annotations to a batch of examples in a
// dataset. Annotations are upserted by annotation config name for each
// example, so re-annotating the same example with the same config name
// overwrites the previous value. req.Dataset accepts a name or ID; req.Space
// is required when req.Dataset is a name.
func (c *Client) AnnotateExamples(
	ctx context.Context,
	req AnnotateExamplesRequest,
) error {
	prerelease.Warn("datasets.annotate_examples", prerelease.Beta)
	id, err := resolve.FindDatasetID(ctx, c.gen, req.Dataset, req.Space)
	if err != nil {
		return err
	}
	body := generated.DatasetsExamplesAnnotateJSONRequestBody{
		Annotations: req.Annotations,
	}
	resp, err := c.gen.DatasetsExamplesAnnotateWithResponse(ctx, id, body)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}
