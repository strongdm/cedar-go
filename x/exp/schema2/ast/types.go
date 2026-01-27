package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

// IsType is the interface implemented by all type expressions.
type IsType interface {
	isType()
}

// Primitive types

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() {}

// String returns a StringType.
func String() StringType { return StringType{} }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() {}

// Long returns a LongType.
func Long() LongType { return LongType{} }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() {}

// Bool returns a BoolType.
func Bool() BoolType { return BoolType{} }

// Extension types

// ExtensionType represents a Cedar extension type (ipaddr, decimal, datetime, duration).
type ExtensionType struct {
	Name types.Ident
}

func (ExtensionType) isType() {}

// IPAddr returns an ExtensionType for ipaddr.
func IPAddr() ExtensionType { return ExtensionType{Name: "ipaddr"} }

// Decimal returns an ExtensionType for decimal.
func Decimal() ExtensionType { return ExtensionType{Name: "decimal"} }

// Datetime returns an ExtensionType for datetime.
func Datetime() ExtensionType { return ExtensionType{Name: "datetime"} }

// Duration returns an ExtensionType for duration.
func Duration() ExtensionType { return ExtensionType{Name: "duration"} }

// Collection types

// SetType represents a Cedar Set type with an element type.
type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

// Set returns a SetType with the given element type.
func Set(element IsType) SetType {
	return SetType{Element: element}
}

// Record types

// Pair represents a key-value pair in a record type.
type Pair struct {
	Key         types.String
	Type        IsType
	Optional    bool
	Annotations []Annotation
}

// Attribute creates a required attribute pair.
func Attribute(key types.String, t IsType) Pair {
	return Pair{Key: key, Type: t, Optional: false}
}

// Optional creates an optional attribute pair.
func Optional(key types.String, t IsType) Pair {
	return Pair{Key: key, Type: t, Optional: true}
}

func (p Pair) Annotate(ann ...Annotation) Pair {
	p.Annotations = append(p.Annotations, ann...)
	return p
}

// RecordType represents a Cedar Record type with attributes.
type RecordType struct {
	Pairs []Pair
}

func (RecordType) isType() { _ = 0 }

// Record returns a RecordType with the given pairs.
func Record(pairs ...Pair) RecordType {
	return RecordType{Pairs: pairs}
}

// Reference types

// EntityTypeRef represents a reference to an entity type.
type EntityTypeRef struct {
	Name types.EntityType
}

func (EntityTypeRef) isType() { _ = 0 }

// EntityType creates an EntityTypeRef from an entity type name.
func EntityType(name types.EntityType) EntityTypeRef {
	return EntityTypeRef{Name: name}
}

// Ref is an alias for EntityType for more concise syntax.
func Ref(name types.EntityType) EntityTypeRef {
	return EntityType(name)
}

// TypeRef represents a reference to a common type by name.
type TypeRef struct {
	Name types.Path
}

func (TypeRef) isType() { _ = 0 }

// Type creates a TypeRef from a path name.
func Type(name types.Path) TypeRef {
	return TypeRef{Name: name}
}
