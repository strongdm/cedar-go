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
	MemberOf    []EntityRef
	AppliesTo   *AppliesTo
}

type AppliesTo struct {
	Principals []EntityTypeRef
	Resources  []EntityTypeRef
	Context    IsType
}

type EntityRef struct {
	Type EntityTypeRef
	ID   types.String
}

// EntityRefFromID creates an EntityRef with only an ID.
// Type is inferred as Action.
func EntityRefFromID(id types.String) EntityRef {
	return EntityRef{
		ID: id,
	}
}

// NewEntityRef creates an EntityRef with type and ID.
func NewEntityRef(typ types.EntityType, id types.String) EntityRef {
	return EntityRef{
		Type: EntityTypeRef(typ),
		ID:   id,
	}
}
