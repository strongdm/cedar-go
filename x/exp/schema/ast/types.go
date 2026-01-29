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

type ExtensionType types.Ident

func (ExtensionType) isType() { _ = 0 }

func IPAddr() ExtensionType { return ExtensionType("ipaddr") }

func Decimal() ExtensionType { return ExtensionType("decimal") }

func Datetime() ExtensionType { return ExtensionType("datetime") }

func Duration() ExtensionType { return ExtensionType("duration") }

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

type RecordType map[types.String]Attribute

func (RecordType) isType() { _ = 0 }

type EntityTypeRef types.EntityType

func (EntityTypeRef) isType() { _ = 0 }

func EntityType(name types.EntityType) EntityTypeRef {
	return EntityTypeRef(name)
}

type TypeRef types.Path

func (TypeRef) isType() { _ = 0 }

func Type(name types.Path) TypeRef {
	return TypeRef(name)
}
