package schema

import (
	"github.com/cedar-policy/cedar-go/types"
)

// Type represents a Cedar schema type. It is implemented by the various
// concrete type definitions like [TypeString], [TypeLong], [TypeBoolean],
// [TypeSet], [TypeRecord], [TypeName], and [TypeExtension].
type Type interface {
	isSchemaType()
	// MarshalCedar returns the Cedar format representation of the type.
	MarshalCedar() []byte
}

// Primitive type markers - no-op statements for code coverage instrumentation
func (TypeString) isSchemaType()  { _ = 0 }
func (TypeLong) isSchemaType()    { _ = 0 }
func (TypeBoolean) isSchemaType() { _ = 0 }

// Composite type markers - no-op statements for code coverage instrumentation
func (TypeSet) isSchemaType()    { _ = 0 }
func (TypeRecord) isSchemaType() { _ = 0 }

// Reference type markers - no-op statements for code coverage instrumentation
func (TypeName) isSchemaType()      { _ = 0 }
func (TypeExtension) isSchemaType() { _ = 0 }

// TypeString represents the Cedar String type.
type TypeString struct{}

// String returns a TypeString.
func String() TypeString {
	return TypeString{}
}

func (TypeString) MarshalCedar() []byte {
	return []byte("String")
}

// TypeLong represents the Cedar Long type.
type TypeLong struct{}

// Long returns a TypeLong.
func Long() TypeLong {
	return TypeLong{}
}

func (TypeLong) MarshalCedar() []byte {
	return []byte("Long")
}

// TypeBoolean represents the Cedar Boolean type.
type TypeBoolean struct{}

// Boolean returns a TypeBoolean.
func Boolean() TypeBoolean {
	return TypeBoolean{}
}

func (TypeBoolean) MarshalCedar() []byte {
	return []byte("Bool")
}

// TypeSet represents a Cedar Set type with an element type.
type TypeSet struct {
	Element Type
}

// Set returns a TypeSet with the given element type.
func Set(element Type) TypeSet {
	return TypeSet{Element: element}
}

func (t TypeSet) MarshalCedar() []byte {
	return append(append([]byte("Set<"), t.Element.MarshalCedar()...), '>')
}

// TypeRecord represents a Cedar Record type with named attributes.
type TypeRecord struct {
	Attributes []Attribute
}

// Record returns an empty TypeRecord. Use AddAttribute to add attributes.
func Record() TypeRecord {
	return TypeRecord{}
}

// AddAttribute returns a new TypeRecord with the attribute added.
func (r TypeRecord) AddAttribute(attr Attribute) TypeRecord {
	r.Attributes = append(r.Attributes, attr)
	return r
}

// AddRequired adds a required attribute to the record.
func (r TypeRecord) AddRequired(name types.Ident, t Type) TypeRecord {
	return r.AddAttribute(Attribute{Name: name, Type: t, Required: true})
}

// AddOptional adds an optional attribute to the record.
func (r TypeRecord) AddOptional(name types.Ident, t Type) TypeRecord {
	return r.AddAttribute(Attribute{Name: name, Type: t, Required: false})
}

func (t TypeRecord) MarshalCedar() []byte {
	buf := []byte("{")
	for i, attr := range t.Attributes {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = append(buf, '\n')
		buf = append(buf, "  "...)
		buf = append(buf, attr.MarshalCedar()...)
	}
	if len(t.Attributes) > 0 {
		buf = append(buf, ",\n"...)
	}
	buf = append(buf, '}')
	return buf
}

// TypeName represents a reference to an entity type or common type by name.
// This corresponds to "EntityOrCommon" in the JSON schema format.
type TypeName struct {
	Name types.Path
}

// EntityOrCommonType returns a TypeName referencing an entity or common type.
func EntityOrCommonType(name types.Path) TypeName {
	return TypeName{Name: name}
}

func (t TypeName) MarshalCedar() []byte {
	return []byte(t.Name)
}

// TypeExtension represents a Cedar extension type like ipaddr, decimal, datetime, or duration.
type TypeExtension struct {
	Name types.Path
}

// ExtensionType returns a TypeExtension for the given extension type name.
func ExtensionType(name types.Path) TypeExtension {
	return TypeExtension{Name: name}
}

func (t TypeExtension) MarshalCedar() []byte {
	return []byte(t.Name)
}

// Attribute represents an attribute within a record type.
type Attribute struct {
	Name        types.Ident
	Type        Type
	Required    bool
	Annotations Annotations
}

func (a Attribute) MarshalCedar() []byte {
	buf := []byte{}
	for _, ann := range a.Annotations {
		buf = append(buf, ann.MarshalCedar()...)
		buf = append(buf, '\n')
	}
	buf = append(buf, []byte(a.Name)...)
	if !a.Required {
		buf = append(buf, '?')
	}
	buf = append(buf, ": "...)
	buf = append(buf, a.Type.MarshalCedar()...)
	return buf
}

// Attr creates a new attribute.
func Attr(name types.Ident, t Type, required bool) Attribute {
	return Attribute{Name: name, Type: t, Required: required}
}

// RequiredAttr creates a required attribute.
func RequiredAttr(name types.Ident, t Type) Attribute {
	return Attribute{Name: name, Type: t, Required: true}
}

// OptionalAttr creates an optional attribute.
func OptionalAttr(name types.Ident, t Type) Attribute {
	return Attribute{Name: name, Type: t, Required: false}
}

// Annotation represents a Cedar annotation (e.g., @doc("description")).
type Annotation struct {
	Key   types.Ident
	Value types.String
}

func (a Annotation) MarshalCedar() []byte {
	if a.Value == "" {
		return append([]byte("@"), []byte(a.Key)...)
	}
	buf := []byte("@")
	buf = append(buf, []byte(a.Key)...)
	buf = append(buf, '(')
	buf = append(buf, marshalString(a.Value)...)
	buf = append(buf, ')')
	return buf
}

// Annotations is a slice of Annotation.
type Annotations []Annotation

// Annotate adds an annotation.
func (a Annotations) Annotate(key types.Ident, value types.String) Annotations {
	return append(a, Annotation{Key: key, Value: value})
}

// marshalString produces a quoted string suitable for Cedar output.
func marshalString(s types.String) []byte {
	// Simple implementation - escape quotes and backslashes
	buf := []byte{'"'}
	for _, c := range []byte(s) {
		switch c {
		case '"':
			buf = append(buf, '\\', '"')
		case '\\':
			buf = append(buf, '\\', '\\')
		case '\n':
			buf = append(buf, '\\', 'n')
		case '\r':
			buf = append(buf, '\\', 'r')
		case '\t':
			buf = append(buf, '\\', 't')
		default:
			buf = append(buf, c)
		}
	}
	buf = append(buf, '"')
	return buf
}
