package resolver

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// commonTypeEntry represents a common type that may or may not be resolved yet.
type commonTypeEntry struct {
	resolved bool
	node     ast.CommonType
}

// resolveData contains cached information for efficient type resolution.
type resolveData struct {
	schema               *ast.Schema
	namespacePath        types.Path
	schemaCommonTypes    map[string]*commonTypeEntry // Fully qualified name -> common type entry
	namespaceCommonTypes map[string]*commonTypeEntry // Unqualified name -> common type entry
}

// entityExistsInEmptyNamespace checks if an entity with the given name exists in the empty namespace (global scope).
func (rd *resolveData) entityExistsInEmptyNamespace(name types.EntityType) bool {
	// Check in top-level entities and enums
	if _, exists := rd.schema.Entities[name]; exists {
		return true
	}
	if _, exists := rd.schema.Enums[name]; exists {
		return true
	}
	return false
}

// newResolveData creates a new resolveData with cached common types from the schema.
func newResolveData(schema *ast.Schema) *resolveData {
	rd := &resolveData{
		schema:               schema,
		namespacePath:        "",
		schemaCommonTypes:    make(map[string]*commonTypeEntry),
		namespaceCommonTypes: make(map[string]*commonTypeEntry),
	}

	// Build schema-wide common types map (fully qualified names)
	// Top-level common types (unqualified)
	for name, ct := range schema.CommonTypes {
		ctCopy := ct
		rd.schemaCommonTypes[string(name)] = &commonTypeEntry{
			resolved: false,
			node:     ctCopy,
		}
	}

	// Namespace common types (fully qualified)
	for nsPath, ns := range schema.Namespaces {
		for name, ct := range ns.CommonTypes {
			ctCopy := ct
			fullName := string(nsPath) + "::" + string(name)
			rd.schemaCommonTypes[fullName] = &commonTypeEntry{
				resolved: false,
				node:     ctCopy,
			}
		}
	}

	return rd
}

// withNamespace returns a new resolveData with the given namespace.
func (rd *resolveData) withNamespace(namespacePath types.Path) *resolveData {
	if namespacePath == rd.namespacePath {
		return rd
	}

	// Create new namespace-local cache for the new namespace
	namespaceCommonTypes := make(map[string]*commonTypeEntry)
	if namespacePath != "" {
		if ns, exists := rd.schema.Namespaces[namespacePath]; exists {
			for name, ct := range ns.CommonTypes {
				ctCopy := ct
				namespaceCommonTypes[string(name)] = &commonTypeEntry{
					resolved: false,
					node:     ctCopy,
				}
			}
		}
	}

	return &resolveData{
		schema:               rd.schema,
		namespacePath:        namespacePath,
		schemaCommonTypes:    rd.schemaCommonTypes, // Reuse schema-wide cache
		namespaceCommonTypes: namespaceCommonTypes, // New namespace-specific cache
	}
}

// Resolve returns a ResolvedSchema with all type references resolved and indexed.
// Type references within namespaces are resolved relative to their namespace.
// Top-level type references are resolved as-is.
// Returns an error if any type reference cannot be resolved or if there are naming conflicts.
func Resolve(s *ast.Schema) (*Schema, error) {
	resolved := &Schema{
		Namespaces: make(map[types.Path]Namespace),
		Entities:   make(map[types.EntityType]Entity),
		Enums:      make(map[types.EntityType]Enum),
		Actions:    make(map[types.EntityUID]Action),
	}

	rd := newResolveData(s)

	// Process top-level common types (resolve but don't add to output)
	for _, ct := range s.CommonTypes {
		_ = resolveCommonType(rd, ct)
	}

	// Process top-level entities
	for entityName, entityNode := range s.Entities {
		resolvedEntity := resolveEntity(rd, entityNode, entityName)
		// No need to check for enum conflicts here since enums are processed after entities
		resolved.Entities[resolvedEntity.Name] = resolvedEntity
	}

	// Process top-level enums
	for enumName, enumNode := range s.Enums {
		resolvedEnum := resolveEnum(rd, enumNode, enumName)
		// Check for conflicts with entities
		if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
			return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
		}
		resolved.Enums[resolvedEnum.Name] = resolvedEnum
	}

	// Process top-level actions
	for actionID, actionNode := range s.Actions {
		resolvedAction, err := resolveAction(rd, actionNode, actionID)
		if err != nil {
			return nil, err
		}
		actionUID := types.NewEntityUID("Action", actionID)
		resolved.Actions[actionUID] = resolvedAction
	}

	// Process namespaces
	for nsPath, ns := range s.Namespaces {
		// Store namespace annotations
		resolved.Namespaces[nsPath] = Namespace{
			Name:        nsPath,
			Annotations: ns.Annotations,
		}

		// Create resolve data with this namespace
		nsRd := rd.withNamespace(nsPath)

		// Process namespace common types
		for _, ct := range ns.CommonTypes {
			_ = resolveCommonType(nsRd, ct)
		}

		// Process namespace entities
		for entityName, entityNode := range ns.Entities {
			qualifiedName := types.EntityType(string(nsPath) + "::" + string(entityName))
			resolvedEntity := resolveEntity(nsRd, entityNode, qualifiedName)
			// Check for conflicts
			if _, exists := resolved.Entities[resolvedEntity.Name]; exists {
				return nil, fmt.Errorf("entity type %q is defined multiple times", resolvedEntity.Name)
			}
			// No need to check for enum conflicts here since enums are processed after entities
			resolved.Entities[resolvedEntity.Name] = resolvedEntity
		}

		// Process namespace enums
		for enumName, enumNode := range ns.Enums {
			qualifiedName := types.EntityType(string(nsPath) + "::" + string(enumName))
			resolvedEnum := resolveEnum(nsRd, enumNode, qualifiedName)
			// Check for conflicts
			if _, exists := resolved.Enums[resolvedEnum.Name]; exists {
				return nil, fmt.Errorf("enum type %q is defined multiple times", resolvedEnum.Name)
			}
			if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
				return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
			}
			resolved.Enums[resolvedEnum.Name] = resolvedEnum
		}

		// Process namespace actions
		for actionID, actionNode := range ns.Actions {
			resolvedAction, err := resolveAction(nsRd, actionNode, actionID)
			if err != nil {
				return nil, err
			}
			actionType := types.EntityType(string(nsPath) + "::Action")
			actionUID := types.NewEntityUID(actionType, actionID)
			// No need to check for duplicate actions - map keys prevent duplicates within a namespace
			resolved.Actions[actionUID] = resolvedAction
		}
	}

	return resolved, nil
}

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
		if len(a.AppliesTo.Principals) > 0 {
			resolved.AppliesTo.Principals = make([]types.EntityType, len(a.AppliesTo.Principals))
			for i, ref := range a.AppliesTo.Principals {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.Principals[i] = resolvedRef.Name
			}
		}

		// Convert ResourceTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesTo.Resources) > 0 {
			resolved.AppliesTo.Resources = make([]types.EntityType, len(a.AppliesTo.Resources))
			for i, ref := range a.AppliesTo.Resources {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.Resources[i] = resolvedRef.Name
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
