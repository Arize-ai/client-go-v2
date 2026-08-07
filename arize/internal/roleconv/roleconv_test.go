package roleconv

import (
	"errors"
	"testing"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

type errorRole struct{}

func (errorRole) MarshalJSON() ([]byte, error) {
	return nil, errors.New("invalid role")
}

func TestIsZeroReturnsFalseOnMarshalError(t *testing.T) {
	if IsZero(errorRole{}) {
		t.Fatal("IsZero should return false when MarshalJSON fails")
	}
}

func TestSpaceRoleAssignmentRequest(t *testing.T) {
	tests := []struct {
		name        string
		buildInput  func() generated.SpaceRoleAssignment
		wantErr     bool
		checkResult func(t *testing.T, got generated.SpaceRoleAssignmentRequest)
	}{
		{
			name: "predefined role",
			buildInput: func() generated.SpaceRoleAssignment {
				var r generated.SpaceRoleAssignment
				if err := r.FromPredefinedRoleAssignment(generated.PredefinedRoleAssignment{
					Name: generated.UserSpaceRoleMEMBER,
					Type: generated.SpaceRoleAssignmentTypePREDEFINED,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.SpaceRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.PredefinedRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected PredefinedRoleAssignmentRequest, got %T", v)
				}
				if req.Name != generated.UserSpaceRoleMEMBER {
					t.Errorf("Name = %v, want MEMBER", req.Name)
				}
			},
		},
		{
			name: "custom role",
			buildInput: func() generated.SpaceRoleAssignment {
				var r generated.SpaceRoleAssignment
				if err := r.FromCustomRoleAssignment(generated.CustomRoleAssignment{
					Id:   "role-abc",
					Type: generated.SpaceRoleAssignmentTypeCUSTOM,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.SpaceRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.CustomRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected CustomRoleAssignmentRequest, got %T", v)
				}
				if req.Id != "role-abc" {
					t.Errorf("Id = %v, want role-abc", req.Id)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SpaceRoleAssignmentRequest(tt.buildInput())
			if (err != nil) != tt.wantErr {
				t.Fatalf("SpaceRoleAssignmentRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				tt.checkResult(t, got)
			}
		})
	}
}

func TestOrgRoleAssignmentRequest(t *testing.T) {
	tests := []struct {
		name        string
		buildInput  func() generated.OrganizationRoleAssignment
		wantErr     bool
		checkResult func(t *testing.T, got generated.OrganizationRoleAssignmentRequest)
	}{
		{
			name: "predefined role",
			buildInput: func() generated.OrganizationRoleAssignment {
				var r generated.OrganizationRoleAssignment
				if err := r.FromOrganizationPredefinedRoleAssignment(generated.OrganizationPredefinedRoleAssignment{
					Name: generated.OrganizationRoleMEMBER,
					Type: generated.OrganizationRoleAssignmentTypePREDEFINED,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.OrganizationRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.OrganizationPredefinedRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected OrganizationPredefinedRoleAssignmentRequest, got %T", v)
				}
				if req.Name != generated.OrganizationRoleMEMBER {
					t.Errorf("Name = %v, want MEMBER", req.Name)
				}
			},
		},
		{
			name: "custom role",
			buildInput: func() generated.OrganizationRoleAssignment {
				var r generated.OrganizationRoleAssignment
				if err := r.FromOrganizationCustomRoleAssignment(generated.OrganizationCustomRoleAssignment{
					Id:   "org-role-xyz",
					Type: generated.OrganizationRoleAssignmentTypeCUSTOM,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.OrganizationRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.OrganizationCustomRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected OrganizationCustomRoleAssignmentRequest, got %T", v)
				}
				if req.Id != "org-role-xyz" {
					t.Errorf("Id = %v, want org-role-xyz", req.Id)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OrgRoleAssignmentRequest(tt.buildInput())
			if (err != nil) != tt.wantErr {
				t.Fatalf("OrgRoleAssignmentRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				tt.checkResult(t, got)
			}
		})
	}
}

func TestAccountRoleAssignmentRequest(t *testing.T) {
	tests := []struct {
		name        string
		buildInput  func() generated.UserRoleAssignment
		wantErr     bool
		checkResult func(t *testing.T, got generated.UserRoleAssignmentRequest)
	}{
		{
			name: "predefined role",
			buildInput: func() generated.UserRoleAssignment {
				var r generated.UserRoleAssignment
				if err := r.FromPredefinedUserRoleAssignment(generated.PredefinedUserRoleAssignment{
					Name: generated.UserRoleMEMBER,
					Type: generated.UserRoleAssignmentTypePREDEFINED,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.UserRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.PredefinedUserRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected PredefinedUserRoleAssignmentRequest, got %T", v)
				}
				if req.Name != generated.UserRoleMEMBER {
					t.Errorf("Name = %v, want MEMBER", req.Name)
				}
			},
		},
		{
			name: "custom role",
			buildInput: func() generated.UserRoleAssignment {
				var r generated.UserRoleAssignment
				if err := r.FromCustomUserRoleAssignment(generated.CustomUserRoleAssignment{
					Id:   "user-role-abc",
					Type: generated.UserRoleAssignmentTypeCUSTOM,
				}); err != nil {
					t.Fatal(err)
				}
				return r
			},
			checkResult: func(t *testing.T, got generated.UserRoleAssignmentRequest) {
				v, err := got.ValueByDiscriminator()
				if err != nil {
					t.Fatalf("ValueByDiscriminator: %v", err)
				}
				req, ok := v.(generated.CustomUserRoleAssignmentRequest)
				if !ok {
					t.Fatalf("expected CustomUserRoleAssignmentRequest, got %T", v)
				}
				if req.Id != "user-role-abc" {
					t.Errorf("Id = %v, want user-role-abc", req.Id)
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := AccountRoleAssignmentRequest(tt.buildInput())
			if (err != nil) != tt.wantErr {
				t.Fatalf("AccountRoleAssignmentRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				tt.checkResult(t, got)
			}
		})
	}
}
