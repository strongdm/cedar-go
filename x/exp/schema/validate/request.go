package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Request validates a request against the schema.
func Request(s *resolved.Schema, req types.Request) error {
	// Look up action
	action, ok := s.Actions[req.Action]
	if !ok {
		return fmt.Errorf("action %s not found in schema", req.Action)
	}

	if action.AppliesTo == nil {
		return fmt.Errorf("action %s has no appliesTo", req.Action)
	}

	// Validate principal type
	if err := validateRequestEntityType(s, req.Principal, "principal"); err != nil {
		return err
	}
	if !slices.Contains(action.AppliesTo.Principals, req.Principal.Type) {
		return fmt.Errorf("principal type %q not valid for action %s", req.Principal.Type, req.Action)
	}

	// Validate resource type
	if err := validateRequestEntityType(s, req.Resource, "resource"); err != nil {
		return err
	}
	if !slices.Contains(action.AppliesTo.Resources, req.Resource.Type) {
		return fmt.Errorf("resource type %q not valid for action %s", req.Resource.Type, req.Action)
	}

	// Validate enum IDs for principal/resource
	if schemaEnum, ok := s.Enums[req.Principal.Type]; ok {
		if !isValidEnumID(req.Principal, schemaEnum) {
			return fmt.Errorf("invalid enum ID %q for principal type %q", req.Principal.ID, req.Principal.Type)
		}
	}
	if schemaEnum, ok := s.Enums[req.Resource.Type]; ok {
		if !isValidEnumID(req.Resource, schemaEnum) {
			return fmt.Errorf("invalid enum ID %q for resource type %q", req.Resource.ID, req.Resource.Type)
		}
	}

	// Validate context
	if err := checkRecord(req.Context, action.AppliesTo.Context); err != nil {
		return fmt.Errorf("context: %w", err)
	}

	return nil
}

func validateRequestEntityType(s *resolved.Schema, uid types.EntityUID, role string) error {
	et := uid.Type
	if _, ok := s.Entities[et]; ok {
		return nil
	}
	if _, ok := s.Enums[et]; ok {
		return nil
	}
	return fmt.Errorf("%s type %q not found in schema", role, et)
}
