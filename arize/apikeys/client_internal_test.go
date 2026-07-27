package apikeys

import (
	"errors"
	"testing"
)

type errorRole struct{}

func (errorRole) MarshalJSON() ([]byte, error) {
	return nil, errors.New("invalid role")
}

func TestIsZeroRoleReturnsFalseOnMarshalError(t *testing.T) {
	if isZeroRole(errorRole{}) {
		t.Fatal("isZeroRole should return false when MarshalJSON fails")
	}
}
