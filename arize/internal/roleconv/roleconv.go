package roleconv

import (
	"fmt"

	"github.com/Arize-ai/client-go-v2/arize/internal/generated"
)

// IsZero reports whether a union role value is unset. The generated union types
// store their discriminated value as json.RawMessage; a nil payload marshals to
// "null".
func IsZero(r interface{ MarshalJSON() ([]byte, error) }) bool {
	b, err := r.MarshalJSON()
	return err == nil && string(b) == "null"
}

// SpaceRoleAssignmentRequest converts a SpaceRoleAssignment (response
// discriminated union) into the request variant required by write endpoints.
func SpaceRoleAssignmentRequest(role generated.SpaceRoleAssignment) (generated.SpaceRoleAssignmentRequest, error) {
	value, err := role.ValueByDiscriminator()
	if err != nil {
		return generated.SpaceRoleAssignmentRequest{}, err
	}
	var request generated.SpaceRoleAssignmentRequest
	switch assignment := value.(type) {
	case generated.PredefinedRoleAssignment:
		err = request.FromPredefinedRoleAssignmentRequest(generated.PredefinedRoleAssignmentRequest{Name: assignment.Name, Type: assignment.Type})
	case generated.CustomRoleAssignment:
		err = request.FromCustomRoleAssignmentRequest(generated.CustomRoleAssignmentRequest{Id: assignment.Id, Type: assignment.Type})
	default:
		err = fmt.Errorf("unsupported role assignment %T", value)
	}
	return request, err
}

// OrgRoleAssignmentRequest converts an OrganizationRoleAssignment into the
// request variant required by write endpoints.
func OrgRoleAssignmentRequest(role generated.OrganizationRoleAssignment) (generated.OrganizationRoleAssignmentRequest, error) {
	value, err := role.ValueByDiscriminator()
	if err != nil {
		return generated.OrganizationRoleAssignmentRequest{}, err
	}
	var request generated.OrganizationRoleAssignmentRequest
	switch assignment := value.(type) {
	case generated.OrganizationPredefinedRoleAssignment:
		err = request.FromOrganizationPredefinedRoleAssignmentRequest(generated.OrganizationPredefinedRoleAssignmentRequest{
			Name: assignment.Name,
			Type: assignment.Type,
		})
	case generated.OrganizationCustomRoleAssignment:
		err = request.FromOrganizationCustomRoleAssignmentRequest(generated.OrganizationCustomRoleAssignmentRequest{
			Id:   assignment.Id,
			Type: assignment.Type,
		})
	default:
		err = fmt.Errorf("unsupported role assignment %T", value)
	}
	return request, err
}

// AccountRoleAssignmentRequest converts a UserRoleAssignment into the request
// variant required by write endpoints.
func AccountRoleAssignmentRequest(role generated.UserRoleAssignment) (generated.UserRoleAssignmentRequest, error) {
	value, err := role.ValueByDiscriminator()
	if err != nil {
		return generated.UserRoleAssignmentRequest{}, err
	}
	var request generated.UserRoleAssignmentRequest
	switch assignment := value.(type) {
	case generated.PredefinedUserRoleAssignment:
		err = request.FromPredefinedUserRoleAssignmentRequest(generated.PredefinedUserRoleAssignmentRequest{
			Name: assignment.Name,
			Type: assignment.Type,
		})
	case generated.CustomUserRoleAssignment:
		err = request.FromCustomUserRoleAssignmentRequest(generated.CustomUserRoleAssignmentRequest{
			Id:   assignment.Id,
			Type: assignment.Type,
		})
	default:
		err = fmt.Errorf("unsupported role assignment %T", value)
	}
	return request, err
}
