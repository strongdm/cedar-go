package validate

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Policy validates a policy against the schema, performing RBAC scope validation
// and expression type checking.
func (v *Validator) Policy(policy *ast.Policy) error {
	var errs []error

	// RBAC scope validation
	principalTypes, err := v.validatePrincipalScope(policy.Principal)
	if err != nil {
		errs = append(errs, fmt.Errorf("principal scope: %w", err))
	}

	actionUIDs, err := v.validateActionScope(policy.Action)
	if err != nil {
		errs = append(errs, fmt.Errorf("action scope: %w", err))
	}

	resourceTypes, err := v.validateResourceScope(policy.Resource)
	if err != nil {
		errs = append(errs, fmt.Errorf("resource scope: %w", err))
	}

	// Check action application
	if err := v.validateActionApplication(principalTypes, resourceTypes, actionUIDs); err != nil {
		errs = append(errs, err)
	}

	// Expression type checking
	allEnvs := v.generateRequestEnvs()
	envs := v.filterEnvsForPolicy(allEnvs, principalTypes, resourceTypes, actionUIDs)

	if len(envs) > 0 && len(policy.Conditions) > 0 {
		if err := v.typecheckConditions(envs, policy.Conditions); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

// validatePrincipalScope validates the principal scope and returns the entity types it constrains to.
func (v *Validator) validatePrincipalScope(scope ast.IsPrincipalScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil // matches any type
	case ast.ScopeTypeEq:
		return v.validateScopeEntity(sc.Entity)
	case ast.ScopeTypeIn:
		if _, err := v.validateScopeEntity(sc.Entity); err != nil {
			return nil, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		return v.validateScopeType(sc.Type)
	case ast.ScopeTypeIsIn:
		types, err := v.validateScopeType(sc.Type)
		if err != nil {
			return nil, err
		}
		if _, err := v.validateScopeEntity(sc.Entity); err != nil {
			return nil, err
		}
		if err := v.validateIsInScope(sc.Type, sc.Entity.Type); err != nil {
			return nil, err
		}
		return types, nil
	default:
		return nil, fmt.Errorf("unknown principal scope type %T", scope)
	}
}

// validateActionScope validates the action scope and returns the action UIDs it constrains to.
func (v *Validator) validateActionScope(scope ast.IsActionScopeNode) ([]types.EntityUID, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil // matches any action
	case ast.ScopeTypeEq:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			return nil, fmt.Errorf("action %s not found in schema", sc.Entity)
		}
		return []types.EntityUID{sc.Entity}, nil
	case ast.ScopeTypeIn:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			return nil, fmt.Errorf("action %s not found in schema", sc.Entity)
		}
		return []types.EntityUID{sc.Entity}, nil
	case ast.ScopeTypeInSet:
		uids := make([]types.EntityUID, 0, len(sc.Entities))
		for _, uid := range sc.Entities {
			if _, ok := v.schema.Actions[uid]; !ok {
				return nil, fmt.Errorf("action %s not found in schema", uid)
			}
			uids = append(uids, uid)
		}
		return uids, nil
	default:
		return nil, fmt.Errorf("unknown action scope type %T", scope)
	}
}

// validateResourceScope validates the resource scope and returns the entity types it constrains to.
func (v *Validator) validateResourceScope(scope ast.IsResourceScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil
	case ast.ScopeTypeEq:
		return v.validateScopeEntity(sc.Entity)
	case ast.ScopeTypeIn:
		if _, err := v.validateScopeEntity(sc.Entity); err != nil {
			return nil, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		return v.validateScopeType(sc.Type)
	case ast.ScopeTypeIsIn:
		types, err := v.validateScopeType(sc.Type)
		if err != nil {
			return nil, err
		}
		if _, err := v.validateScopeEntity(sc.Entity); err != nil {
			return nil, err
		}
		if err := v.validateIsInScope(sc.Type, sc.Entity.Type); err != nil {
			return nil, err
		}
		return types, nil
	default:
		return nil, fmt.Errorf("unknown resource scope type %T", scope)
	}
}

func (v *Validator) validateScopeEntity(uid types.EntityUID) ([]types.EntityType, error) {
	et := uid.Type
	if _, ok := v.schema.Entities[et]; ok {
		return []types.EntityType{et}, nil
	}
	if schemaEnum, ok := v.schema.Enums[et]; ok {
		if !isValidEnumID(uid, schemaEnum) {
			return nil, fmt.Errorf("invalid enum value %q for type %q", uid.ID, et)
		}
		return []types.EntityType{et}, nil
	}
	if isActionEntity(et) {
		if _, ok := v.schema.Actions[uid]; ok {
			return []types.EntityType{et}, nil
		}
	}
	return nil, fmt.Errorf("entity type %q not found in schema", et)
}

