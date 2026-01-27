package ast

import (
	"iter"

	"github.com/cedar-policy/cedar-go/types"
)

// IsDeclaration is the interface implemented by nodes that can be
// declarations within a namespace (CommonType, Entity, Enum, Action).
type IsDeclaration interface {
	IsNode
	isDeclaration()
}

// CommonTypeNode represents a Cedar common type declaration (type alias).
type CommonTypeNode struct {
	Name        types.Ident
	Annotations []Annotation
	Type        IsType
}

func (CommonTypeNode) isNode()        {}
func (CommonTypeNode) isDeclaration() {}

// CommonType creates a new CommonTypeNode with the given name and type.
func CommonType(name types.Ident, t IsType) CommonTypeNode {
	return CommonTypeNode{
		Name: name,
		Type: t,
	}
}

// Annotate adds an annotation to the common type and returns the node for chaining.
func (c CommonTypeNode) Annotate(key types.Ident, value types.String) CommonTypeNode {
	c.Annotations = append(c.Annotations, Annotation{Key: key, Value: value})
	return c
}

// EntityNode represents a Cedar entity type declaration.
type EntityNode struct {
	Name        types.EntityType
	Annotations []Annotation
	MemberOfVal []EntityTypeRef
	ShapeVal    *RecordType
	TagsVal     IsType
}

func (EntityNode) isNode()        {}
func (EntityNode) isDeclaration() {}

// Entity creates a new EntityNode with the given name.
func Entity(name types.EntityType) EntityNode {
	return EntityNode{Name: name}
}

// MemberOf sets the entity types this entity can be a member of.
func (e EntityNode) MemberOf(parents ...EntityTypeRef) EntityNode {
	e.MemberOfVal = parents
	return e
}

// Shape sets the shape (attributes) of the entity.
func (e EntityNode) Shape(pairs ...Pair) EntityNode {
	r := Record(pairs...)
	e.ShapeVal = &r
	return e
}

// Tags sets the tags type for the entity.
func (e EntityNode) Tags(t IsType) EntityNode {
	e.TagsVal = t
	return e
}

// Annotate adds an annotation to the entity and returns the node for chaining.
func (e EntityNode) Annotate(key types.Ident, value types.String) EntityNode {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// EnumNode represents a Cedar enum entity type declaration.
type EnumNode struct {
	Name        types.EntityType
	Annotations []Annotation
	Values      []types.String
}

func (EnumNode) isNode()        {}
func (EnumNode) isDeclaration() {}

// Enum creates a new EnumNode with the given name and values.
func Enum(name types.EntityType, values ...types.String) EnumNode {
	return EnumNode{
		Name:   name,
		Values: values,
	}
}

// Annotate adds an annotation to the enum and returns the node for chaining.
func (e EnumNode) Annotate(key types.Ident, value types.String) EnumNode {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// EntityUIDs returns an iterator over EntityUID values for each enum value.
// The Name field should already be fully qualified after calling Resolve().
func (e EnumNode) EntityUIDs() iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(e.Name, v)) {
				return
			}
		}
	}
}

// ActionNode represents a Cedar action declaration.
type ActionNode struct {
	Name         types.String
	Annotations  []Annotation
	MemberOfVal  []EntityRef
	AppliesToVal *AppliesTo
}

func (ActionNode) isNode()        {}
func (ActionNode) isDeclaration() {}

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

// UID creates an EntityRef with just an ID (type is inferred as Action).
func UID(id types.String) EntityRef {
	return EntityRef{
		ID: id,
	}
}

// EntityUID creates an EntityRef with an explicit type and ID.
func EntityUID(typ types.EntityType, id types.String) EntityRef {
	return EntityRef{
		Type: EntityTypeRef{Name: typ},
		ID:   id,
	}
}

// Action creates a new ActionNode with the given name.
func Action(name types.String) ActionNode {
	return ActionNode{Name: name}
}

// MemberOf sets the action groups this action is a member of.
func (a ActionNode) MemberOf(refs ...EntityRef) ActionNode {
	a.MemberOfVal = refs
	return a
}

// Principal sets the principal types for the action.
func (a ActionNode) Principal(principals ...EntityTypeRef) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.PrincipalTypes = principals
	return a
}

// Resource sets the resource types for the action.
func (a ActionNode) Resource(resources ...EntityTypeRef) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.ResourceTypes = resources
	return a
}

// Context sets the context type for the action.
func (a ActionNode) Context(t IsType) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.Context = t
	return a
}

// Annotate adds an annotation to the action and returns the node for chaining.
func (a ActionNode) Annotate(key types.Ident, value types.String) ActionNode {
	a.Annotations = append(a.Annotations, Annotation{Key: key, Value: value})
	return a
}
