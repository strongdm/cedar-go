package validate

import (
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
)

// Request validates a request against the schema.
func (v *Validator) Request(req types.Request) error {
	// Look up action
	action, ok := v.schema.Actions[req.Action]
	if !ok {
		return fmt.Errorf("action %s not found in schema", req.Action)
	}

	if action.AppliesTo == nil {
		return fmt.Errorf("action %s has no appliesTo", req.Action)
	}

	// Validate principal type
	if err := v.validateRequestEntityType(req.Principal, "principal"); err != nil {
		return err
	}
	if !slices.Contains(action.AppliesTo.Principals, req.Principal.Type) {
		return fmt.Errorf("principal type %q not valid for action %s", req.Principal.Type, req.Action)
	}

	// Validate resource type
	if err := v.validateRequestEntityType(req.Resource, "resource"); err != nil {
		return err
	}
	if !slices.Contains(action.AppliesTo.Resources, req.Resource.Type) {
		return fmt.Errorf("resource type %q not valid for action %s", req.Resource.Type, req.Action)
	}

	// Validate enum IDs for principal/resource
	if schemaEnum, ok := v.schema.Enums[req.Principal.Type]; ok {
		if !isValidEnumID(req.Principal, schemaEnum) {
			return fmt.Errorf("invalid enum ID %q for principal type %q", req.Principal.ID, req.Principal.Type)
		}
	}
	if schemaEnum, ok := v.schema.Enums[req.Resource.Type]; ok {
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

func (v *Validator) validateRequestEntityType(uid types.EntityUID, role string) error {
	et := uid.Type
	if _, ok := v.schema.Entities[et]; ok {
		return nil
	}
	if _, ok := v.schema.Enums[et]; ok {
		return nil
	}
	return fmt.Errorf("%s type %q not found in schema", role, et)
}
