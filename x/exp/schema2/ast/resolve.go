package ast

import (
	"fmt"
	"iter"

	"github.com/cedar-policy/cedar-go/types"
)

// commonTypeEntry represents a common type that may or may not be resolved yet.
type commonTypeEntry struct {
	resolved bool
	node     CommonTypeNode
}

// resolveData contains cached information for efficient type resolution.
type resolveData struct {
	schema               *Schema
	namespace            *NamespaceNode
	schemaCommonTypes    map[string]*commonTypeEntry // Fully qualified name -> common type entry
	namespaceCommonTypes map[string]*commonTypeEntry // Unqualified name -> common type entry
}

// entityExistsInEmptyNamespace checks if an entity with the given name exists in the empty namespace (global scope).
func (rd *resolveData) entityExistsInEmptyNamespace(name types.EntityType) bool {
	if rd.schema == nil {
		return false
	}

	nameStr := string(name)
	for _, node := range rd.schema.Nodes {
		switch n := node.(type) {
		case EntityNode:
			if string(n.Name) == nameStr {
				return true
			}
		case EnumNode:
			if string(n.Name) == nameStr {
				return true
			}
		}
	}
	return false
}

// newResolveData creates a new resolveData with cached common types from the schema.
func newResolveData(schema *Schema, namespace *NamespaceNode) *resolveData {
	rd := &resolveData{
		schema:               schema,
		namespace:            namespace,
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

	// Build namespace-local common types map (unqualified names)
	// Populate with unresolved entries for the current namespace
	if namespace != nil {
		for ct := range namespace.CommonTypes() {
			ctCopy := ct
			rd.namespaceCommonTypes[string(ct.Name)] = &commonTypeEntry{
				resolved: false,
				node:     ctCopy,
			}
		}
	}

	return rd
}

// withNamespace returns a new resolveData with the given namespace.
func (rd *resolveData) withNamespace(namespace *NamespaceNode) *resolveData {
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

// ResolvedEntity represents an entity type with all type references fully resolved.
// All EntityTypeRef references have been converted to types.EntityType.
type ResolvedEntity struct {
	Name        types.EntityType   // Fully qualified entity type
	Annotations []Annotation       // Entity annotations
	MemberOf    []types.EntityType // Fully qualified parent entity types
	Shape       *RecordType        // Entity shape (with all type references resolved)
	Tags        IsType             // Tags type (with all type references resolved)
}

// ResolvedEnum represents an enum type with all references fully resolved.
type ResolvedEnum struct {
	Name        types.EntityType // Fully qualified enum type
	Annotations []Annotation     // Enum annotations
	Values      []types.String   // Enum values
}

// EntityUIDs returns an iterator over EntityUID values for each enum value.
// The Name field should already be fully qualified.
func (e ResolvedEnum) EntityUIDs() iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(e.Name, v)) {
				return
			}
		}
	}
}

// ResolvedAppliesTo represents the appliesTo clause with all type references fully resolved.
// All EntityTypeRef references have been converted to types.EntityType.
type ResolvedAppliesTo struct {
	PrincipalTypes []types.EntityType // Fully qualified principal entity types
	ResourceTypes  []types.EntityType // Fully qualified resource entity types
	Context        IsType             // Context type (with all type references resolved)
}

// ResolvedAction represents an action with all type references fully resolved.
// All EntityTypeRef and EntityRef references have been converted to types.EntityType and types.EntityUID.
type ResolvedAction struct {
	Name        types.String       // Action name (local, not qualified)
	Annotations []Annotation       // Action annotations
	MemberOf    []types.EntityUID  // Fully qualified parent action UIDs
	AppliesTo   *ResolvedAppliesTo // AppliesTo clause with all type references resolved
}

// ResolvedSchema represents a schema with all type references resolved and indexed for efficient lookup.
type ResolvedSchema struct {
	Entities   map[types.EntityType]ResolvedEntity // Fully qualified entity type -> ResolvedEntity
	Enums      map[types.EntityType]ResolvedEnum   // Fully qualified entity type -> ResolvedEnum
	Actions    map[types.EntityUID]ResolvedAction  // Fully qualified action UID -> ResolvedAction
	Namespaces map[types.Path][]Annotation         // Namespace path -> Annotations
}

