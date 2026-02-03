package schema2

import (
	"fmt"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/resolve"
)

// Resolve converts the schema to a ResolvedSchema with fully-qualified names,
// inlined common types, and computed entity hierarchies.
//
// Resolution performs validation and returns an error if:
//   - Any referenced type is not defined
//   - There are cycles in the entity hierarchy
//   - Other structural issues are detected
func (s *Schema) Resolve() (*ResolvedSchema, error) {
	result, err := resolve.Resolve(s.ast)
	if err != nil {
		return nil, err
	}

	rs := &ResolvedSchema{
		entityTypes: make(map[types.EntityType]*ResolvedEntityType),
		actions:     make(map[types.EntityUID]*ResolvedAction),
	}

	// Convert resolved entity types
	for name, et := range result.EntityTypes {
		rs.entityTypes[name] = convertEntityType(et)
	}

	// Convert resolved actions
	for uid, a := range result.Actions {
		rs.actions[uid] = convertAction(a)
	}

	return rs, nil
}

// MustResolve converts the schema to a ResolvedSchema, panicking on error.
// This is useful in tests and initialization code where schema errors should
// cause a panic rather than require error handling.
func (s *Schema) MustResolve() *ResolvedSchema {
	rs, err := s.Resolve()
	if err != nil {
		panic("schema2: MustResolve: " + err.Error())
	}
	return rs
}

func convertEntityType(et *resolve.EntityType) *ResolvedEntityType {
	ret := &ResolvedEntityType{
		name:        et.Name,
		descendants: et.Descendants,
	}

	if et.Attributes != nil {
		ret.attributes = convertRecordType(et.Attributes)
	}

	if et.Tags != nil {
		ret.tags = convertType(et.Tags)
	}

	switch k := et.Kind.(type) {
	case resolve.StandardKind:
		ret.kind = StandardEntityType{}
	case resolve.EnumKind:
		ret.kind = EnumEntityType{values: k.Values}
	}

	return ret
}

func convertAction(a *resolve.Action) *ResolvedAction {
	ra := &ResolvedAction{
		name:     a.Name,
		memberOf: a.MemberOf,
	}

	if a.AppliesTo != nil {
		ra.appliesTo = &ResolvedAppliesTo{
			principals: a.AppliesTo.Principals,
			resources:  a.AppliesTo.Resources,
		}
	}

	if a.Context != nil {
		ra.context = convertRecordType(a.Context)
	}

	return ra
}

func convertType(t resolve.Type) ResolvedType {
	switch t := t.(type) {
	case resolve.PrimitiveType:
		return ResolvedPrimitiveType{name: t.Name}
	case resolve.EntityRefType:
		return ResolvedEntityRefType{name: t.Name}
	case resolve.SetType:
		return ResolvedSetType{element: convertType(t.Element)}
	case *resolve.RecordType:
		return convertRecordType(t)
	case resolve.ExtensionType:
		return ResolvedExtensionType{name: t.Name}
	default:
		panic("schema2: convertType: unexpected type " + fmt.Sprintf("%T", t))
	}
}

func convertRecordType(rt *resolve.RecordType) *ResolvedRecordType {
	rrt := &ResolvedRecordType{
		attributes: make(map[string]*ResolvedAttribute),
	}
	for name, attr := range rt.Attributes {
		rrt.attributes[name] = &ResolvedAttribute{
			name:     name,
			typ:      convertType(attr.Type),
			required: attr.Required,
		}
	}
	return rrt
}
