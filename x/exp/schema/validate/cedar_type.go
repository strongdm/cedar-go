package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// cedarType represents Cedar types for the type checker.
// Implementations: typeNever, typeTrue, typeFalse, typeBool, typeLong,
// typeString, typeSet, typeRecord, typeEntity, typeExtension.
type cedarType = any

type typeNever struct{}                  // bottom type, subtype of all
type typeTrue struct{}                   // singleton bool true
type typeFalse struct{}                  // singleton bool false
type typeBool struct{}                   // Bool primitive
type typeLong struct{}                   // Long primitive
type typeString struct{}                 // String primitive
type typeSet struct{ element cedarType } // Set with element type
type typeRecord struct {                 // Record with attribute types
	attrs map[types.String]attributeType
}
type typeEntity struct{ lub entityLUB }       // Entity with LUB of types
type typeExtension struct{ name types.Ident } // Extension type (ipaddr, decimal, etc.)

type attributeType struct {
	typ      cedarType
	required bool
}

// entityLUB represents the least upper bound (LUB) of a set of entity types.
// In type theory, the LUB is the most specific type that is a supertype of all
// given types. For entities, this is the union of the entity type names: e.g.
// the LUB of User and Admin is {User, Admin}. Elements are stored sorted for
// deterministic equality comparison.
type entityLUB struct {
	elements []types.EntityType // sorted, unique
}

// singleEntityLUB is an optimized constructor for the common single-element case,
// avoiding the clone/sort/compact overhead of newEntityLUB.
func singleEntityLUB(et types.EntityType) entityLUB {
	return entityLUB{elements: []types.EntityType{et}}
}

// isDisjoint returns true if the two entity LUBs have no entity types in common.
func (a entityLUB) isDisjoint(b entityLUB) bool {
	// Both LUBs are sorted, so we can check for intersection efficiently
	i, j := 0, 0
	for i < len(a.elements) && j < len(b.elements) {
		if a.elements[i] == b.elements[j] {
			return false // found a common element
		}
		if a.elements[i] < b.elements[j] {
			i++
		} else {
			j++
		}
	}
	return true // no common elements found
}

// isSubtype returns true if a is a subtype of b.
// Only called from extension function argument type checking,
// which only uses typeString and typeExtension argument types.
func (v *Validator) isSubtype(a, b cedarType) bool {
	switch bv := b.(type) {
	case typeString:
		_, ok := a.(typeString)
		return ok
	default:
		av, ok := a.(typeExtension)
		return ok && av.name == bv.(typeExtension).name
	}
}

// leastUpperBound computes the LUB of two types.
func (v *Validator) leastUpperBound(a, b cedarType) (cedarType, error) {
	if _, ok := a.(typeNever); ok {
		return b, nil
	}
	if _, ok := b.(typeNever); ok {
		return a, nil
	}

	switch av := a.(type) {
	case typeTrue:
		switch b.(type) {
		case typeTrue:
			return typeTrue{}, nil
		case typeFalse, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeFalse:
		switch b.(type) {
		case typeFalse:
			return typeFalse{}, nil
		case typeTrue, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeBool:
		switch b.(type) {
		case typeTrue, typeFalse, typeBool:
			return typeBool{}, nil
		case typeNever, typeLong, typeString, typeSet, typeRecord, typeEntity, typeExtension:
		}
	case typeLong:
		if _, ok := b.(typeLong); ok {
			return typeLong{}, nil
		}
	case typeString:
		if _, ok := b.(typeString); ok {
			return typeString{}, nil
		}
	case typeSet:
		if bv, ok := b.(typeSet); ok {
			elem, err := v.leastUpperBound(av.element, bv.element)
			if err != nil {
				return nil, err
			}
			return typeSet{element: elem}, nil
		}
	case typeRecord:
		if bv, ok := b.(typeRecord); ok {
			return v.lubRecord(av, bv)
		}
	case typeEntity:
		if bv, ok := b.(typeEntity); ok {
			return typeEntity{lub: unionLUB(av.lub, bv.lub)}, nil
		}
	case typeExtension:
		if bv, ok := b.(typeExtension); ok && av.name == bv.name {
			return av, nil
		}
	case typeNever:
		// typeNever handled above
	}

	return nil, fmt.Errorf("incompatible types for least upper bound")
}

func (v *Validator) lubRecord(a, b typeRecord) (cedarType, error) {
	// Strict mode: records with different key sets cannot be combined (no width subtyping)
	if v.strict {
		if len(a.attrs) != len(b.attrs) {
			return nil, fmt.Errorf("record types have different attributes in strict mode")
		}
		for k := range a.attrs {
			if _, ok := b.attrs[k]; !ok {
				return nil, fmt.Errorf("record types have different attributes in strict mode")
			}
		}
	}

	attrs := make(map[types.String]attributeType)
	// Attributes in both
	for k, aAttr := range a.attrs {
		if bAttr, ok := b.attrs[k]; ok {
			lub, err := v.leastUpperBound(aAttr.typ, bAttr.typ)
			if err != nil {
				if v.strict {
					return nil, err
				}
				// Permissive mode: drop attributes with incompatible types
				continue
			}
			attrs[k] = attributeType{
				typ:      lub,
				required: aAttr.required && bAttr.required,
			}
		} else {
			attrs[k] = attributeType{typ: aAttr.typ, required: false}
		}
	}
	for k, bAttr := range b.attrs {
		if _, ok := a.attrs[k]; !ok {
			attrs[k] = attributeType{typ: bAttr.typ, required: false}
		}
	}
	return typeRecord{attrs: attrs}, nil
}

func unionLUB(a, b entityLUB) entityLUB {
	combined := append(slices.Clone(a.elements), b.elements...)
	slices.Sort(combined)
	combined = slices.Compact(combined)
	return entityLUB{elements: combined}
}

// schemaTypeToCedarType converts a resolved schema type to a cedarType.
func schemaTypeToCedarType(t resolved.IsType) cedarType {
	var result cedarType
	switch t := t.(type) {
	case resolved.StringType:
		result = typeString{}
	case resolved.LongType:
		result = typeLong{}
	case resolved.BoolType:
		result = typeBool{}
	case resolved.ExtensionType:
		result = typeExtension{name: types.Ident(t)}
	case resolved.SetType:
		result = typeSet{element: schemaTypeToCedarType(t.Element)}
	case resolved.RecordType:
		result = schemaRecordToCedarType(t)
	case resolved.EntityType:
		result = typeEntity{lub: singleEntityLUB(types.EntityType(t))}
	}
	return result
}

func schemaRecordToCedarType(rec resolved.RecordType) typeRecord {
	attrs := make(map[types.String]attributeType, len(rec))
	for name, attr := range rec {
		attrs[name] = attributeType{
			typ:      schemaTypeToCedarType(attr.Type),
			required: !attr.Optional,
		}
	}
	return typeRecord{attrs: attrs}
}

// lookupAttributeType looks up an attribute on a type using schema information.
// Called only when ty is already known to be a record or entity type.
func (v *Validator) lookupAttributeType(ty cedarType, attr types.String) *attributeType {
	if tv, ok := ty.(typeRecord); ok {
		if a, ok := tv.attrs[attr]; ok {
			return &a
		}
		return nil
	}
	// Only called with typeRecord or typeEntity
	return v.lookupEntityAttr(ty.(typeEntity).lub, attr)
}

func (v *Validator) lookupEntityAttr(lub entityLUB, attr types.String) *attributeType {
	var result *attributeType
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			return nil
		}
		schemaAttr, ok := entity.Shape[attr]
		if !ok {
			return nil
		}
		at := &attributeType{
			typ:      schemaTypeToCedarType(schemaAttr.Type),
			required: !schemaAttr.Optional,
		}
		if result == nil {
			result = at
		} else {
			lubType, err := v.leastUpperBound(result.typ, at.typ)
			if err != nil {
				return nil
			}
			result = &attributeType{
				typ:      lubType,
				required: result.required && at.required,
			}
		}
	}
	return result
}