// convertResolvedEntityToNode converts a ResolvedEntity back to an EntityNode by converting
// Resolve returns a ResolvedSchema with all type references resolved and indexed.
// Type references within namespaces are resolved relative to their namespace.
// Top-level type references are resolved as-is.
// Returns an error if any type reference cannot be resolved or if there are naming conflicts.
func (s *Schema) Resolve() (*ResolvedSchema, error) {
	resolved := &ResolvedSchema{
		Entities:   make(map[types.EntityType]ResolvedEntity),
		Enums:      make(map[types.EntityType]ResolvedEnum),
		Actions:    make(map[types.EntityUID]ResolvedAction),
		Namespaces: make(map[types.Path][]Annotation),
	}

	rd := newResolveData(s, nil)

	for _, node := range s.Nodes {
		switch n := node.(type) {
		case NamespaceNode:
			// Store namespace annotations
			if n.Name != "" {
				resolved.Namespaces[n.Name] = n.Annotations
			}

			// Create resolve data with this namespace
			nsRd := rd.withNamespace(&n)

			// Resolve all declarations in the namespace
			for _, decl := range n.Declarations {
				switch d := decl.(type) {
				case CommonTypeNode:
					// Common types are resolved but not added to the maps
					_, err := d.resolve(nsRd)
					if err != nil {
						return nil, err
					}

				case EntityNode:
					resolvedEntity, err := d.resolve(nsRd)
					if err != nil {
						return nil, err
					}
					// Check for conflicts with existing entities or enums
					if _, exists := resolved.Entities[resolvedEntity.Name]; exists {
						return nil, fmt.Errorf("entity type %q is defined multiple times", resolvedEntity.Name)
					}
					if _, exists := resolved.Enums[resolvedEntity.Name]; exists {
						return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEntity.Name)
					}
					resolved.Entities[resolvedEntity.Name] = resolvedEntity

				case EnumNode:
					resolvedEnum := d.resolve(nsRd)
					// Check for conflicts with existing enums or entities
					if _, exists := resolved.Enums[resolvedEnum.Name]; exists {
						return nil, fmt.Errorf("enum type %q is defined multiple times", resolvedEnum.Name)
					}
					if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
						return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
					}
					resolved.Enums[resolvedEnum.Name] = resolvedEnum

				case ActionNode:
					resolvedAction, err := d.resolve(nsRd)
					if err != nil {
						return nil, err
					}
					// Construct EntityUID from qualified action type
					var actionType types.EntityType
					if n.Name == "" {
						actionType = "Action"
					} else {
						actionType = types.EntityType(string(n.Name) + "::Action")
					}
					actionUID := types.NewEntityUID(actionType, resolvedAction.Name)
					// Check for duplicate actions
					if _, exists := resolved.Actions[actionUID]; exists {
						return nil, fmt.Errorf("action %q is defined multiple times", actionUID)
					}
					resolved.Actions[actionUID] = resolvedAction
				}
			}

		case EntityNode:
			resolvedEntity, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}
			// Check for conflicts
			if _, exists := resolved.Entities[resolvedEntity.Name]; exists {
				return nil, fmt.Errorf("entity type %q is defined multiple times", resolvedEntity.Name)
			}
			if _, exists := resolved.Enums[resolvedEntity.Name]; exists {
				return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEntity.Name)
			}
			resolved.Entities[resolvedEntity.Name] = resolvedEntity

		case EnumNode:
			resolvedEnum := n.resolve(rd)
			// Check for conflicts
			if _, exists := resolved.Enums[resolvedEnum.Name]; exists {
				return nil, fmt.Errorf("enum type %q is defined multiple times", resolvedEnum.Name)
			}
			if _, exists := resolved.Entities[resolvedEnum.Name]; exists {
				return nil, fmt.Errorf("type %q is defined as both an entity and an enum", resolvedEnum.Name)
			}
			resolved.Enums[resolvedEnum.Name] = resolvedEnum

		case ActionNode:
			resolvedAction, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}
			// Top-level actions use "Action" as the type
			actionUID := types.NewEntityUID("Action", resolvedAction.Name)
			// Check for duplicate actions
			if _, exists := resolved.Actions[actionUID]; exists {
				return nil, fmt.Errorf("action %q is defined multiple times", actionUID)
			}
			resolved.Actions[actionUID] = resolvedAction

		case CommonTypeNode:
			// Common types are resolved but not added to the maps
			// They are used during resolution via the cache
			_, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}
		}
	}

	return resolved, nil
}
