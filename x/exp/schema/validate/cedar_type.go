package validate

import (
	"fmt"
	"slices"
	"strings"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// cedarType is the sum type representing Cedar types for the type checker.
type cedarType interface {
	isCedarType()
}

type typeNever struct{}                  // bottom type, subtype of all
type typeTrue struct{}                   // singleton bool true
type typeFalse struct{}                  // singleton bool false
type typeBool struct{}                   // Bool primitive
type typeLong struct{}                   // Long primitive
type typeString struct{}                 // String primitive
type typeSet struct{ element cedarType } // Set with element type
type typeRecord struct {                 // Record with attribute types
	attrs          map[types.String]attributeType
	openAttributes bool
}
type typeEntity struct{ lub entityLUB }       // Entity with LUB of types
type typeAnyEntity struct{}                   // Unknown entity type
type typeExtension struct{ name types.Ident } // Extension type (ipaddr, decimal, etc.)

func (typeNever) isCedarType()     { _ = 0 }
func (typeTrue) isCedarType()      { _ = 0 }
func (typeFalse) isCedarType()     { _ = 0 }
func (typeBool) isCedarType()      { _ = 0 }
func (typeLong) isCedarType()      { _ = 0 }
func (typeString) isCedarType()    { _ = 0 }
func (typeSet) isCedarType()       { _ = 0 }
func (typeRecord) isCedarType()    { _ = 0 }
func (typeEntity) isCedarType()    { _ = 0 }
func (typeAnyEntity) isCedarType() { _ = 0 }
func (typeExtension) isCedarType() { _ = 0 }

type attributeType struct {
	typ      cedarType
	required bool
}

// entityLUB represents the least upper bound of a set of entity types.
type entityLUB struct {
	elements []types.EntityType // sorted, unique
}

func newEntityLUB(types ...types.EntityType) entityLUB {
	elems := slices.Clone(types)
	slices.Sort(elems)
	elems = slices.Compact(elems)
	return entityLUB{elements: elems}
}

func singleEntityLUB(et types.EntityType) entityLUB {
	return entityLUB{elements: []types.EntityType{et}}
}

// isSubtype returns true if a is a subtype of b.
func (v *Validator) isSubtype(a, b cedarType) bool {
	// Never is subtype of everything
	if _, ok := a.(typeNever); ok {
		return true
	}
	switch bv := b.(type) {
	case typeNever:
		return false
	case typeTrue:
		_, ok := a.(typeTrue)
		return ok
	case typeFalse:
		_, ok := a.(typeFalse)
		return ok
	case typeBool:
		switch a.(type) {
		case typeBool, typeTrue, typeFalse:
			return true
		}
		return false
	case typeLong:
		_, ok := a.(typeLong)
		return ok
	case typeString:
		_, ok := a.(typeString)
		return ok
	case typeSet:
		av, ok := a.(typeSet)
		if !ok {
			return false
		}
		return v.isSubtype(av.element, bv.element)
	case typeRecord:
		av, ok := a.(typeRecord)
		if !ok {
			return false
		}
		return v.isSubtypeRecord(av, bv)
	case typeEntity:
		av, ok := a.(typeEntity)
		if !ok {
			if _, ok := a.(typeAnyEntity); ok {
				return false
			}
			return false
		}
		if v.strict {
			return equalLUB(av.lub, bv.lub)
		}
		return isSubsetLUB(av.lub, bv.lub)
	case typeAnyEntity:
		if v.strict {
			_, ok := a.(typeAnyEntity)
			return ok
		}
		switch a.(type) {
		case typeEntity, typeAnyEntity:
			return true
		}
		return false
	case typeExtension:
		av, ok := a.(typeExtension)
		if !ok {
			return false
		}
		return av.name == bv.name
	}
	return false
}

func (v *Validator) isSubtypeRecord(a, b typeRecord) bool {
	// Strict: no width subtyping (always reject extra attrs)
	// Permissive: only reject extras when b is closed
	if v.strict || !b.openAttributes {
		for k := range a.attrs {
			if _, ok := b.attrs[k]; !ok {
				return false
			}
		}
	}
	for k, bAttr := range b.attrs {
		aAttr, ok := a.attrs[k]
		if !ok {
			if bAttr.required {
				return false
			}
			continue
		}
		if !v.isSubtype(aAttr.typ, bAttr.typ) {
			return false
		}
		if v.strict {
			// Strict: required/optional must match exactly
			if aAttr.required != bAttr.required {
				return false
			}
		} else {
			// Permissive: only fail if b requires but a doesn't
			if bAttr.required && !aAttr.required {
				return false
			}
		}
	}
	return true
}

func equalLUB(a, b entityLUB) bool {
	return slices.Equal(a.elements, b.elements)
}

func isSubsetLUB(a, b entityLUB) bool {
	for _, ae := range a.elements {
		if !slices.Contains(b.elements, ae) {
			return false
		}
	}
	return true
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
		}
	case typeFalse:
		switch b.(type) {
		case typeFalse:
			return typeFalse{}, nil
		case typeTrue, typeBool:
			return typeBool{}, nil
		}
	case typeBool:
		switch b.(type) {
		case typeTrue, typeFalse, typeBool:
			return typeBool{}, nil
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
		switch bv := b.(type) {
		case typeEntity:
			return typeEntity{lub: unionLUB(av.lub, bv.lub)}, nil
		case typeAnyEntity:
			return typeAnyEntity{}, nil
		}
	case typeAnyEntity:
		switch b.(type) {
		case typeEntity, typeAnyEntity:
			return typeAnyEntity{}, nil
		}
	case typeExtension:
		if bv, ok := b.(typeExtension); ok && av.name == bv.name {
			return av, nil
		}
	}

	return nil, fmt.Errorf("incompatible types for least upper bound")
}

