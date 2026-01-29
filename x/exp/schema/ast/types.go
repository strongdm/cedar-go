package ast

import (
	"github.com/cedar-policy/cedar-go/types"
)

//sumtype:decl
type IsType interface {
	isType()
}

type StringType struct{}

func (StringType) isType() { _ = 0 }

func String() StringType { return StringType{} }

type LongType struct{}

func (LongType) isType() { _ = 0 }

func Long() LongType { return LongType{} }

type BoolType struct{}

func (BoolType) isType() { _ = 0 }

func Bool() BoolType { return BoolType{} }

type ExtensionType struct {
	Name types.Ident
}

func (ExtensionType) isType() { _ = 0 }

func IPAddr() ExtensionType { return ExtensionType{Name: "ipaddr"} }

func Decimal() ExtensionType { return ExtensionType{Name: "decimal"} }

func Datetime() ExtensionType { return ExtensionType{Name: "datetime"} }

func Duration() ExtensionType { return ExtensionType{Name: "duration"} }

type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

func Set(element IsType) SetType {
	return SetType{Element: element}
}

type Attribute struct {
	Type        IsType
	Optional    bool
	Annotations Annotations
}

type Attributes map[types.String]Attribute

type RecordType struct {
	Attributes Attributes
}

func (RecordType) isType() { _ = 0 }

func Record(attrs Attributes) RecordType {
	return RecordType{Attributes: attrs}
}

type EntityTypeRef struct {
	Name types.EntityType
}

func (EntityTypeRef) isType() { _ = 0 }

func EntityType(name types.EntityType) EntityTypeRef {
	return EntityTypeRef{Name: name}
}

type TypeRef struct {
	Name types.Path
}

func (TypeRef) isType() { _ = 0 }

func Type(name types.Path) TypeRef {
	return TypeRef{Name: name}
}
