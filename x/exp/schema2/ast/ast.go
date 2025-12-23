// Package ast provides types and functions for constructing Cedar schema ASTs programmatically.
package ast

import "github.com/cedar-policy/cedar-go/types"

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

// ResolvedSchema represents a Cedar schema where all type references
// have been resolved to their fully qualified names.
type ResolvedSchema struct {
	Nodes []IsNode
}
