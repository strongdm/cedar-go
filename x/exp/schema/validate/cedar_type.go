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

type typeNever struct{}                                        // bottom type, subtype of all
type typeTrue struct{}                                         // singleton bool true
type typeFalse struct{}                                        // singleton bool false
type typeBool struct{}                                         // Bool primitive
type typeLong struct{}                                         // Long primitive
type typeString struct{}                                       // String primitive
type typeSet struct{ element cedarType }                       // Set with element type
type typeRecord struct {                                       // Record with attribute types
	attrs          map[types.String]attributeType
	openAttributes bool
}
type typeEntity struct{ lub entityLUB }                        // Entity with LUB of types
type typeAnyEntity struct{}                                    // Unknown entity type
type typeExtension struct{ name types.Ident }                  // Extension type (ipaddr, decimal, etc.)

func (typeNever) isCedarType()     {}
func (typeTrue) isCedarType()      {}
func (typeFalse) isCedarType()     {}
func (typeBool) isCedarType()      {}
func (typeLong) isCedarType()      {}
func (typeString) isCedarType()    {}
func (typeSet) isCedarType()       {}
func (typeRecord) isCedarType()    {}
func (typeEntity) isCedarType()    {}
func (typeAnyEntity) isCedarType() {}
func (typeExtension) isCedarType() {}

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
func isSubtype(a, b cedarType) bool {
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
		return isSubtype(av.element, bv.element)
	case typeRecord:
		av, ok := a.(typeRecord)
		if !ok {
			return false
		}
		return isSubtypeRecord(av, bv)
	case typeEntity:
		av, ok := a.(typeEntity)
		if !ok {
			if _, ok := a.(typeAnyEntity); ok {
				return false
			}
			return false
		}
		return isSubsetLUB(av.lub, bv.lub)
	case typeAnyEntity:
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

func isSubtypeRecord(a, b typeRecord) bool {
	// If b is open, a just needs to have all of b's required attrs as subtypes
	// If b is closed, a must not have extra attrs
	if !b.openAttributes {
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
		if !isSubtype(aAttr.typ, bAttr.typ) {
			return false
		}
		// If b requires it, a must also require it
		if bAttr.required && !aAttr.required {
			return false
		}
	}
	return true
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
func leastUpperBound(a, b cedarType) (cedarType, error) {
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
			elem, err := leastUpperBound(av.element, bv.element)
			if err != nil {
				return nil, err
			}
			return typeSet{element: elem}, nil
		}
	case typeRecord:
		if bv, ok := b.(typeRecord); ok {
			return lubRecord(av, bv)
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

func lubRecord(a, b typeRecord) (cedarType, error) {
	attrs := make(map[types.String]attributeType)
	// Attributes in both
	for k, aAttr := range a.attrs {
		if bAttr, ok := b.attrs[k]; ok {
			lub, err := leastUpperBound(aAttr.typ, bAttr.typ)
			if err != nil {
				return nil, err
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
func lookupAttributeType(s *resolved.Schema, ty cedarType, attr types.String) *attributeType {
	switch tv := ty.(type) {
	case typeRecord:
		if a, ok := tv.attrs[attr]; ok {
			return &a
		}
		return nil
	case typeEntity:
		return lookupEntityAttr(s, tv.lub, attr)
	default:
		return nil
	}
}

func lookupEntityAttr(s *resolved.Schema, lub entityLUB, attr types.String) *attributeType {
	if len(lub.elements) == 0 {
		return nil
	}
	var result *attributeType
	for _, et := range lub.elements {
		entity, ok := s.Entities[et]
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
			lub, err := leastUpperBound(result.typ, at.typ)
			if err != nil {
				return nil
			}
			result = &attributeType{
				typ:      lub,
				required: result.required && at.required,
			}
		}
	}
	return result
}

// mayHaveAttr returns true if the type might have the given attribute.
func mayHaveAttr(s *resolved.Schema, ty cedarType, attr types.String) bool {
	switch tv := ty.(type) {
	case typeRecord:
		if tv.openAttributes {
			return true
		}
		_, ok := tv.attrs[attr]
		return ok
	case typeEntity:
		return mayEntityHaveAttr(s, tv.lub, attr)
	case typeAnyEntity:
		return true
	default:
		return false
	}
}

func mayEntityHaveAttr(s *resolved.Schema, lub entityLUB, attr types.String) bool {
	for _, et := range lub.elements {
		entity, ok := s.Entities[et]
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
func entityHasTags(s *resolved.Schema, lub entityLUB) bool {
	if len(lub.elements) == 0 {
		return false
	}
	for _, et := range lub.elements {
		entity, ok := s.Entities[et]
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
func entityTagType(s *resolved.Schema, lub entityLUB) cedarType {
	if len(lub.elements) == 0 {
		return typeNever{}
	}
	var result cedarType = typeNever{}
	for _, et := range lub.elements {
		entity, ok := s.Entities[et]
		if !ok {
			return typeNever{}
		}
		if entity.Tags == nil {
			return typeNever{}
		}
		tagType := schemaTypeToCedarType(entity.Tags)
		lub, err := leastUpperBound(result, tagType)
		if err != nil {
			return typeNever{}
		}
		result = lub
	}
	return result
}

// checkStrictEntityLUB checks if two types have compatible entity types in strict mode.
// In strict mode, entity LUBs between unrelated entity types are disallowed.
func checkStrictEntityLUB(s *resolved.Schema, a, b cedarType) error {
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
	if !entityLUBsRelated(s, ae.lub, be.lub) {
		return fmt.Errorf("entity types are incompatible in strict mode")
	}
	return nil
}

// entityLUBsRelated returns true if any entity type in LUB a is related to
// any entity type in LUB b (same type, or ancestor/descendant relationship).
func entityLUBsRelated(s *resolved.Schema, a, b entityLUB) bool {
	for _, at := range a.elements {
		for _, bt := range b.elements {
			if at == bt {
				return true
			}
			if isEntityDescendant(s, at, bt) || isEntityDescendant(s, bt, at) {
				return true
			}
		}
	}
	return false
}

// isEntityDescendant returns true if childType can be a descendant (member) of ancestorType.
// This means childType lists ancestorType (directly or transitively) in its ParentTypes.
func isEntityDescendant(s *resolved.Schema, childType, ancestorType types.EntityType) bool {
	entity, ok := s.Entities[childType]
	if !ok {
		return false
	}
	for _, parent := range entity.ParentTypes {
		if parent == ancestorType {
			return true
		}
		if isEntityDescendant(s, parent, ancestorType) {
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
