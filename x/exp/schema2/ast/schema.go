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

// NamespaceNode represents a Cedar namespace declaration.
type NamespaceNode struct {
	Name         types.Path
	Annotations  []Annotation
	Declarations []IsDeclaration
}

func (NamespaceNode) isNode() {}

// Namespace creates a new NamespaceNode with the given path and declarations.
func Namespace(path types.Path, decls ...IsDeclaration) NamespaceNode {
	return NamespaceNode{
		Name:         path,
		Declarations: decls,
	}
}

// Annotate adds an annotation to the namespace and returns the node for chaining.
func (n NamespaceNode) Annotate(key types.Ident, value types.String) NamespaceNode {
	n.Annotations = append(n.Annotations, Annotation{Key: key, Value: value})
	return n
}

// CommonTypes returns an iterator over all CommonTypeNode declarations in the namespace.
func (n NamespaceNode) CommonTypes() iter.Seq[CommonTypeNode] {
	return func(yield func(CommonTypeNode) bool) {
		for _, decl := range n.Declarations {
			if ct, ok := decl.(CommonTypeNode); ok {
				if !yield(ct) {
					return
				}
			}
		}
	}
}

// Entities returns an iterator over all EntityNode declarations in the namespace.
func (n NamespaceNode) Entities() iter.Seq[EntityNode] {
	return func(yield func(EntityNode) bool) {
		for _, decl := range n.Declarations {
			if e, ok := decl.(EntityNode); ok {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// Enums returns an iterator over all EnumNode declarations in the namespace.
func (n NamespaceNode) Enums() iter.Seq[EnumNode] {
	return func(yield func(EnumNode) bool) {
		for _, decl := range n.Declarations {
			if e, ok := decl.(EnumNode); ok {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// Actions returns an iterator over all ActionNode declarations in the namespace.
func (n NamespaceNode) Actions() iter.Seq[ActionNode] {
	return func(yield func(ActionNode) bool) {
		for _, decl := range n.Declarations {
			if a, ok := decl.(ActionNode); ok {
				if !yield(a) {
					return
				}
			}
		}
	}
}
