package validate

import (
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// requestEnv represents the type environment for type checking a policy condition.
type requestEnv struct {
	principalType types.EntityType
	actionUID     types.EntityUID
	resourceType  types.EntityType
	contextType   typeRecord
}

// generateRequestEnvs builds request environments from the schema for all action/principal/resource combos.
func generateRequestEnvs(s *resolved.Schema) []requestEnv {
	var envs []requestEnv
	for uid, action := range s.Actions {
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
func filterEnvsForPolicy(s *resolved.Schema, envs []requestEnv, principalTypes, resourceTypes []types.EntityType, actionUIDs []types.EntityUID) []requestEnv {
	var filtered []requestEnv
	for _, env := range envs {
		if !matchesPrincipalConstraint(env.principalType, principalTypes) {
			continue
		}
		if !matchesResourceConstraint(env.resourceType, resourceTypes) {
			continue
		}
		if !matchesActionConstraint(s, env.actionUID, actionUIDs) {
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

func matchesActionConstraint(s *resolved.Schema, actionUID types.EntityUID, constraints []types.EntityUID) bool {
	if len(constraints) == 0 {
		return true
	}
	for _, c := range constraints {
		if actionUID == c {
			return true
		}
		// Check if actionUID is in the action group of c
		if isActionInGroup(s, actionUID, c) {
			return true
		}
	}
	return false
}

// isActionInGroup checks if actionUID is a descendant of groupUID in the action hierarchy.
func isActionInGroup(s *resolved.Schema, actionUID, groupUID types.EntityUID) bool {
	action, ok := s.Actions[actionUID]
	if !ok {
		return false
	}
	for parent := range action.Entity.Parents.All() {
		if parent == groupUID {
			return true
		}
		if isActionInGroup(s, parent, groupUID) {
			return true
		}
	}
	return false
}
