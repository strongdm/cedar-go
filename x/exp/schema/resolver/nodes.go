package resolver

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// resolve returns a new CommonTypeNode with all type references resolved.
func resolveCommonType(rd *resolveData, c ast.CommonType) ast.CommonType {
	resolvedType := resolveType(rd, c.Type)
	return ast.CommonType{
		Annotations: c.Annotations,
		Type:        resolvedType,
	}
}

// resolve returns a ResolvedEntity with all type references resolved and name fully qualified.
func resolveEntity(rd *resolveData, e ast.Entity, name types.EntityType) Entity {
	resolved := Entity{
		Name:        name,
		Annotations: e.Annotations,
	}

	// Resolve and convert MemberOf references from []EntityTypeRef to []types.EntityType
	if len(e.MemberOf) > 0 {
		resolved.MemberOf = make([]types.EntityType, len(e.MemberOf))
		for i, ref := range e.MemberOf {
			resolvedRef := resolveEntityTypeRef(rd, ref)
			resolved.MemberOf[i] = resolvedRef.Name
		}
	}

	// Resolve Shape
	if e.Shape != nil {
		resolvedShape := resolveRecord(rd, *e.Shape)
		resolved.Shape = &resolvedShape
	}

	// Resolve Tags
	if e.Tags != nil {
		resolvedTags := resolveType(rd, e.Tags)
		resolved.Tags = resolvedTags
	}

	return resolved
}

// resolve returns a ResolvedEnum with name fully qualified.
func resolveEnum(rd *resolveData, e ast.Enum, name types.EntityType) Enum {
	return Enum{
		Name:        name,
		Annotations: e.Annotations,
		Values:      e.Values,
	}
}

// resolve returns a ResolvedAction with all type references resolved and converted to types.EntityType and types.EntityUID.
func resolveAction(rd *resolveData, a ast.Action, name types.String) (Action, error) {
	resolved := Action{
		Name:        name,
		Annotations: a.Annotations,
	}

	// Resolve and convert MemberOf references from []EntityRef to []types.EntityUID
	if len(a.MemberOf) > 0 {
		resolved.MemberOf = make([]types.EntityUID, len(a.MemberOf))
		for i, ref := range a.MemberOf {
			resolvedType := resolveEntityTypeRef(rd, ref.Type)
			resolved.MemberOf[i] = types.NewEntityUID(resolvedType.Name, ref.ID)
		}
	}

	// Resolve and convert AppliesTo
	if a.AppliesTo != nil {
		resolved.AppliesTo = &AppliesTo{}

		// Convert PrincipalTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesTo.PrincipalTypes) > 0 {
			resolved.AppliesTo.PrincipalTypes = make([]types.EntityType, len(a.AppliesTo.PrincipalTypes))
			for i, ref := range a.AppliesTo.PrincipalTypes {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.PrincipalTypes[i] = resolvedRef.Name
			}
		}

		// Convert ResourceTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesTo.ResourceTypes) > 0 {
			resolved.AppliesTo.ResourceTypes = make([]types.EntityType, len(a.AppliesTo.ResourceTypes))
			for i, ref := range a.AppliesTo.ResourceTypes {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.ResourceTypes[i] = resolvedRef.Name
			}
		}

		// Resolve Context type
		if a.AppliesTo.Context != nil {
			resolvedContext := resolveType(rd, a.AppliesTo.Context)
			recordContext, ok := resolvedContext.(ast.RecordType)
			if resolvedContext != nil && !ok {
				return Action{}, fmt.Errorf("action %q context resolved to %T", name, resolvedContext)
			}
			resolved.AppliesTo.Context = recordContext
		}
	}

	return resolved, nil
}
