package resolver

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// commonTypeEntry represents a common type that may or may not be resolved yet.
type commonTypeEntry struct {
	resolved bool
	node     ast.CommonTypeNode
}

// resolveData contains cached information for efficient type resolution.
type resolveData struct {
	schema               *ast.Schema
	namespace            *ast.NamespaceNode
	schemaCommonTypes    map[string]*commonTypeEntry // Fully qualified name -> common type entry
	namespaceCommonTypes map[string]*commonTypeEntry // Unqualified name -> common type entry
}

// entityExistsInEmptyNamespace checks if an entity with the given name exists in the empty namespace (global scope).
func (rd *resolveData) entityExistsInEmptyNamespace(name types.EntityType) bool {
	nameStr := string(name)
	for _, node := range rd.schema.Nodes {
		switch n := node.(type) {
		case ast.EntityNode:
			if string(n.Name) == nameStr {
				return true
			}
		case ast.EnumNode:
			if string(n.Name) == nameStr {
				return true
			}
		}
	}
	return false
}

// newResolveData creates a new resolveData with cached common types from the schema.
func newResolveData(schema *ast.Schema) *resolveData {
	rd := &resolveData{
		schema:               schema,
		namespace:            nil,
		schemaCommonTypes:    make(map[string]*commonTypeEntry),
		namespaceCommonTypes: make(map[string]*commonTypeEntry),
	}

	// Build schema-wide common types map (fully qualified names)
	// Populate with unresolved entries that will be resolved lazily
	for ns, ct := range schema.CommonTypes() {
		ctCopy := ct
		var fullName string
		if ns == nil {
			fullName = string(ct.Name)
		} else {
			fullName = string(ns.Name) + "::" + string(ct.Name)
		}
		rd.schemaCommonTypes[fullName] = &commonTypeEntry{
			resolved: false,
			node:     ctCopy,
		}
	}

	return rd
}

// withNamespace returns a new resolveData with the given namespace.
func (rd *resolveData) withNamespace(namespace *ast.NamespaceNode) *resolveData {
	if namespace == rd.namespace {
		return rd
	}

	// Create new namespace-local cache for the new namespace
	namespaceCommonTypes := make(map[string]*commonTypeEntry)
	if namespace != nil {
		for ct := range namespace.CommonTypes() {
			ctCopy := ct
			namespaceCommonTypes[string(ct.Name)] = &commonTypeEntry{
				resolved: false,
				node:     ctCopy,
			}
		}
	}

	return &resolveData{
		schema:               rd.schema,
		namespace:            namespace,
		schemaCommonTypes:    rd.schemaCommonTypes, // Reuse schema-wide cache
		namespaceCommonTypes: namespaceCommonTypes, // New namespace-specific cache
	}
}

// resolveDeclaration resolves a single declaration node and adds it to the resolved schema.
// It handles common types, entities, enums, and actions.
func resolveDeclaration(decl ast.IsDeclaration, rd *resolveData, resolved *ResolvedSchema, namespaceName types.Path) error {
	switch d := decl.(type) {
	case ast.CommonTypeNode:
		// Common types are resolved but not added to the maps
		_, err := resolveCommonTypeNode(rd, d)
		if err != nil {
			return err
		}

	case ast.EntityNode:
		resolvedEntity, err := resolveEntityNode(rd, d)
		if err != nil {
			return err
		}
		// Check for conflicts with existing entities or enums
		if _, exists := resolved.Entities[resolvedEntity.Name]; exists {
			return fmt.Errorf("entity type %q is defined multiple times", resolvedEntity.Name)
		}
		if _, exists := resolved.Enums[resolvedEntity.Name]; exists {
			return fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEntity.Name)
		}
		resolved.Entities[resolvedEntity.Name] = resolvedEntity

	case ast.EnumNode:
		resolvedEnum := resolveEnumNode(rd, d)
		// Check for conflicts with existing enums or entities
		if _, exists := resolved.Enums[resolvedEnum.Name]; exists {
			return fmt.Errorf("enum type %q is defined multiple times", resolvedEnum.Name)
		}
		if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
			return fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
		}
		resolved.Enums[resolvedEnum.Name] = resolvedEnum

	case ast.ActionNode:
		resolvedAction, err := resolveActionNode(rd, d)
		if err != nil {
			return err
		}
		// Construct EntityUID from qualified action type
		var actionType types.EntityType
		if namespaceName == "" {
			actionType = "Action"
		} else {
			actionType = types.EntityType(string(namespaceName) + "::Action")
		}
		actionUID := types.NewEntityUID(actionType, resolvedAction.Name)
		// Check for duplicate actions
		if _, exists := resolved.Actions[actionUID]; exists {
			return fmt.Errorf("action %q is defined multiple times", actionUID)
		}
		resolved.Actions[actionUID] = resolvedAction
	}
	return nil
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

	for _, node := range s.Nodes {
		switch n := node.(type) {
		case ast.NamespaceNode:
			// Store namespace annotations
			if n.Name != "" {
				resolved.Namespaces[n.Name] = ResolvedNamespace{
					Name:        n.Name,
					Annotations: n.Annotations,
				}
			}

			// Create resolve data with this namespace
			nsRd := rd.withNamespace(&n)

			// Resolve all declarations in the namespace
			for _, decl := range n.Declarations {
				if err := resolveDeclaration(decl, nsRd, resolved, n.Name); err != nil {
					return nil, err
				}
			}

		case ast.EntityNode:
			if err := resolveDeclaration(n, rd, resolved, ""); err != nil {
				return nil, err
			}

		case ast.EnumNode:
			if err := resolveDeclaration(n, rd, resolved, ""); err != nil {
				return nil, err
			}

		case ast.ActionNode:
			if err := resolveDeclaration(n, rd, resolved, ""); err != nil {
				return nil, err
			}

		case ast.CommonTypeNode:
			if err := resolveDeclaration(n, rd, resolved, ""); err != nil {
				return nil, err
			}
		}
	}

	return resolved, nil
}
