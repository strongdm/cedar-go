// Package resolved contains the resolved schema types produced by schema.Resolve().
//
// A resolved schema has all type references validated and converted to fully-qualified
// types.EntityType and types.EntityUID values. Common types are inlined.
//
// Example:
//
//	resolved, err := schema.Resolve()
//	if err != nil {
//		log.Fatal(err)
//	}
//	for nsPath, ns := range resolved.Namespaces {
//		for entityType := range ns.EntityTypes {
//			fmt.Printf("%s::%s\n", nsPath, entityType)
//		}
//	}
package resolved

import "github.com/cedar-policy/cedar-go/types"

// Schema is the result of calling schema.Resolve(). All type references
// have been validated and converted to fully-qualified types.EntityType values.
type Schema struct {
	Namespaces map[types.Path]*Namespace
}

// Namespace contains resolved definitions.
type Namespace struct {
	EntityTypes map[types.EntityType]*EntityType
	EnumTypes   map[types.EntityType]*EnumType
	Actions     map[types.EntityUID]*Action
	CommonTypes map[types.Path]*Type // inlined during resolution
	Annotations Annotations
}

// EntityType has parent types as types.EntityType.
type EntityType struct {
	MemberOfTypes []types.EntityType
	Shape         *RecordType
	Tags          Type
	Annotations   Annotations
}

// EnumType represents a resolved enumerated entity type.
type EnumType struct {
	Values      []string // the allowed entity IDs
	Annotations Annotations
}

// Action has principal/resource types as types.EntityType.
type Action struct {
	MemberOf       []types.EntityUID
	PrincipalTypes []types.EntityType
	ResourceTypes  []types.EntityType
	Context        *RecordType
	Annotations    Annotations
}

// Type represents a fully-resolved Cedar type.
// Implementations: Primitive, Set, *RecordType, EntityRef, Extension
type Type interface {
	isResolvedType()
}

// PrimitiveKind represents Cedar primitive types.
type PrimitiveKind int

const (
	PrimitiveLong PrimitiveKind = iota
	PrimitiveString
	PrimitiveBool
)

func (k PrimitiveKind) String() string {
	switch k {
	case PrimitiveLong:
		return "Long"
	case PrimitiveString:
		return "String"
	case PrimitiveBool:
		return "Bool"
	default:
		return "Unknown"
	}
}

// Primitive is a resolved primitive type.
type Primitive struct {
	Kind PrimitiveKind
}

func (Primitive) isResolvedType() {}

// Set is a resolved Set<T> type.
type Set struct {
	Element Type
}

func (Set) isResolvedType() {}

// RecordType is a resolved record type.
type RecordType struct {
	Attributes map[string]*Attribute
}

func (*RecordType) isResolvedType() {}

// Attribute is a resolved attribute in a record.
type Attribute struct {
	Type        Type
	Required    bool
	Annotations Annotations
}

// EntityRef references a resolved entity type.
type EntityRef struct {
	EntityType types.EntityType
}

func (EntityRef) isResolvedType() {}

// Extension is a resolved extension type.
type Extension struct {
	Name string // e.g., "ipaddr"
}

func (Extension) isResolvedType() {}

// Annotations are key-value metadata attached to schema elements.
type Annotations map[string]string
