package validate

import (
	"errors"
	"fmt"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Policy validates a policy against the schema, performing scope validation
// and expression type checking.
func (v *Validator) Policy(policy *ast.Policy) error {
	var errs []error

	// RBAC scope validation
	principalTypes, err := v.validatePrincipalScope(policy.Principal)
	if err != nil {
		errs = append(errs, fmt.Errorf("principal scope: %w", err))
	}

	// Validate action scope: check that actions exist and get the full set
	// of actions (including descendants) for action application checking
	actionUIDs, err := v.validateAndGetActionUIDs(policy.Action)
	if err != nil {
		errs = append(errs, fmt.Errorf("action scope: %w", err))
	}

	resourceTypes, err := v.validateResourceScope(policy.Resource)
	if err != nil {
		errs = append(errs, fmt.Errorf("resource scope: %w", err))
	}

	// Check action application (runs even if there were action scope errors)
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
// It returns both the types and any validation errors (types may be empty even if there's an error).
func (v *Validator) validatePrincipalScope(scope ast.IsPrincipalScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil // matches any type
	case ast.ScopeTypeEq:
		entityTypes, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIn:
		_, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		entityTypes, err := v.validateScopeType(sc.Type)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	default:
		// ast.ScopeTypeIsIn is the only remaining case
		isIn := scope.(ast.ScopeTypeIsIn)
		entityTypes, err := v.validateScopeType(isIn.Type)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		if _, err := v.validateScopeEntity(isIn.Entity); err != nil {
			return []types.EntityType{}, err
		}
		// Check if the "is" type can actually be "in" the "in" entity's type
		// by intersecting with types that can be in the entity
		typesIn := v.getEntityTypesIn(isIn.Entity.Type)
		if slices.Contains(typesIn, isIn.Type) {
			return entityTypes, nil
		}
		// Type mismatch: return empty list (no error, but action application will fail)
		return []types.EntityType{}, nil
	}
}

// validateAndGetActionUIDs validates that actions in the scope exist in the schema,
// and returns the full set of action UIDs (including descendants) for further validation.
// It returns both validation errors and the action UIDs (which may include non-existent actions).
func (v *Validator) validateAndGetActionUIDs(scope ast.IsActionScopeNode) ([]types.EntityUID, error) {
	var errs []error
	var actionUIDs []types.EntityUID

	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil // matches any action
	case ast.ScopeTypeEq:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			errs = append(errs, fmt.Errorf("action %s not found in schema", sc.Entity))
		}
		actionUIDs = []types.EntityUID{sc.Entity}
	case ast.ScopeTypeIn:
		if _, ok := v.schema.Actions[sc.Entity]; !ok {
			errs = append(errs, fmt.Errorf("action %s not found in schema", sc.Entity))
		}
		actionUIDs = v.getActionsInSet([]types.EntityUID{sc.Entity})
	case ast.ScopeTypeInSet:
		for _, uid := range sc.Entities {
			if _, ok := v.schema.Actions[uid]; !ok {
				errs = append(errs, fmt.Errorf("action %s not found in schema", uid))
			}
		}
		actionUIDs = v.getActionsInSet(sc.Entities)
	}

	return actionUIDs, errors.Join(errs...)
}

// validateResourceScope validates the resource scope and returns the entity types it constrains to.
// It returns both the types and any validation errors (types may be empty even if there's an error).
func (v *Validator) validateResourceScope(scope ast.IsResourceScopeNode) ([]types.EntityType, error) {
	switch sc := scope.(type) {
	case ast.ScopeTypeAll:
		return nil, nil
	case ast.ScopeTypeEq:
		entityTypes, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	case ast.ScopeTypeIn:
		_, err := v.validateScopeEntity(sc.Entity)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return v.getEntityTypesIn(sc.Entity.Type), nil
	case ast.ScopeTypeIs:
		entityTypes, err := v.validateScopeType(sc.Type)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		return entityTypes, nil
	default:
		// ast.ScopeTypeIsIn is the only remaining case
		isIn := scope.(ast.ScopeTypeIsIn)
		entityTypes, err := v.validateScopeType(isIn.Type)
		if err != nil {
			// Return empty list so action application check can still run
			return []types.EntityType{}, err
		}
		if _, err := v.validateScopeEntity(isIn.Entity); err != nil {
			return []types.EntityType{}, err
		}
		// Check if the "is" type can actually be "in" the "in" entity's type
		// by intersecting with types that can be in the entity
		typesIn := v.getEntityTypesIn(isIn.Entity.Type)
		if slices.Contains(typesIn, isIn.Type) {
			return entityTypes, nil
		}
		// Type mismatch: return empty list (no error, but action application will fail)
		return []types.EntityType{}, nil
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

	// Collect relevant actions (actionUIDs already includes descendants for `in` scopes)
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

// getActionsInSet returns all action UIDs that are in the given set of actions,
// including the actions themselves and all their descendants.
func (v *Validator) getActionsInSet(uids []types.EntityUID) []types.EntityUID {
	result := make([]types.EntityUID, 0, len(uids))
	for _, uid := range uids {
		result = append(result, uid)
		for aUID := range v.schema.Actions {
			if aUID == uid {
				continue
			}
			if v.isActionDescendant(aUID, uid) {
				result = append(result, aUID)
			}
		}
	}
	return result
}

// isActionDescendant checks if actionUID is a descendant of ancestorUID.
func (v *Validator) isActionDescendant(actionUID, ancestorUID types.EntityUID) bool {
	action := v.schema.Actions[actionUID]
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
				// If err is a joined error, flatten it to avoid losing errors
				// when wrapping with fmt.Errorf
				if ue, ok := err.(interface{ Unwrap() []error }); ok {
					for _, e := range ue.Unwrap() {
						errs = append(errs, fmt.Errorf("condition %d: %w", i, e))
					}
				} else {
					errs = append(errs, fmt.Errorf("condition %d: %w", i, err))
				}
			}
			if t != nil && !isBoolType(t) {
				errs = append(errs, fmt.Errorf("condition %d: expected boolean type, got %T", i, t))
			}
		}
	}
	return errors.Join(errs...)
}
