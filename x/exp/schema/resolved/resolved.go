package resolved

import "github.com/cedar-policy/cedar-go/types"

// Schema is the result of calling schema.Schema.Resolve().
// All type references are fully-qualified.
type Schema struct {
	Namespaces map[types.Path]*Namespace
}

// Namespace contains resolved definitions.
type Namespace struct {
	EntityTypes map[types.EntityType]*EntityType
	EnumTypes   map[types.EntityType]*EnumType
	Actions     map[types.EntityUID]*Action
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
	Values      []string
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
type Type interface {
	resolvedType()
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

func (Primitive) resolvedType() {}

// Set is a resolved Set<T> type.
type Set struct {
	Element Type
}

func (Set) resolvedType() {}

// RecordType is a resolved record type.
type RecordType struct {
	Attributes map[string]*Attribute
}

func (*RecordType) resolvedType() {}

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

func (EntityRef) resolvedType() {}

// Extension is a resolved extension type.
type Extension struct {
	Name string
}

func (Extension) resolvedType() {}

// Annotations are key-value metadata attached to schema elements.
type Annotations map[string]string
