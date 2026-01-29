package resolver

import (
	"fmt"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// commonTypeEntry represents a common type that may or may not be resolved yet.
type commonTypeEntry struct {
	resolved bool
	in       ast.CommonType
	out      IsType
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
		rd.schemaCommonTypes[string(name)] = &commonTypeEntry{
			resolved: false,
			in:       ct,
		}
	}

	// Namespace common types (fully qualified)
	for nsPath, ns := range schema.Namespaces {
		for name, ct := range ns.CommonTypes {
			fullName := string(nsPath) + "::" + string(name)
			rd.schemaCommonTypes[fullName] = &commonTypeEntry{
				resolved: false,
				in:       ct,
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
				namespaceCommonTypes[string(name)] = &commonTypeEntry{
					resolved: false,
					in:       ct,
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

// Resolve converts an AST schema to a resolver.Schema with resolved types and indexed declarations.
//
// Common types are inlined. Entity and enum names are fully qualified.
// Namespace-relative references are resolved in context. Declarations are indexed in flat maps.
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
		_ = resolveType(rd, ct.Type)
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
			Annotations: Annotations(ns.Annotations),
		}

		// Create resolve data with this namespace
		nsRd := rd.withNamespace(nsPath)

		// Process namespace common types
		for _, ct := range ns.CommonTypes {
			_ = resolveType(nsRd, ct.Type)
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

// resolve returns a ResolvedEntity with all type references resolved and name fully qualified.
func resolveEntity(rd *resolveData, e ast.Entity, name types.EntityType) Entity {
	resolved := Entity{
		Name:        name,
		Annotations: Annotations(e.Annotations),
	}

	// Resolve and convert MemberOf references from []EntityTypeRef to []types.EntityType
	if len(e.MemberOf) > 0 {
		resolved.MemberOf = make([]types.EntityType, len(e.MemberOf))
		for i, ref := range e.MemberOf {
			resolvedRef := resolveEntityTypeRef(rd, ref)
			resolved.MemberOf[i] = types.EntityType(resolvedRef)
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
		Annotations: Annotations(e.Annotations),
		Values:      e.Values,
	}
}

// resolve returns a ResolvedAction with all type references resolved and converted to types.EntityType and types.EntityUID.
func resolveAction(rd *resolveData, a ast.Action, name types.String) (Action, error) {
	resolved := Action{
		Name:        name,
		Annotations: Annotations(a.Annotations),
	}

	// Resolve and convert MemberOf references from []EntityRef to []types.EntityUID
	if len(a.MemberOf) > 0 {
		resolved.MemberOf = make([]types.EntityUID, len(a.MemberOf))
		for i, ref := range a.MemberOf {
			resolvedType := resolveEntityTypeRef(rd, ref.Type)
			resolved.MemberOf[i] = types.NewEntityUID(types.EntityType(resolvedType), ref.ID)
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
				resolved.AppliesTo.Principals[i] = types.EntityType(resolvedRef)
			}
		}

		// Convert ResourceTypes from []EntityTypeRef to []types.EntityType
		if len(a.AppliesTo.Resources) > 0 {
			resolved.AppliesTo.Resources = make([]types.EntityType, len(a.AppliesTo.Resources))
			for i, ref := range a.AppliesTo.Resources {
				resolvedRef := resolveEntityTypeRef(rd, ref)
				resolved.AppliesTo.Resources[i] = types.EntityType(resolvedRef)
			}
		}

		// Resolve Context type
		if a.AppliesTo.Context != nil {
			resolvedContext := resolveType(rd, a.AppliesTo.Context)
			recordContext, ok := resolvedContext.(RecordType)
			if resolvedContext != nil && !ok {
				return Action{}, fmt.Errorf("action %q context resolved to %T", name, resolvedContext)
			}
			resolved.AppliesTo.Context = recordContext
		}
	}

	return resolved, nil
}

func resolveType(rd *resolveData, in ast.IsType) IsType {
	switch t := in.(type) {
	case ast.SetType:
		return resolveSet(rd, t)
	case ast.RecordType:
		return resolveRecord(rd, t)
	case ast.EntityTypeRef:
		return resolveEntityTypeRef(rd, t)
	case ast.TypeRef:
		return resolveTypeRef(rd, t)
	case ast.StringType:
		return StringType{}
	case ast.LongType:
		return LongType{}
	case ast.BoolType:
		return BoolType{}
	case ast.ExtensionType:
		return ExtensionType(t)
	default:
		panic(fmt.Sprintf("unknown type: %T", t))
	}
}

// resolve returns a new SetType with the element type resolved.
func resolveSet(rd *resolveData, s ast.SetType) SetType {
	resolved := resolveType(rd, s.Element)
	return SetType{Element: resolved}
}

// resolve returns a new RecordType with all attribute types resolved.
func resolveRecord(rd *resolveData, r ast.RecordType) RecordType {
	resolvedAttrs := make(RecordType)
	for key, attr := range r {
		resolvedType := resolveType(rd, attr.Type)
		resolvedAttrs[key] = Attribute{
			Type:        resolvedType,
			Optional:    attr.Optional,
			Annotations: Annotations(attr.Annotations),
		}
	}
	return RecordType(resolvedAttrs)
}

// willResolve resolves the entity type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it checks if the entity exists
// in the empty namespace first before qualifying it with the current namespace.
// This method never returns an error.
func resolveEntityTypeRef(rd *resolveData, e ast.EntityTypeRef) EntityType {
	if rd.namespacePath == "" {
		return EntityType(e)
	}

	name := string(e)
	// If already qualified (contains "::"), return as-is
	if strings.Contains(name, "::") || (len(name) > 0 && name[0] == ':') {
		return EntityType(e)
	}

	// Check if this entity exists in the empty namespace (global)
	if rd.entityExistsInEmptyNamespace(types.EntityType(e)) {
		// Keep it unqualified to reference the global entity
		return EntityType(e)
	}

	// Otherwise, qualify it with the current namespace
	return EntityType(types.EntityType(string(rd.namespacePath) + "::" + name))
}

// resolve resolves the type reference relative to the given namespace and schema.
// It searches for a matching CommonType in the namespace first, then in the entire schema.
// If found, it returns the resolved concrete type. Otherwise, it treats it as an EntityTypeRef.
func resolveTypeRef(rd *resolveData, t ast.TypeRef) IsType {
	name := string(t)

	// Try to find the type in the current namespace first (for unqualified names)
	if rd.namespacePath != "" && len(name) > 0 && name[0] != ':' && !strings.Contains(name, "::") {
		// Check namespace-local cache first
		if entry, found := rd.namespaceCommonTypes[name]; found {
			// If already resolved, return cached type
			if entry.resolved {
				return entry.out
			}
			// Resolve lazily
			resolvedType := resolveType(rd, entry.in.Type)
			// Cache the resolved node
			entry.out = resolvedType
			entry.resolved = true
			return resolvedType
		}
	}

	// Check schema-wide cache
	if entry, found := rd.schemaCommonTypes[name]; found {
		// If already resolved, return cached type
		if entry.resolved {
			return entry.out
		}
		// Resolve lazily with the common type's namespace context
		// Find the namespace for this common type by checking where it's declared
		var nsPath types.Path
		// Check top-level common types first
		if _, exists := rd.schema.CommonTypes[types.Ident(name)]; exists {
			nsPath = ""
		} else {
			// Check namespace common types
			for path, ns := range rd.schema.Namespaces {
				// Extract unqualified name from fully qualified name
				prefix := string(path) + "::"
				if strings.HasPrefix(name, prefix) {
					unqualifiedName := strings.TrimPrefix(name, prefix)
					if _, exists := ns.CommonTypes[types.Ident(unqualifiedName)]; exists {
						nsPath = path
						break
					}
				}
			}
		}
		ctRd := rd.withNamespace(nsPath)

		// Resolve lazily
		resolvedType := resolveType(ctRd, entry.in.Type)
		// Cache the resolved node
		entry.out = resolvedType
		entry.resolved = true
		return resolvedType
	}

	// Check for known extension types (with or without __cedar:: prefix)
	extensionName := name
	if strings.HasPrefix(name, "__cedar::") {
		extensionName = strings.TrimPrefix(name, "__cedar::")
	}
	if _, ok := knownExtensions[extensionName]; ok {
		return ExtensionType(types.Ident(extensionName))
	}

	// Not found, treat as EntityTypeRef
	return EntityType(types.EntityType(name))
}

var knownExtensions = map[string]struct{}{
	"decimal":  {},
	"duration": {},
	"datetime": {},
	"ipaddr":   {},
}
