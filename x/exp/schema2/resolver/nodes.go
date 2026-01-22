package resolver

import (
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// resolve returns a new CommonTypeNode with all type references resolved.
func resolveCommonTypeNode(rd *resolveData, c ast.CommonTypeNode) (ast.CommonTypeNode, error) {
	resolvedType, err := resolveType(rd, c.Type)
	if err != nil {
		return ast.CommonTypeNode{}, err
	}
	return ast.CommonTypeNode{
		Name:        c.Name,
		Annotations: c.Annotations,
		Type:        resolvedType,
	}, nil
}

// resolve returns a ResolvedEntity with all type references resolved and name fully qualified.
func resolveEntityNode(rd *resolveData, e ast.EntityNode) (ResolvedEntity, error) {
	// Qualify the entity name with namespace if present
	name := e.Name
	if rd.namespace != nil && rd.namespace.Name != "" {
		name = types.EntityType(string(rd.namespace.Name) + "::" + string(e.Name))
	}

	resolved := ResolvedEntity{
		Name:        name,
		Annotations: e.Annotations,
	}

	// Resolve and convert MemberOf references from []EntityTypeRef to []types.EntityType
	if len(e.MemberOfVal) > 0 {
		resolved.MemberOf = make([]types.EntityType, len(e.MemberOfVal))
		for i, ref := range e.MemberOfVal {
			resolvedRef := resolveEntityTypeRef(rd, ref)
			resolved.MemberOf[i] = resolvedRef.Name
		}
	}

	// Resolve Shape
	if e.ShapeVal != nil {
		resolvedShape, err := resolveRecord(rd, *e.ShapeVal)
		if err != nil {
			return ResolvedEntity{}, err
		}
		resolved.Shape = &resolvedShape
	}

	// Resolve Tags
	if e.TagsVal != nil {
		resolvedTags, err := resolveType(rd, e.TagsVal)
		if err != nil {
			return ResolvedEntity{}, err
		}
		resolved.Tags = resolvedTags
	}

	return resolved, nil
}

// resolve returns a ResolvedEnum with name fully qualified.
func resolveEnumNode(rd *resolveData, e ast.EnumNode) ResolvedEnum {
	// Qualify the enum name with namespace if present
	name := e.Name
	if rd.namespace != nil && rd.namespace.Name != "" {
		name = types.EntityType(string(rd.namespace.Name) + "::" + string(e.Name))
	}
	return ResolvedEnum{
		Name:        name,
		Annotations: e.Annotations,
		Values:      e.Values,
	}
}

// resolve returns a ResolvedAction with all type references resolved and converted to types.EntityType and types.EntityUID.
func resolveActionNode(rd *resolveData, a ast.ActionNode) (ResolvedAction, error) {
	resolved := ResolvedAction{
		Name:        a.Name,
		Annotations: a.Annotations,
	}

	// Resolve and convert MemberOf references from []EntityRef to []types.EntityUID
	if len(a.MemberOfVal) > 0 {
		resolved.MemberOf = make([]types.EntityUID, len(a.MemberOfVal))
		for i, ref := range a.MemberOfVal {
			resolvedType := resolveEntityTypeRef(rd, ref.Type)
			resolved.MemberOf[i] = types.NewEntityUID(resolvedType.Name, ref.ID)
		}
	}

	// Resolve and convert AppliesTo
	if a.AppliesToVal != nil {
		resolved.AppliesTo = &ResolvedAppliesTo{}

		// Convert PrincipalTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesToVal.PrincipalTypes) > 0 {
			resolved.AppliesTo.PrincipalTypes = make([]types.EntityType, len(a.AppliesToVal.PrincipalTypes))
			for i, ref := range a.AppliesToVal.PrincipalTypes {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.PrincipalTypes[i] = resolvedRef.Name
			}
		}

		// Convert ResourceTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesToVal.ResourceTypes) > 0 {
			resolved.AppliesTo.ResourceTypes = make([]types.EntityType, len(a.AppliesToVal.ResourceTypes))
			for i, ref := range a.AppliesToVal.ResourceTypes {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.ResourceTypes[i] = resolvedRef.Name
			}
		}

		// Resolve Context type
		if a.AppliesToVal.Context != nil {
			resolvedContext, err := resolveType(rd, a.AppliesToVal.Context)
			if err != nil {
				return ResolvedAction{}, err
			}
			resolved.AppliesTo.Context = resolvedContext
		}
	}

	return resolved, nil
}
