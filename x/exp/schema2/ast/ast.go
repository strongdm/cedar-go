// Package ast provides types and functions for constructing Cedar schema ASTs programmatically.
package ast

import (
	"iter"

	"github.com/cedar-policy/cedar-go/types"
)

// IsNode is the interface implemented by all schema nodes.
type IsNode interface {
	isNode()
}

// IsDeclaration is the interface implemented by nodes that can be
// declarations within a namespace (CommonType, Entity, Enum, Action).
type IsDeclaration interface {
	IsNode
	isDeclaration()
}

// IsType is the interface implemented by all type expressions.
type IsType interface {
	isType()
	// resolve returns a new type with all references resolved using the provided resolve data.
	// Returns an error if a type reference cannot be resolved.
	resolve(rd *resolveData) (IsType, error)
}

// resolveData contains cached information for efficient type resolution.
type resolveData struct {
	schema               *Schema
	namespace            *NamespaceNode
	schemaCommonTypes    map[string]IsType // Fully qualified name -> resolved Type
	namespaceCommonTypes map[string]IsType // Unqualified name -> resolved Type
}

// Annotation represents a Cedar annotation (@key("value")).
type Annotation struct {
	Key   types.Ident
	Value types.String
}

// Schema represents a Cedar schema containing a list of nodes.
type Schema struct {
	Nodes []IsNode
}

// NewSchema creates a new Schema from the given nodes.
func NewSchema(nodes ...IsNode) *Schema {
	return &Schema{Nodes: nodes}
}

// newResolveData creates a new resolveData with cached common types from the schema.
func newResolveData(schema *Schema, namespace *NamespaceNode) *resolveData {
	rd := &resolveData{
		schema:               schema,
		namespace:            namespace,
		schemaCommonTypes:    make(map[string]IsType),
		namespaceCommonTypes: make(map[string]IsType),
	}

	// Build schema-wide common types map (fully qualified names)
	// We resolve the types lazily in TypeRef.Resolve to avoid circular dependencies

	// Build namespace-local common types map (unqualified names)
	// We resolve the types lazily in TypeRef.Resolve to avoid circular dependencies

	return rd
}

// withNamespace returns a new resolveData with the given namespace.
func (rd *resolveData) withNamespace(namespace *NamespaceNode) *resolveData {
	if namespace == rd.namespace {
		return rd
	}

	return &resolveData{
		schema:               rd.schema,
		namespace:            namespace,
		schemaCommonTypes:    rd.schemaCommonTypes,    // Reuse schema-wide cache
		namespaceCommonTypes: rd.namespaceCommonTypes, // Reuse namespace cache (will be lazily populated)
	}
}

// ResolvedSchema represents a schema with all type references resolved and indexed for efficient lookup.
type ResolvedSchema struct {
	Entities map[types.EntityType]EntityNode // Fully qualified entity type -> EntityNode
	Enums    map[types.EntityType]EnumNode   // Fully qualified entity type -> EnumNode
	Actions  map[types.EntityUID]ActionNode  // Fully qualified action UID -> ActionNode
}

