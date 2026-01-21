package ast

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
)

// Primitive types

// StringType represents the Cedar String type.
type StringType struct{}

func (StringType) isType() { _ = 0 }

// resolve returns the StringType unchanged (no references to resolve).
func (s StringType) resolve(rd *resolveData) (IsType, error) { return s, nil }

// String returns a StringType.
func String() StringType { return StringType{} }

// LongType represents the Cedar Long type.
type LongType struct{}

func (LongType) isType() { _ = 0 }

// resolve returns the LongType unchanged (no references to resolve).
func (l LongType) resolve(rd *resolveData) (IsType, error) { return l, nil }

// Long returns a LongType.
func Long() LongType { return LongType{} }

// BoolType represents the Cedar Bool type.
type BoolType struct{}

func (BoolType) isType() { _ = 0 }

// resolve returns the BoolType unchanged (no references to resolve).
func (b BoolType) resolve(rd *resolveData) (IsType, error) { return b, nil }

// Bool returns a BoolType.
func Bool() BoolType { return BoolType{} }

// Extension types

// ExtensionType represents a Cedar extension type (ipaddr, decimal, datetime, duration).
type ExtensionType struct {
	Name types.Ident
}

func (ExtensionType) isType() { _ = 0 }

// resolve returns the ExtensionType unchanged (no references to resolve).
func (e ExtensionType) resolve(rd *resolveData) (IsType, error) { return e, nil }

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

// resolve returns a new SetType with the element type resolved.
func (s SetType) resolve(rd *resolveData) (IsType, error) {
	resolved, err := s.Element.resolve(rd)
	if err != nil {
		return nil, err
	}
	return SetType{Element: resolved}, nil
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

// resolve returns a new RecordType with all attribute types resolved.
func (r RecordType) resolve(rd *resolveData) (IsType, error) {
	resolved := make([]Pair, len(r.Pairs))
	for i, p := range r.Pairs {
		resolvedType, err := p.Type.resolve(rd)
		if err != nil {
			return nil, err
		}
		resolved[i] = Pair{
			Key:      p.Key,
			Type:     resolvedType,
			Optional: p.Optional,
		}
	}
	return RecordType{Pairs: resolved}, nil
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

// resolve resolves the entity type reference relative to the given namespace.
// If the name is unqualified and namespace is provided, it is qualified with the namespace.
func (e EntityTypeRef) resolve(rd *resolveData) (IsType, error) {
	if rd.namespace == nil {
		return e, nil
	}
	// If the name doesn't contain "::", qualify it with the namespace
	name := string(e.Name)
	if len(name) > 0 && name[0] != ':' && !containsDoubleColon(name) {
		return EntityTypeRef{Name: types.EntityType(string(rd.namespace.Name) + "::" + name)}, nil
	}
	return e, nil
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

// resolve resolves the type reference relative to the given namespace and schema.
// It searches for a matching CommonType in the namespace first, then in the entire schema.
// If found, it returns the resolved concrete type. Otherwise, it returns an error.
func (t TypeRef) resolve(rd *resolveData) (IsType, error) {
	name := string(t.Name)

	// Try to find the type in the current namespace first (for unqualified names)
	if rd.namespace != nil && len(name) > 0 && name[0] != ':' && !containsDoubleColon(name) {
		// Check cache first
		if resolved, found := rd.namespaceCommonTypes[name]; found {
			return resolved, nil
		}

		// Search within the current namespace
		for ct := range rd.namespace.CommonTypes() {
			if string(ct.Name) == name {
				// Found it, resolve and cache it
				resolved, err := ct.Type.resolve(rd)
				if err != nil {
					return nil, err
				}
				rd.namespaceCommonTypes[name] = resolved
				return resolved, nil
			}
		}

		// Not found in namespace, qualify the name for schema search
		name = string(rd.namespace.Name) + "::" + name
	}

	// Check schema-wide cache
	if resolved, found := rd.schemaCommonTypes[name]; found {
		return resolved, nil
	}

	// Search the entire schema for a matching CommonType
	if rd.schema != nil {
		for ns, ct := range rd.schema.CommonTypes() {
			fullName := ct.FullName(ns)
			if string(fullName) == name {
				// Found it, resolve with the common type's namespace and cache it
				ctRd := rd.withNamespace(ns)
				resolved, err := ct.Type.resolve(ctRd)
				if err != nil {
					return nil, err
				}
				rd.schemaCommonTypes[name] = resolved
				return resolved, nil
			}
		}
	}

	// Not found, return an error
	return nil, fmt.Errorf("type %q not found", name)
}

// Type creates a TypeRef from a path name.
func Type(name types.Path) TypeRef {
	return TypeRef{Name: name}
}
