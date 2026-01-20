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

// Namespaces returns an iterator over all NamespaceNode declarations in the schema.
// This allows you to iterate through all namespace declarations defined at the schema level.
func (s *Schema) Namespaces() iter.Seq[*NamespaceNode] {
	return func(yield func(*NamespaceNode) bool) {
		for _, node := range s.Nodes {
			if ns, ok := node.(*NamespaceNode); ok {
				if !yield(ns) {
					return
				}
			}
		}
	}
}

// CommonTypes returns an iterator over all CommonTypeNode declarations in the schema.
// This includes both top-level declarations and declarations within namespaces.
// Items are yielded in the order they appear in the schema.
// The iterator yields both the namespace path and the node.
func (s *Schema) CommonTypes() iter.Seq2[types.Path, *CommonTypeNode] {
	return func(yield func(types.Path, *CommonTypeNode) bool) {
		for _, node := range s.Nodes {
			if ct, ok := node.(*CommonTypeNode); ok {
				if !yield("", ct) {
					return
				}
			} else if ns, ok := node.(*NamespaceNode); ok {
				for ct := range ns.CommonTypes() {
					if !yield(ns.Name, ct) {
						return
					}
				}
			}
		}
	}
}

// Entities returns an iterator over all EntityNode declarations in the schema.
// This includes both top-level declarations and declarations within namespaces.
// Items are yielded in the order they appear in the schema.
// The iterator yields both the namespace path and the node.
func (s *Schema) Entities() iter.Seq2[types.Path, *EntityNode] {
	return func(yield func(types.Path, *EntityNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(*EntityNode); ok {
				if !yield("", e) {
					return
				}
			} else if ns, ok := node.(*NamespaceNode); ok {
				for e := range ns.Entities() {
					if !yield(ns.Name, e) {
						return
					}
				}
			}
		}
	}
}

// Enums returns an iterator over all EnumNode declarations in the schema.
// This includes both top-level declarations and declarations within namespaces.
// Items are yielded in the order they appear in the schema.
// The iterator yields both the namespace path and the node.
func (s *Schema) Enums() iter.Seq2[types.Path, *EnumNode] {
	return func(yield func(types.Path, *EnumNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(*EnumNode); ok {
				if !yield("", e) {
					return
				}
			} else if ns, ok := node.(*NamespaceNode); ok {
				for e := range ns.Enums() {
					if !yield(ns.Name, e) {
						return
					}
				}
			}
		}
	}
}

// Actions returns an iterator over all ActionNode declarations in the schema.
// This includes both top-level declarations and declarations within namespaces.
// Items are yielded in the order they appear in the schema.
// The iterator yields both the namespace path and the node.
func (s *Schema) Actions() iter.Seq2[types.Path, *ActionNode] {
	return func(yield func(types.Path, *ActionNode) bool) {
		for _, node := range s.Nodes {
			if a, ok := node.(*ActionNode); ok {
				if !yield("", a) {
					return
				}
			} else if ns, ok := node.(*NamespaceNode); ok {
				for a := range ns.Actions() {
					if !yield(ns.Name, a) {
						return
					}
				}
			}
		}
	}
}

// ResolvedSchema represents a Cedar schema where all type references
// have been resolved to their fully qualified names.
type ResolvedSchema struct {
	Nodes []IsNode
}
