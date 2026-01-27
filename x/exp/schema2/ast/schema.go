// Package ast provides types and functions for constructing Cedar schema ASTs programmatically.
package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Annotations map[types.Ident]types.String
type Entities map[types.EntityType]EntityNode
type Enums map[types.EntityType]EnumNode
type Actions map[types.String]ActionNode
type CommonTypes map[types.Ident]CommonTypeNode
type Namespaces map[types.Path]NamespaceNode

// Schema represents a Cedar schema containing a list of nodes.
type Schema struct {
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
	Namespaces  Namespaces
}

// NamespaceNode represents a Cedar namespace declaration.
type NamespaceNode struct {
	Annotations Annotations
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
}
