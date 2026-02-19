package validate

import (
	"slices"

	"github.com/cedar-policy/cedar-go/types"
)

// requestEnv represents the type environment for type checking a policy condition.
type requestEnv struct {
	principalType types.EntityType
	actionUID     types.EntityUID
	resourceType  types.EntityType
	contextType   typeRecord
}

// generateRequestEnvs builds request environments from the schema for all action/principal/resource combos.
func (v *Validator) generateRequestEnvs() []requestEnv {
	var envs []requestEnv
	for uid, action := range v.schema.Actions {
		if action.AppliesTo == nil {
			continue
		}
		ctx := schemaRecordToCedarType(action.AppliesTo.Context)
		for _, pt := range action.AppliesTo.Principals {
			for _, rt := range action.AppliesTo.Resources {
				envs = append(envs, requestEnv{
					principalType: pt,
					actionUID:     uid,
					resourceType:  rt,
					contextType:   ctx,
				})
			}
		}
	}
	return envs
}

// filterEnvsForPolicy filters request environments to only those that match the policy's scope constraints.
func (v *Validator) filterEnvsForPolicy(envs []requestEnv, principalTypes, resourceTypes []types.EntityType, actionUIDs []types.EntityUID) []requestEnv {
	var filtered []requestEnv
	for _, env := range envs {
		if !matchesPrincipalConstraint(env.principalType, principalTypes) {
			continue
		}
		if !matchesResourceConstraint(env.resourceType, resourceTypes) {
			continue
		}
		if !v.matchesActionConstraint(env.actionUID, actionUIDs) {
			continue
		}
		filtered = append(filtered, env)
	}
	return filtered
}

func matchesPrincipalConstraint(pt types.EntityType, constraints []types.EntityType) bool {
	if len(constraints) == 0 {
		return true // ScopeTypeAll
	}
	return slices.Contains(constraints, pt)
}

func matchesResourceConstraint(rt types.EntityType, constraints []types.EntityType) bool {
	if len(constraints) == 0 {
		return true
	}
	return slices.Contains(constraints, rt)
}

func (v *Validator) matchesActionConstraint(actionUID types.EntityUID, constraints []types.EntityUID) bool {
	if len(constraints) == 0 {
		return true
	}
	for _, c := range constraints {
		if actionUID == c {
			return true
		}
		// Check if actionUID is in the action group of c
		if v.isActionInGroup(actionUID, c) {
			return true
		}
	}
	return false
}

// isActionInGroup checks if actionUID is a descendant of groupUID in the action hierarchy.
func (v *Validator) isActionInGroup(actionUID, groupUID types.EntityUID) bool {
	action, ok := v.schema.Actions[actionUID]
	if !ok {
		return false
	}
	for parent := range action.Entity.Parents.All() {
		if parent == groupUID {
			return true
		}
		if v.isActionInGroup(parent, groupUID) {
			return true
		}
	}
	return false
}
