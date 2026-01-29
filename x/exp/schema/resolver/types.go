package resolver

import (
	"github.com/cedar-policy/cedar-go/types"
)

// IsType is the interface implemented by all type expressions.
//
//sumtype:decl
type IsType interface {
	isType()
}

// Primitive types

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() { _ = 0 }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() { _ = 0 }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() { _ = 0 }

// ExtensionType represents a Cedar extension type (ipaddr, decimal, datetime, duration).
type ExtensionType struct {
	Name types.Ident
}

func (ExtensionType) isType() { _ = 0 }

// Collection types

// SetType represents a Cedar Set type with an element type.
type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

// Record types
type Attribute struct {
	Type        IsType
	Optional    bool
	Annotations Annotations
}

type Attributes map[types.String]Attribute

// RecordType represents a Cedar Record type with attributes.
type RecordType struct {
	Attributes Attributes
}

func (RecordType) isType() { _ = 0 }

// EntityTypeRef represents a reference to an entity type.
type EntityTypeRef struct {
	Name types.EntityType
}

func (EntityTypeRef) isType() { _ = 0 }