func (v *Validator) lubRecord(a, b typeRecord) (cedarType, error) {
	attrs := make(map[types.String]attributeType)
	// Attributes in both
	for k, aAttr := range a.attrs {
		if bAttr, ok := b.attrs[k]; ok {
			lub, err := v.leastUpperBound(aAttr.typ, bAttr.typ)
			if err != nil {
				return nil, err
			}
			if v.strict && aAttr.required != bAttr.required {
				return nil, fmt.Errorf("incompatible record types for least upper bound")
			}
			attrs[k] = attributeType{
				typ:      lub,
				required: aAttr.required && bAttr.required,
			}
		} else {
			if v.strict {
				return nil, fmt.Errorf("incompatible record types for least upper bound")
			}
			attrs[k] = attributeType{typ: aAttr.typ, required: false}
		}
	}
	for k, bAttr := range b.attrs {
		if _, ok := a.attrs[k]; !ok {
			if v.strict {
				return nil, fmt.Errorf("incompatible record types for least upper bound")
			}
			attrs[k] = attributeType{typ: bAttr.typ, required: false}
		}
	}
	return typeRecord{
		attrs:          attrs,
		openAttributes: a.openAttributes || b.openAttributes,
	}, nil
}

func unionLUB(a, b entityLUB) entityLUB {
	combined := append(slices.Clone(a.elements), b.elements...)
	slices.Sort(combined)
	combined = slices.Compact(combined)
	return entityLUB{elements: combined}
}

// schemaTypeToCedarType converts a resolved schema type to a cedarType.
func schemaTypeToCedarType(t resolved.IsType) cedarType {
	switch t := t.(type) {
	case resolved.StringType:
		return typeString{}
	case resolved.LongType:
		return typeLong{}
	case resolved.BoolType:
		return typeBool{}
	case resolved.ExtensionType:
		return typeExtension{name: types.Ident(t)}
	case resolved.SetType:
		return typeSet{element: schemaTypeToCedarType(t.Element)}
	case resolved.RecordType:
		return schemaRecordToCedarType(t)
	case resolved.EntityType:
		return typeEntity{lub: singleEntityLUB(types.EntityType(t))}
	default:
		return typeNever{}
	}
}

