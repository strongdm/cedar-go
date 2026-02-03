// Package ast provides types and functions for constructing Cedar schema ASTs programmatically.
package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Annotations map[types.Ident]types.String
type Entities map[types.EntityType]Entity
type Enums map[types.EntityType]Enum
type Actions map[types.String]Action
type CommonTypes map[types.Ident]CommonType
type Namespaces map[types.Path]Namespace

type Schema struct {
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
	Namespaces  Namespaces
}

type Namespace struct {
	Annotations Annotations
	Entities    Entities
	Enums       Enums
	Actions     Actions
	CommonTypes CommonTypes
}

type CommonType struct {
	Annotations Annotations
	Type        IsType
}

type Entity struct {
	Annotations Annotations
	MemberOf    []EntityTypeRef
	Shape       *RecordType
	Tags        IsType
}

type Enum struct {
	Annotations Annotations
	Values      []types.String
}

// Action defines what principals can do to resources.
// If AppliesTo is nil, the action never applies.
type Action struct {
	Annotations Annotations
	MemberOf    []ParentRef
	AppliesTo   *AppliesTo
}

type AppliesTo struct {
	Principals []EntityTypeRef
	Resources  []EntityTypeRef
	Context    IsType
}

type ParentRef struct {
	Type EntityTypeRef
	ID   types.String
}

// ParentRefFromID creates an ParentRef with only an ID.
// Type is inferred as Action.
func ParentRefFromID(id types.String) ParentRef {
	return ParentRef{
		ID: id,
	}
}

// NewParentRef creates an ParentRef with type and ID.
func NewParentRef(typ types.EntityType, id types.String) ParentRef {
	return ParentRef{
		Type: EntityTypeRef(typ),
		ID:   id,
	}
}
