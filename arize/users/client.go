// Package users provides access to the Arize account Users API.
package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/apierrors"
	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
	"github.com/Arize-ai/client-go-v2/arize/internal/optfields"
	"github.com/Arize-ai/client-go-v2/arize/internal/prerelease"
	"github.com/Arize-ai/client-go-v2/arize/internal/resolve"
)

// Client provides access to the Arize Users API.
type Client struct {
	gen *generated.ClientWithResponses
}

// New constructs a Client from a generated ClientWithResponses.
func New(gen *generated.ClientWithResponses) *Client {
	return &Client{gen: gen}
}

// List returns a paginated list of account users. Defaults to a page size of 50.
func (c *Client) List(ctx context.Context, req ListRequest) (*ListUsers, error) {
	prerelease.Warn("users.list", prerelease.Beta)
	params := &generated.ListUsersParams{
		Email:  optfields.PtrIfSet(req.Email),
		Limit:  optfields.PtrWithDefault(req.Limit, optfields.DefaultListLimit),
		Cursor: optfields.PtrIfSet(req.Cursor),
	}
	if len(req.Status) > 0 {
		status := generated.UserStatusQueryParam(req.Status)
		params.Status = &status
	}
	resp, err := c.gen.ListUsersWithResponse(ctx, params)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Update updates a user's display name and/or developer permission by ID.
// Leave a field nil to preserve its current value. At least one of Name or
// IsDeveloper must be non-nil.
func (c *Client) Update(ctx context.Context, req UpdateRequest) (*User, error) {
	prerelease.Warn("users.update", prerelease.Beta)
	if req.Name == nil && req.IsDeveloper == nil {
		return nil, fmt.Errorf("users: update requires at least one of Name or IsDeveloper")
	}
	body := generated.UpdateUserRequest{
		Name:        req.Name,
		IsDeveloper: req.IsDeveloper,
	}
	resp, err := c.gen.UpdateUserWithResponse(ctx, req.UserID, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// Create creates a new account user and returns it.
//
// The endpoint returns the created user on 201, or the existing user on a 200
// idempotency hit (same email, when InviteMode is not InviteModeNone). Both are
// returned as *User. The 201 response also carries a one-time temporary
// password when InviteMode is InviteModeTemporaryPassword; that value is not
// surfaced by this method.
func (c *Client) Create(ctx context.Context, req CreateRequest) (*User, error) {
	prerelease.Warn("users.create", prerelease.Beta)
	body := generated.CreateUserRequest{
		Name:        req.Name,
		Email:       generated.Email(req.Email),
		Role:        req.Role,
		InviteMode:  req.InviteMode,
		IsDeveloper: req.IsDeveloper,
	}
	resp, err := c.gen.CreateUserWithResponse(ctx, body)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	if resp.JSON200 != nil {
		return resp.JSON200, nil
	}
	if resp.JSON201 != nil {
		created := resp.JSON201
		return &User{
			Id:          created.Id,
			Name:        created.Name,
			Email:       created.Email,
			CreatedAt:   created.CreatedAt,
			Status:      created.Status,
			Role:        created.Role,
			IsDeveloper: created.IsDeveloper,
		}, nil
	}
	return nil, fmt.Errorf("users: create returned unexpected status %d", resp.StatusCode())
}

// Delete soft-deletes a user by ID. Cascades to organization and space
// memberships, API keys, and role bindings. Idempotent: deleting an already
// inactive user succeeds.
func (c *Client) Delete(ctx context.Context, req DeleteRequest) error {
	prerelease.Warn("users.delete", prerelease.Beta)
	resp, err := c.gen.DeleteUserWithResponse(ctx, req.UserID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ResendInvitation resends the invitation email for a pending (invited) user.
// The target user must be in the invited state.
func (c *Client) ResendInvitation(ctx context.Context, req ResendInvitationRequest) error {
	prerelease.Warn("users.resend_invitation", prerelease.Beta)
	resp, err := c.gen.ResendUserInvitationWithResponse(ctx, req.UserID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// ResetPassword triggers a password-reset email for a user. The user must
// authenticate via password (not SSO/SAML) and must have verified their account.
func (c *Client) ResetPassword(ctx context.Context, req ResetPasswordRequest) error {
	prerelease.Warn("users.reset_password", prerelease.Beta)
	resp, err := c.gen.ResetUserPasswordWithResponse(ctx, req.UserID)
	if err != nil {
		return err
	}
	return apierrors.CheckResponse(resp.HTTPResponse, resp.Body)
}

// Get returns a single user. req.User accepts either a user ID or an email
// address; an email is resolved to an ID via the users list endpoint
// (case-insensitive exact match), and a non-matching email yields a
// *ResourceNotFoundError.
func (c *Client) Get(ctx context.Context, req GetRequest) (*User, error) {
	prerelease.Warn("users.get", prerelease.Beta)
	id, err := resolve.FindUserID(ctx, c.gen, req.User)
	if err != nil {
		return nil, err
	}
	resp, err := c.gen.GetUserWithResponse(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := apierrors.CheckResponse(resp.HTTPResponse, resp.Body); err != nil {
		return nil, err
	}
	return resp.JSON200, nil
}

// BulkDelete deletes users by ID and/or email. At least one of UserIDs or
// Emails must be provided. Each email is resolved to a user ID client-side via
// the users list endpoint (case-insensitive exact match); an unresolved email
// is recorded as NotFound rather than aborting the batch. Each deletion's
// outcome is returned as a BulkUserDeletionResult.
// Per-user delete failures are recorded as Failed and do not abort the batch.
func (c *Client) BulkDelete(ctx context.Context, req BulkDeleteRequest) ([]BulkUserDeletionResult, error) {
	prerelease.Warn("users.bulk_delete", prerelease.Beta)
	if len(req.UserIDs) == 0 && len(req.Emails) == 0 {
		return nil, fmt.Errorf("users: bulk delete requires at least one of UserIDs or Emails")
	}

	results := make([]BulkUserDeletionResult, 0, len(req.UserIDs)+len(req.Emails))
	idsToDelete := make([]string, 0, len(req.UserIDs)+len(req.Emails))
	idsToDelete = append(idsToDelete, req.UserIDs...)

	idToEmail := make(map[string]string, len(req.Emails))
	for _, email := range req.Emails {
		id, err := resolve.FindUserID(ctx, c.gen, email)
		if err != nil {
			var rnfe *resolve.ResourceNotFoundError
			if errors.As(err, &rnfe) {
				results = append(results, BulkUserDeletionResult{
					Email:  email,
					Status: DeletionStatusNotFound,
					Error:  fmt.Sprintf("no user found with email %q", email),
				})
				continue
			}
			return nil, err
		}
		idToEmail[id] = email
		idsToDelete = append(idsToDelete, id)
	}

	for _, id := range idsToDelete {
		if err := c.Delete(ctx, DeleteRequest{UserID: id}); err != nil {
			results = append(results, BulkUserDeletionResult{
				UserID: id,
				Email:  idToEmail[id],
				Status: DeletionStatusFailed,
				Error:  err.Error(),
			})
			continue
		}
		results = append(results, BulkUserDeletionResult{UserID: id, Email: idToEmail[id], Status: DeletionStatusDeleted})
	}

	return results, nil
}