func schemaRecordToCedarType(rec resolved.RecordType) typeRecord {
	attrs := make(map[types.String]attributeType, len(rec))
	for name, attr := range rec {
		attrs[name] = attributeType{
			typ:      schemaTypeToCedarType(attr.Type),
			required: !attr.Optional,
		}
	}
	return typeRecord{attrs: attrs, openAttributes: false}
}

// lookupAttributeType looks up an attribute on a type using schema information.
func (v *Validator) lookupAttributeType(ty cedarType, attr types.String) *attributeType {
	switch tv := ty.(type) {
	case typeRecord:
		if a, ok := tv.attrs[attr]; ok {
			return &a
		}
		return nil
	case typeEntity:
		return v.lookupEntityAttr(tv.lub, attr)
	default:
		return nil
	}
}

func (v *Validator) lookupEntityAttr(lub entityLUB, attr types.String) *attributeType {
	if len(lub.elements) == 0 {
		return nil
	}
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

// mayHaveAttr returns true if the type might have the given attribute.
func (v *Validator) mayHaveAttr(ty cedarType, attr types.String) bool {
	switch tv := ty.(type) {
	case typeRecord:
		if tv.openAttributes {
			return true
		}
		_, ok := tv.attrs[attr]
		return ok
	case typeEntity:
		return v.mayEntityHaveAttr(tv.lub, attr)
	case typeAnyEntity:
		return true
	default:
		return false
	}
}

func (v *Validator) mayEntityHaveAttr(lub entityLUB, attr types.String) bool {
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			continue
		}
		if _, ok := entity.Shape[attr]; ok {
			return true
		}
	}
	return false
}

// entityHasTags returns true if all entities in the LUB have tags defined.
func (v *Validator) entityHasTags(lub entityLUB) bool {
	if len(lub.elements) == 0 {
		return false
	}
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			return false
		}
		if entity.Tags == nil {
			return false
		}
	}
	return true
}

// entityTagType returns the LUB of the tag types for all entities in the LUB.
func (v *Validator) entityTagType(lub entityLUB) cedarType {
	if len(lub.elements) == 0 {
		return typeNever{}
	}
	var result cedarType = typeNever{}
	for _, et := range lub.elements {
		entity, ok := v.schema.Entities[et]
		if !ok {
			return typeNever{}
		}
		if entity.Tags == nil {
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
	if _, ok := b.(typeNever); ok {
		return nil
	}
	ae, aOk := a.(typeEntity)
	be, bOk := b.(typeEntity)
	if !aOk || !bOk {
		return nil
	}
	if !v.entityLUBsRelated(ae.lub, be.lub) {
		return fmt.Errorf("entity types are incompatible in strict mode")
	}
	return nil
}

// entityLUBsRelated returns true if any entity type in LUB a is related to
// any entity type in LUB b (same type, or ancestor/descendant relationship).
func (v *Validator) entityLUBsRelated(a, b entityLUB) bool {
	for _, at := range a.elements {
		for _, bt := range b.elements {
			if at == bt {
				return true
			}
			if v.isEntityDescendant(at, bt) || v.isEntityDescendant(bt, at) {
				return true
			}
		}
	}
	return false
}

// isEntityDescendant returns true if childType can be a descendant (member) of ancestorType.
// This means childType lists ancestorType (directly or transitively) in its ParentTypes.
func (v *Validator) isEntityDescendant(childType, ancestorType types.EntityType) bool {
	entity, ok := v.schema.Entities[childType]
	if !ok {
		return false
	}
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

// isActionEntity returns true if the entity type is an action type.
func isActionEntity(et types.EntityType) bool {
	s := string(et)
	return s == "Action" || strings.HasSuffix(s, "::Action")
}