func (v *Validator) validateScopeType(et types.EntityType) ([]types.EntityType, error) {
	if _, ok := v.schema.Entities[et]; ok {
		return []types.EntityType{et}, nil
	}
	if _, ok := v.schema.Enums[et]; ok {
		return []types.EntityType{et}, nil
	}
	return nil, fmt.Errorf("entity type %q not found in schema", et)
}

// validateActionApplication checks that at least one action's AppliesTo intersects
// the policy's principal AND resource constraints.
func (v *Validator) validateActionApplication(principalTypes, resourceTypes []types.EntityType, actionUIDs []types.EntityUID) error {
	// If we have no constraints on anything, it's valid
	if principalTypes == nil && resourceTypes == nil && actionUIDs == nil {
		return nil
	}

	// Collect relevant actions
	var actions []resolved.Action
	if actionUIDs == nil {
		for _, a := range v.schema.Actions {
			actions = append(actions, a)
		}
	} else {
		for _, uid := range actionUIDs {
			if a, ok := v.schema.Actions[uid]; ok {
				actions = append(actions, a)
			}
			// Also include actions that are descendants of the specified actions
			for aUID, a := range v.schema.Actions {
				if aUID == uid {
					continue
				}
				if v.isActionDescendant(aUID, uid) {
					actions = append(actions, a)
				}
			}
		}
	}

	for _, action := range actions {
		if action.AppliesTo == nil {
			continue
		}
		principalMatch := principalTypes == nil
		if !principalMatch {
			for _, pt := range principalTypes {
				if slices.Contains(action.AppliesTo.Principals, pt) {
					principalMatch = true
					break
				}
			}
		}
		resourceMatch := resourceTypes == nil
		if !resourceMatch {
			for _, rt := range resourceTypes {
				if slices.Contains(action.AppliesTo.Resources, rt) {
					resourceMatch = true
					break
				}
			}
		}
		if principalMatch && resourceMatch {
			return nil
		}
	}

	return fmt.Errorf("no action applies to the given principal and resource type constraints")
}

// isActionDescendant checks if actionUID is a descendant of ancestorUID.
func (v *Validator) isActionDescendant(actionUID, ancestorUID types.EntityUID) bool {
	action, ok := v.schema.Actions[actionUID]
	if !ok {
		return false
	}
	for parent := range action.Entity.Parents.All() {
		if parent == ancestorUID {
			return true
		}
		if v.isActionDescendant(parent, ancestorUID) {
			return true
		}
	}
	return false
}

// validateIsInScope checks that the `is` type can actually be "in" the `in` entity's type.
// For `principal is X in Y::""`, X must be a type that can be a descendant of Y's type.
func (v *Validator) validateIsInScope(isType, inType types.EntityType) error {
	// Collect all entity types that can be "in" inType (i.e., descendants + itself)
	typesIn := v.getEntityTypesIn(inType)
	if !slices.Contains(typesIn, isType) {
		return fmt.Errorf("entity type %q can never be a member of entity type %q", isType, inType)
	}
	return nil
}

// getEntityTypesIn returns all entity types that can be "in" (descendants of) the given entity type,
// including the type itself.
func (v *Validator) getEntityTypesIn(target types.EntityType) []types.EntityType {
	result := []types.EntityType{target}
	// Find all entity types whose ParentTypes include the target (direct children)
	for name, entity := range v.schema.Entities {
		if slices.Contains(entity.ParentTypes, target) {
			result = append(result, name)
		}
	}
	// Transitive closure: find types whose parents include types already in result
	changed := true
	for changed {
		changed = false
		for name, entity := range v.schema.Entities {
			if slices.Contains(result, name) {
				continue
			}
			for _, parent := range entity.ParentTypes {
				if slices.Contains(result, parent) {
					result = append(result, name)
					changed = true
					break
				}
			}
		}
	}
	return result
}

func (v *Validator) typecheckConditions(envs []requestEnv, conditions []ast.ConditionType) error {
	var errs []error
	for _, env := range envs {
		for i, cond := range conditions {
			caps := newCapabilitySet()
			t, _, err := v.typeOfExpr(&env, cond.Body, caps)
			if err != nil {
				errs = append(errs, fmt.Errorf("condition %d: %w", i, err))
				continue
			}
			if !isBoolType(t) {
				errs = append(errs, fmt.Errorf("condition %d: expected boolean type, got %T", i, t))
			}
		}
	}
	return errors.Join(errs...)
}
