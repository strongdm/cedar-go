package schema

// TypeExpr represents a type expression in an unresolved Cedar schema.
// Use the constructor functions (Long, String, Bool, Set, Record, Entity,
// Extension, IPAddr, Decimal, Datetime, Duration, NamedType) to create values.
type TypeExpr interface {
	typeExpr()
}

// PrimitiveKind identifies a Cedar primitive type.
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

// PrimitiveTypeExpr represents Long, String, or Bool.
type PrimitiveTypeExpr struct {
	Kind PrimitiveKind
}

func (PrimitiveTypeExpr) typeExpr() {}

// SetTypeExpr represents Set<T>.
type SetTypeExpr struct {
	Element TypeExpr
}

func (SetTypeExpr) typeExpr() {}

// RecordTypeExpr represents a record with named attributes.
type RecordTypeExpr struct {
	Attributes map[string]*Attribute
}

func (*RecordTypeExpr) typeExpr() {}

// Attribute is a named field in a record type.
type Attribute struct {
	Type        TypeExpr
	Required    bool
	Annotations Annotations
}

// EntityRefExpr is an explicit entity type reference.
// In JSON this comes from {"type": "Entity", "name": "..."}.
type EntityRefExpr struct {
	Name string
}

func (EntityRefExpr) typeExpr() {}

// ExtensionTypeExpr represents an extension type (ipaddr, decimal, datetime, duration).
type ExtensionTypeExpr struct {
	Name string
}

func (ExtensionTypeExpr) typeExpr() {}

// EntityNameExpr is an unresolved name in an entity-only position
// (memberOfTypes, principalTypes, resourceTypes). Resolves only against entity types.
type EntityNameExpr struct {
	Name string
}

func (EntityNameExpr) typeExpr() {}

// TypeNameExpr is an unresolved name in a type position (attribute types,
// common type bodies). Resolves with priority: common > entity > primitive/extension.
type TypeNameExpr struct {
	Name string
}

func (TypeNameExpr) typeExpr() {}
