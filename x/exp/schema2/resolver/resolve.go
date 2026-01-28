package resolver

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
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
func Resolve(s *ast.Schema) (*ResolvedSchema, error) {
	resolved := &ResolvedSchema{
		Namespaces: make(map[types.Path]ResolvedNamespace),
		Entities:   make(map[types.EntityType]ResolvedEntity),
		Enums:      make(map[types.EntityType]ResolvedEnum),
		Actions:    make(map[types.EntityUID]ResolvedAction),
	}

	rd := newResolveData(s)

	// Process top-level common types (resolve but don't add to output)
	for _, ct := range s.CommonTypes {
		_ = resolveCommonTypeNode(rd, ct)
	}

	// Process top-level entities
	for entityName, entityNode := range s.Entities {
		resolvedEntity := resolveEntityNode(rd, entityNode, entityName)
		// No need to check for enum conflicts here since enums are processed after entities
		resolved.Entities[resolvedEntity.Name] = resolvedEntity
	}

	// Process top-level enums
	for enumName, enumNode := range s.Enums {
		resolvedEnum := resolveEnumNode(rd, enumNode, enumName)
		// Check for conflicts with entities
		if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
			return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
		}
		resolved.Enums[resolvedEnum.Name] = resolvedEnum
	}

	// Process top-level actions
	for actionID, actionNode := range s.Actions {
		resolvedAction, err := resolveActionNode(rd, actionNode, actionID)
		if err != nil {
			return nil, err
		}
		actionUID := types.NewEntityUID("Action", actionID)
		resolved.Actions[actionUID] = resolvedAction
	}

	// Process namespaces
	for nsPath, ns := range s.Namespaces {
		// Store namespace annotations
		resolved.Namespaces[nsPath] = ResolvedNamespace{
			Name:        nsPath,
			Annotations: ns.Annotations,
		}

		// Create resolve data with this namespace
		nsRd := rd.withNamespace(nsPath)

		// Process namespace common types
		for _, ct := range ns.CommonTypes {
			_ = resolveCommonTypeNode(nsRd, ct)
		}

		// Process namespace entities
		for entityName, entityNode := range ns.Entities {
			qualifiedName := types.EntityType(string(nsPath) + "::" + string(entityName))
			resolvedEntity := resolveEntityNode(nsRd, entityNode, qualifiedName)
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
			resolvedEnum := resolveEnumNode(nsRd, enumNode, qualifiedName)
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
			resolvedAction, err := resolveActionNode(nsRd, actionNode, actionID)
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
