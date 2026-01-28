package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// CommonTypeNode represents a Cedar common type declaration (type alias).
type CommonTypeNode struct {
	Annotations Annotations
	Type        IsType
}

// EntityNode represents a Cedar entity type declaration.
type EntityNode struct {
	Annotations Annotations
	MemberOfVal []EntityTypeRef
	ShapeVal    *RecordType
	TagsVal     IsType
}

// EnumNode represents a Cedar enum entity type declaration.
type EnumNode struct {
	Annotations Annotations
	Values      []types.String
}

// ActionNode represents a Cedar action declaration.
type ActionNode struct {
	Annotations  Annotations
	MemberOfVal  []EntityRef
	AppliesToVal *AppliesTo
}

// AppliesTo represents the principal, resource, and context types for an action.
type AppliesTo struct {
	PrincipalTypes []EntityTypeRef
	ResourceTypes  []EntityTypeRef
	Context        IsType
}

// EntityRef represents a reference to a specific entity (type + id).
type EntityRef struct {
	Type EntityTypeRef
	ID   types.String
}

// EntityRefFromID creates an EntityRef with just an ID (type is inferred as Action).
func EntityRefFromID(id types.String) EntityRef {
	return EntityRef{
		ID: id,
	}
}

// NewEntityRef creates an EntityRef with an explicit type and ID.
func NewEntityRef(typ types.EntityType, id types.String) EntityRef {
	return EntityRef{
		Type: EntityTypeRef{Name: typ},
		ID:   id,
	}
}
