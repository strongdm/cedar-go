package resolver

import (
	"github.com/cedar-policy/cedar-go/types"
)

//sumtype:decl
type IsType interface {
	isType()
}

type StringType struct{}

func (StringType) isType() { _ = 0 }

type LongType struct{}

func (LongType) isType() { _ = 0 }

type BoolType struct{}

func (BoolType) isType() { _ = 0 }

type ExtensionType types.Ident

func (ExtensionType) isType() { _ = 0 }

type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

type Attribute struct {
	Type        IsType
	Optional    bool
	Annotations Annotations
}

type RecordType map[types.String]Attribute

func (RecordType) isType() { _ = 0 }

type EntityType types.EntityType

func (EntityType) isType() { _ = 0 }
