// Package schema provides types and functions for working with Cedar schemas.
//
// Schemas define the structure of entities and actions in a Cedar policy environment.
// They can be authored programmatically using builder patterns similar to the ast package,
// or parsed from JSON or Cedar schema formats.
//
// Example usage:
//
//	schema := schema.New().
//	    Namespace("MyApp",
//	        schema.EntityType("User").
//	            MemberOf("Group").
//	            Attributes(
//	                schema.Attr("name", schema.String(), true),
//	                schema.Attr("email", schema.String(), false),
//	            ),
//	        schema.EntityType("Group"),
//	        schema.ActionType("view").
//	            PrincipalTypes("User").
//	            ResourceTypes("Document"),
//	    )
package schema

import "github.com/cedar-policy/cedar-go/types"

// Type represents a Cedar type in a schema.
// Types can be primitive (Boolean, Long, String), complex (Set, Record),
// entity references, extension types, or references to common types.
type Type struct {
	v isType
}

// isType is the interface implemented by all type variants.
type isType interface {
	isType()
	equal(isType) bool
}

// AsIsType returns the underlying type variant.
func (t Type) AsIsType() isType {
	return t.v
}

// newType creates a new Type wrapping the given variant.
func newType(v isType) Type {
	return Type{v: v}
}

// TypeBoolean represents the Cedar Boolean type.
type TypeBoolean struct{}

func (TypeBoolean) isType()               { _ = 0 }
func (TypeBoolean) equal(o isType) bool   { _, ok := o.(TypeBoolean); return ok }

// TypeLong represents the Cedar Long (integer) type.
type TypeLong struct{}

func (TypeLong) isType()               { _ = 0 }
func (TypeLong) equal(o isType) bool   { _, ok := o.(TypeLong); return ok }

// TypeString represents the Cedar String type.
type TypeString struct{}

func (TypeString) isType()               { _ = 0 }
func (TypeString) equal(o isType) bool   { _, ok := o.(TypeString); return ok }

// TypeSet represents a Cedar Set type containing elements of a specific type.
type TypeSet struct {
	Element Type
}

func (TypeSet) isType() { _ = 0 }
func (t TypeSet) equal(o isType) bool {
	if ot, ok := o.(TypeSet); ok {
		return typeEqual(t.Element, ot.Element)
	}
	return false
}

// TypeRecord represents a Cedar Record type with named attributes.
type TypeRecord struct {
	Attributes []Attribute
}

func (TypeRecord) isType() { _ = 0 }
func (t TypeRecord) equal(o isType) bool {
	ot, ok := o.(TypeRecord)
	if !ok || len(t.Attributes) != len(ot.Attributes) {
		return false
	}
	for i, a := range t.Attributes {
		if !a.equal(ot.Attributes[i]) {
			return false
		}
	}
	return true
}

// TypeEntity represents a reference to a Cedar entity type.
type TypeEntity struct {
	Name types.Path
}

func (TypeEntity) isType() { _ = 0 }
func (t TypeEntity) equal(o isType) bool {
	if ot, ok := o.(TypeEntity); ok {
		return t.Name == ot.Name
	}
	return false
}

// TypeExtension represents a Cedar extension type (e.g., ipaddr, decimal).
type TypeExtension struct {
	Name types.Path
}

func (TypeExtension) isType() { _ = 0 }
func (t TypeExtension) equal(o isType) bool {
	if ot, ok := o.(TypeExtension); ok {
		return t.Name == ot.Name
	}
	return false
}

// TypeRef represents a reference to a common type defined in the schema.
type TypeRef struct {
	Name types.Path
}

func (TypeRef) isType() { _ = 0 }
func (t TypeRef) equal(o isType) bool {
	if ot, ok := o.(TypeRef); ok {
		return t.Name == ot.Name
	}
	return false
}

// Type constructors

// Boolean returns the Cedar Boolean type.
func Boolean() Type {
	return newType(TypeBoolean{})
}

// Long returns the Cedar Long type.
func Long() Type {
	return newType(TypeLong{})
}

// String returns the Cedar String type.
func String() Type {
	return newType(TypeString{})
}

// SetOf returns a Cedar Set type containing elements of the given type.
func SetOf(element Type) Type {
	return newType(TypeSet{Element: element})
}

// Record returns a Cedar Record type with the given attributes.
func Record(attrs ...Attribute) Type {
	return newType(TypeRecord{Attributes: attrs})
}

// Entity returns a reference to the named entity type.
// The name can be a simple identifier or a qualified path like "Namespace::Type".
func Entity(name types.Path) Type {
	return newType(TypeEntity{Name: name})
}

// Extension returns a Cedar extension type.
func Extension(name types.Path) Type {
	return newType(TypeExtension{Name: name})
}

// TypeRef returns a reference to a common type.
func Ref(name types.Path) Type {
	return newType(TypeRef{Name: name})
}

// Common extension type shortcuts

// IPAddr returns the ipaddr extension type.
func IPAddr() Type {
	return Extension("ipaddr")
}

// Decimal returns the decimal extension type.
func Decimal() Type {
	return Extension("decimal")
}

// Datetime returns the datetime extension type.
func Datetime() Type {
	return Extension("datetime")
}

// Duration returns the duration extension type.
func Duration() Type {
	return Extension("duration")
}

// Attribute represents an attribute in a Record type.
type Attribute struct {
	Name     types.Ident
	Type     Type
	Required bool
}

// equal returns true if the attributes are equal.
func (a Attribute) equal(b Attribute) bool {
	return a.Name == b.Name && a.Required == b.Required && typeEqual(a.Type, b.Type)
}

// Attr creates a new attribute with the given name, type, and required flag.
func Attr(name types.Ident, t Type, required bool) Attribute {
	return Attribute{
		Name:     name,
		Type:     t,
		Required: required,
	}
}

// OptionalAttr creates an optional attribute with the given name and type.
func OptionalAttr(name types.Ident, t Type) Attribute {
	return Attr(name, t, false)
}

// RequiredAttr creates a required attribute with the given name and type.
func RequiredAttr(name types.Ident, t Type) Attribute {
	return Attr(name, t, true)
}

// typeEqual returns true if two types are equal.
func typeEqual(a, b Type) bool {
	if a.v == nil && b.v == nil {
		return true
	}
	if a.v == nil || b.v == nil {
		return false
	}
	return a.v.equal(b.v)
}