// entityHasTags returns true if all entities in the LUB have tags defined.
func (v *Validator) entityHasTags(lub entityLUB) bool {
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			return false
		}
		if entity.Tags == nil {
			return false
		}
	}
	return len(lub.elements) > 0
}

// entityTagType returns the LUB of the tag types for all entities in the LUB.
func (v *Validator) entityTagType(lub entityLUB) cedarType {
	var result cedarType = typeNever{}
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok || entity.Tags == nil {
			return typeNever{}
		}
		tagType := schemaTypeToCedarType(entity.Tags)
		tagLUB, err := v.leastUpperBound(result, tagType)
		if err != nil {
			return typeNever{}
		}
		result = tagLUB
	}
	return result
}

// checkStrictEntityLUB checks if two types have compatible entity types in strict mode.
// In strict mode, entity LUBs between unrelated entity types are disallowed.
func (v *Validator) checkStrictEntityLUB(a, b cedarType) error {
	if !v.strict {
		return nil
	}
	if _, ok := a.(typeNever); ok {
		return nil
	}
	ae, aOk := a.(typeEntity)
	be, bOk := b.(typeEntity)
	if !aOk || !bOk {
		return nil
	}
	if !entityLUBsRelated(ae.lub, be.lub) {
		return fmt.Errorf("entity types are incompatible in strict mode")
	}
	return nil
}

// entityLUBsRelated returns true if the two entity LUBs share at least one
// common entity type. In strict mode, only exact type matches are considered
// compatible (not ancestor/descendant relationships).
func entityLUBsRelated(a, b entityLUB) bool {
	for _, at := range a.elements {
		for _, bt := range b.elements {
			if at == bt {
				return true
			}
		}
	}
	return false
}

// isEntityDescendant returns true if childType can be a descendant (member) of ancestorType.
// This means childType lists ancestorType (directly or transitively) in its ParentTypes.
func (v *Validator) isEntityDescendant(childType, ancestorType types.EntityType) bool {
	entity := v.schema.Entities[childType]
	for _, parent := range entity.ParentTypes {
		if parent == ancestorType {
			return true
		}
		if v.isEntityDescendant(parent, ancestorType) {
			return true
		}
	}
	return false
}

// anyEntityDescendantOf returns true if any entity type in lhs can be a
// descendant (member) of any entity type in rhs, or if lhs and rhs share a
// common entity type (same type means "in" can be true for the same entity).
func (v *Validator) anyEntityDescendantOf(lhs, rhs entityLUB) bool {
	for _, lt := range lhs.elements {
		for _, rt := range rhs.elements {
			if lt == rt {
				return true
			}
			if v.isEntityDescendant(lt, rt) {
				return true
			}
		}
	}
	return false
}

// isActionEntity returns true if the entity type is an action type.
func isActionEntity(et types.EntityType) bool {
	s := string(et)
	return s == "Action" || strings.HasSuffix(s, "::Action")
}
