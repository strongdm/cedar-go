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
func (s *Schema) Namespaces() iter.Seq[NamespaceNode] {
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
	return func(yield func(*NamespaceNode, CommonTypeNode) bool) {
		for _, node := range s.Nodes {
			if ct, ok := node.(CommonTypeNode); ok {
				if !yield(nil, ct) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for ct := range ns.CommonTypes() {
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
	return func(yield func(*NamespaceNode, EntityNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(EntityNode); ok {
				if !yield(nil, e) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for e := range ns.Entities() {
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
	return func(yield func(*NamespaceNode, EnumNode) bool) {
		for _, node := range s.Nodes {
			if e, ok := node.(EnumNode); ok {
				if !yield(nil, e) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for e := range ns.Enums() {
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
	return func(yield func(*NamespaceNode, ActionNode) bool) {
		for _, node := range s.Nodes {
			if a, ok := node.(ActionNode); ok {
				if !yield(nil, a) {
					return
				}
			} else if ns, ok := node.(NamespaceNode); ok {
				for a := range ns.Actions() {
					if !yield(&ns, a) {
						return
					}
				}
			}
		}
	}
}