// Resolve returns a ResolvedSchema with all type references resolved and indexed.
// Type references within namespaces are resolved relative to their namespace.
// Top-level type references are resolved as-is.
// Returns an error if any type reference cannot be resolved.
func (s *Schema) Resolve() (*ResolvedSchema, error) {
	resolved := &ResolvedSchema{
		Entities: make(map[types.EntityType]EntityNode),
		Enums:    make(map[types.EntityType]EnumNode),
		Actions:  make(map[types.EntityUID]ActionNode),
	}

	rd := newResolveData(s, nil)

	for _, node := range s.Nodes {
		switch n := node.(type) {
		case NamespaceNode:
			// Resolve all declarations in the namespace
			resolvedDecls, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}

			// Iterate over resolved declarations
			for _, decl := range resolvedDecls {
				switch d := decl.(type) {
				case EntityNode:
					// Name is already fully qualified by Resolve
					resolved.Entities[d.Name] = d
				case EnumNode:
					// Name is already fully qualified by Resolve
					resolved.Enums[d.Name] = d
				case ActionNode:
					// Construct EntityUID from qualified action type
					// Extract namespace from a fully qualified entity in this namespace's declarations
					// Or build it from the namespace name
					var actionType types.EntityType
					if n.Name == "" {
						actionType = "Action"
					} else {
						actionType = types.EntityType(string(n.Name) + "::Action")
					}
					actionUID := types.NewEntityUID(actionType, d.Name)
					resolved.Actions[actionUID] = d
				}
			}

		case EntityNode:
			resolvedEntity, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}
			// Name is already fully qualified by resolve (or stays unqualified for top-level)
			resolved.Entities[resolvedEntity.Name] = resolvedEntity

		case EnumNode:
			resolvedEnum := n.resolve(rd)
			// Name is already fully qualified by resolve (or stays unqualified for top-level)
			resolved.Enums[resolvedEnum.Name] = resolvedEnum

		case ActionNode:
			resolvedAction, err := n.resolve(rd)
			if err != nil {
				return nil, err
			}
			// Top-level actions use "Action" as the type
			actionUID := types.NewEntityUID("Action", resolvedAction.Name)
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

// Namespaces returns an iterator over all NamespaceNode declarations in the schema.
func (s *Schema) Namespaces() iter.Seq[NamespaceNode] {
	return s.namespaces()
}

// namespaces returns an iterator over all NamespaceNode declarations in the schema.
func (s *Schema) namespaces() iter.Seq[NamespaceNode] {
	return func(yield func(NamespaceNode) bool) {
		for _, node := range s.Nodes {
			if ns, ok := node.(NamespaceNode); ok {
				if !yield(ns) {
					return
				}
			}
		}
	}
}

// CommonTypes returns an iterator over all CommonTypeNode declarations in the schema.
func (s *Schema) CommonTypes() iter.Seq2[*NamespaceNode, CommonTypeNode] {
	return s.commonTypes()
}

// commonTypes returns an iterator over all CommonTypeNode declarations in the schema.
func (s *Schema) commonTypes() iter.Seq2[*NamespaceNode, CommonTypeNode] {
	return func(yield func(*NamespaceNode, CommonTypeNode) bool) {
		for _, node := range s.Nodes {
			if ct, ok := node.(CommonTypeNode); ok {
				if !yield(nil, ct) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for ct := range ns.commonTypes() {
					if !yield(&ns, ct) {
						return
					}
				}
			}
		}
	}
}

// Entities returns an iterator over all EntityNode declarations in the schema.
func (s *Schema) Entities() iter.Seq2[*NamespaceNode, EntityNode] {
	return s.entities()
}

// entities returns an iterator over all EntityNode declarations in the schema.
func (s *Schema) entities() iter.Seq2[*NamespaceNode, EntityNode] {
	return func(yield func(*NamespaceNode, EntityNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(EntityNode); ok {
				if !yield(nil, e) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for e := range ns.entities() {
					if !yield(&ns, e) {
						return
					}
				}
			}
		}
	}
}

// Enums returns an iterator over all EnumNode declarations in the schema.
func (s *Schema) Enums() iter.Seq2[*NamespaceNode, EnumNode] {
	return s.enums()
}

// enums returns an iterator over all EnumNode declarations in the schema.
func (s *Schema) enums() iter.Seq2[*NamespaceNode, EnumNode] {
	return func(yield func(*NamespaceNode, EnumNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(EnumNode); ok {
				if !yield(nil, e) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for e := range ns.enums() {
					if !yield(&ns, e) {
						return
					}
				}
			}
		}
	}
}

// Actions returns an iterator over all ActionNode declarations in the schema.
func (s *Schema) Actions() iter.Seq2[*NamespaceNode, ActionNode] {
	return s.actions()
}

// actions returns an iterator over all ActionNode declarations in the schema.
func (s *Schema) actions() iter.Seq2[*NamespaceNode, ActionNode] {
	return func(yield func(*NamespaceNode, ActionNode) bool) {
		for _, node := range s.Nodes {
			if a, ok := node.(ActionNode); ok {
				if !yield(nil, a) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for a := range ns.actions() {
					if !yield(&ns, a) {
						return
					}
				}
			}
		}
	}
}
