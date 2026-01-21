package ast

import "github.com/cedar-policy/cedar-go/types"

// Primitive types

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() { _ = 0 }

// Resolve returns the StringType unchanged (no references to resolve).
func (s StringType) Resolve(namespace *NamespaceNode) IsType { return s }

// String returns a StringType.
func String() StringType { return StringType{} }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() { _ = 0 }

// Resolve returns the LongType unchanged (no references to resolve).
func (l LongType) Resolve(namespace *NamespaceNode) IsType { return l }

// Long returns a LongType.
func Long() LongType { return LongType{} }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() { _ = 0 }

// Resolve returns the BoolType unchanged (no references to resolve).
func (b BoolType) Resolve(namespace *NamespaceNode) IsType { return b }

// Bool returns a BoolType.
func Bool() BoolType { return BoolType{} }

// Extension types

// ExtensionType represents a Cedar extension type (ipaddr, decimal, datetime, duration).
type ExtensionType struct {
	Name types.Ident
}

func (ExtensionType) isType() { _ = 0 }

// Resolve returns the ExtensionType unchanged (no references to resolve).
func (e ExtensionType) Resolve(namespace *NamespaceNode) IsType { return e }

// IPAddr returns an ExtensionType for ipaddr.
func IPAddr() ExtensionType { return ExtensionType{Name: "ipaddr"} }

// Decimal returns an ExtensionType for decimal.
func Decimal() ExtensionType { return ExtensionType{Name: "decimal"} }

// Datetime returns an ExtensionType for datetime.
func Datetime() ExtensionType { return ExtensionType{Name: "datetime"} }

// Duration returns an ExtensionType for duration.
func Duration() ExtensionType { return ExtensionType{Name: "duration"} }

// Collection types

// SetType represents a Cedar Set type with an element type.
type SetType struct {
	Element IsType
}

func (SetType) isType() { _ = 0 }

// Resolve returns a new SetType with the element type resolved.
func (s SetType) Resolve(namespace *NamespaceNode) IsType {
	return SetType{Element: s.Element.Resolve(namespace)}
}

// Set returns a SetType with the given element type.
func Set(element IsType) SetType {
	return SetType{Element: element}
}

// Record types

// Pair represents a key-value pair in a record type.
type Pair struct {
	Key      types.String
	Type     IsType
	Optional bool
}

// Attribute creates a required attribute pair.
func Attribute(key types.String, t IsType) Pair {
	return Pair{Key: key, Type: t, Optional: false}
}

// Optional creates an optional attribute pair.
func Optional(key types.String, t IsType) Pair {
	return Pair{Key: key, Type: t, Optional: true}
}

// RecordType represents a Cedar Record type with attributes.
type RecordType struct {
	Pairs []Pair
}

func (RecordType) isType() { _ = 0 }

// Resolve returns a new RecordType with all attribute types resolved.
func (r RecordType) Resolve(namespace *NamespaceNode) IsType {
	resolved := make([]Pair, len(r.Pairs))
	for i, p := range r.Pairs {
		resolved[i] = Pair{
			Key:      p.Key,
			Type:     p.Type.Resolve(namespace),
			Optional: p.Optional,
		}
	}
	return RecordType{Pairs: resolved}
}

// Record returns a RecordType with the given pairs.
func Record(pairs ...Pair) RecordType {
	return RecordType{Pairs: pairs}
}

// Reference types

// EntityTypeRef represents a reference to an entity type.
type EntityTypeRef struct {
	Name types.EntityType
}

func (EntityTypeRef) isType() { _ = 0 }

// Resolve resolves the entity type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it is qualified with the namespace.
func (e EntityTypeRef) Resolve(namespace *NamespaceNode) IsType {
	if namespace == nil {
		return e
	}
	// If the name doesn't contain "::", qualify it with the namespace
	name := string(e.Name)
	if len(name) > 0 && name[0] != ':' && !containsDoubleColon(name) {
		return EntityTypeRef{Name: types.EntityType(string(namespace.Name) + "::" + name)}
	}
	return e
}

func containsDoubleColon(s string) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] == ':' && s[i+1] == ':' {
			return true
		}
	}
	return false
}

// EntityType creates an EntityTypeRef from an entity type name.
func EntityType(name types.EntityType) EntityTypeRef {
	return EntityTypeRef{Name: name}
}

// Ref is an alias for EntityType for more concise syntax.
func Ref(name types.EntityType) EntityTypeRef {
	return EntityType(name)
}

// TypeRef represents a reference to a common type by name.
type TypeRef struct {
	Name types.Path
}

func (TypeRef) isType() { _ = 0 }

// Resolve resolves the type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it is qualified with the namespace.
func (t TypeRef) Resolve(namespace *NamespaceNode) IsType {
	if namespace == nil {
		return t
	}
	// If the name doesn't contain "::", qualify it with the namespace
	name := string(t.Name)
	if len(name) > 0 && name[0] != ':' && !containsDoubleColon(name) {
		return TypeRef{Name: types.Path(string(namespace.Name) + "::" + name)}
	}
	return t
}

// Type creates a TypeRef from a path name.
func Type(name types.Path) TypeRef {
	return TypeRef{Name: name}
}
