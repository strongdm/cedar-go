package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// Type represents a Cedar type for use in schema building.
type Type interface {
	toAST() ast.Type
}

type primitiveType struct {
	name string
}

func (p primitiveType) toAST() ast.Type {
	return ast.PrimitiveType{Name: p.name}
}

// String returns the Cedar String type.
func String() Type {
	return primitiveType{name: "String"}
}

// Long returns the Cedar Long type.
func Long() Type {
	return primitiveType{name: "Long"}
}

// Bool returns the Cedar Bool type.
func Bool() Type {
	return primitiveType{name: "Bool"}
}

type setType struct {
	element Type
}

func (s setType) toAST() ast.Type {
	return &ast.SetType{Element: s.element.toAST()}
}

// Set returns a Set type containing elements of the given type.
func Set(element Type) Type {
	return setType{element: element}
}

type entityRefType struct {
	name string
}

func (e entityRefType) toAST() ast.Type {
	return ast.EntityRefType{Name: e.name}
}

// EntityRef returns a reference to an entity type.
// The name can be unqualified (resolved within current namespace) or qualified.
func EntityRef(name string) Type {
	return entityRefType{name: name}
}

type recordType struct {
	attrs []*Attribute
}

func (r recordType) toAST() ast.Type {
	rt := &ast.RecordType{
		Attributes: make(map[string]*ast.Attribute),
	}
	for _, attr := range r.attrs {
		rt.Attributes[attr.name] = &ast.Attribute{
			Name:     attr.name,
			Type:     attr.typ.toAST(),
			Required: attr.required,
		}
	}
	return rt
}

// Record returns a record type with the given attributes.
func Record(attrs ...*Attribute) Type {
	return recordType{attrs: attrs}
}

type extensionType struct {
	name string
}

func (e extensionType) toAST() ast.Type {
	return ast.ExtensionType{Name: e.name}
}

// Extension returns an extension type (e.g., ipaddr, decimal).
func Extension(name string) Type {
	return extensionType{name: name}
}

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

// Attribute represents a field in a record type.
type Attribute struct {
	name     string
	typ      Type
	required bool
}

// Attr creates a required attribute with the given name and type.
func Attr(name string, t Type) *Attribute {
	return &Attribute{
		name:     name,
		typ:      t,
		required: true,
	}
}

// OptionalAttr creates an optional attribute with the given name and type.
func OptionalAttr(name string, t Type) *Attribute {
	return &Attribute{
		name:     name,
		typ:      t,
		required: false,
	}
}
