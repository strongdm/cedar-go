package ast

import (
	"github.com/cedar-policy/cedar-go/internal/schema/ast"
)

// ToJSON converts the internal schema AST to JSON schema format.
func (s *Schema) ToJSON() ast.JSONSchema {
	js := make(ast.JSONSchema)
	for name, ns := range s.Namespaces {
		js[name] = ns.toJSON()
	}
	return js
}

func (ns *Namespace) toJSON() *ast.JSONNamespace {
	jns := &ast.JSONNamespace{
		EntityTypes: make(map[string]*ast.JSONEntity),
		Actions:     make(map[string]*ast.JSONAction),
		CommonTypes: make(map[string]*ast.JSONCommonType),
	}

	for name, et := range ns.Entities {
		jns.EntityTypes[name] = et.toJSON()
	}

	for name, a := range ns.Actions {
		jns.Actions[name] = a.toJSON()
	}

	for name, t := range ns.CommonTypes {
		jns.CommonTypes[name] = &ast.JSONCommonType{
			JSONType: typeToJSON(t),
		}
	}

	return jns
}

func (et *EntityType) toJSON() *ast.JSONEntity {
	je := &ast.JSONEntity{}

	if len(et.MemberOf) > 0 {
		je.MemberOfTypes = et.MemberOf
	}

	if et.Shape != nil {
		je.Shape = typeToJSON(et.Shape)
	}

	if et.Tags != nil {
		je.Tags = typeToJSON(et.Tags)
	}

	if len(et.Enum) > 0 {
		je.Enum = et.Enum
	}

	return je
}

func (a *Action) toJSON() *ast.JSONAction {
	ja := &ast.JSONAction{}

	if len(a.MemberOf) > 0 {
		ja.MemberOf = make([]*ast.JSONMember, len(a.MemberOf))
		for i, ref := range a.MemberOf {
			ja.MemberOf[i] = &ast.JSONMember{
				ID:   ref.Name,
				Type: ref.Namespace,
			}
		}
	}

	if a.AppliesTo != nil {
		ja.AppliesTo = &ast.JSONAppliesTo{
			PrincipalTypes: a.AppliesTo.Principals,
			ResourceTypes:  a.AppliesTo.Resources,
		}
		if a.AppliesTo.Context != nil {
			ja.AppliesTo.Context = typeToJSON(a.AppliesTo.Context)
		}
	}

	return ja
}

func typeToJSON(t Type) *ast.JSONType {
	switch t := t.(type) {
	case PrimitiveType:
		return &ast.JSONType{Type: t.Name}
	case EntityRefType:
		return &ast.JSONType{Type: "EntityOrCommon", Name: t.Name}
	case *SetType:
		return &ast.JSONType{Type: "Set", Element: typeToJSON(t.Element)}
	case SetType:
		return &ast.JSONType{Type: "Set", Element: typeToJSON(t.Element)}
	case *RecordType:
		return recordTypeToJSON(t)
	case RecordType:
		return recordTypeToJSON(&t)
	case CommonTypeRef:
		return &ast.JSONType{Type: "EntityOrCommon", Name: t.Name}
	case ExtensionType:
		return &ast.JSONType{Type: "Extension", Name: t.Name}
	default:
		return &ast.JSONType{Type: "Record"}
	}
}

func recordTypeToJSON(rt *RecordType) *ast.JSONType {
	jt := &ast.JSONType{
		Type:       "Record",
		Attributes: make(map[string]*ast.JSONAttribute),
	}

	if rt.Attributes != nil {
		for name, attr := range rt.Attributes {
			inner := typeToJSON(attr.Type)
			jt.Attributes[name] = &ast.JSONAttribute{
				Type:       inner.Type,
				Required:   attr.Required,
				Element:    inner.Element,
				Name:       inner.Name,
				Attributes: inner.Attributes,
			}
		}
	}

	return jt
}

// FromJSON converts a JSON schema to the internal schema AST.
func FromJSON(js ast.JSONSchema) *Schema {
	s := NewSchema()
	for name, jns := range js {
		s.Namespaces[name] = namespaceFromJSON(name, jns)
	}
	return s
}

func namespaceFromJSON(name string, jns *ast.JSONNamespace) *Namespace {
	ns := &Namespace{
		Name:        name,
		Entities:    make(map[string]*EntityType),
		Actions:     make(map[string]*Action),
		CommonTypes: make(map[string]Type),
	}

	for ename, je := range jns.EntityTypes {
		ns.Entities[ename] = entityFromJSON(ename, je)
	}

	for aname, ja := range jns.Actions {
		ns.Actions[aname] = actionFromJSON(aname, ja)
	}

	for tname, jt := range jns.CommonTypes {
		ns.CommonTypes[tname] = typeFromJSON(jt.JSONType)
	}

	return ns
}

func entityFromJSON(name string, je *ast.JSONEntity) *EntityType {
	et := &EntityType{
		Name:     name,
		MemberOf: je.MemberOfTypes,
		Enum:     je.Enum,
	}

	if je.Shape != nil {
		if rt, ok := typeFromJSON(je.Shape).(*RecordType); ok {
			et.Shape = rt
		}
	}

	if je.Tags != nil {
		et.Tags = typeFromJSON(je.Tags)
	}

	return et
}

func actionFromJSON(name string, ja *ast.JSONAction) *Action {
	a := &Action{
		Name: name,
	}

	if len(ja.MemberOf) > 0 {
		a.MemberOf = make([]ActionRef, len(ja.MemberOf))
		for i, jm := range ja.MemberOf {
			a.MemberOf[i] = ActionRef{
				Namespace: jm.Type,
				Name:      jm.ID,
			}
		}
	}

	if ja.AppliesTo != nil {
		a.AppliesTo = &AppliesTo{
			Principals: ja.AppliesTo.PrincipalTypes,
			Resources:  ja.AppliesTo.ResourceTypes,
		}
		if ja.AppliesTo.Context != nil {
			a.AppliesTo.Context = typeFromJSON(ja.AppliesTo.Context)
		}
	}

	return a
}

func typeFromJSON(jt *ast.JSONType) Type {
	if jt == nil {
		return nil
	}

	switch jt.Type {
	case "String":
		return PrimitiveType{Name: "String"}
	case "Long":
		return PrimitiveType{Name: "Long"}
	case "Boolean":
		return PrimitiveType{Name: "Bool"}
	case "Bool":
		return PrimitiveType{Name: "Bool"}
	case "Set":
		return &SetType{Element: typeFromJSON(jt.Element)}
	case "Record":
		return recordTypeFromJSON(jt)
	case "EntityOrCommon":
		// Could be entity or common type - we treat as reference
		return EntityRefType{Name: jt.Name}
	case "Extension":
		return ExtensionType{Name: jt.Name}
	default:
		// Unknown type, treat as entity/common reference
		if jt.Name != "" {
			return EntityRefType{Name: jt.Name}
		}
		return &RecordType{Attributes: make(map[string]*Attribute)}
	}
}

func recordTypeFromJSON(jt *ast.JSONType) *RecordType {
	rt := &RecordType{
		Attributes: make(map[string]*Attribute),
	}

	for name, ja := range jt.Attributes {
		inner := &ast.JSONType{
			Type:       ja.Type,
			Element:    ja.Element,
			Name:       ja.Name,
			Attributes: ja.Attributes,
		}
		rt.Attributes[name] = &Attribute{
			Name:     name,
			Type:     typeFromJSON(inner),
			Required: ja.Required,
		}
	}

	return rt
}
