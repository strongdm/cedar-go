package schema

import (
	"encoding/json"
	"fmt"
)

func (s *Schema) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return &ParseError{Filename: s.filename, Message: fmt.Sprintf("invalid JSON: %v", err)}
	}
	s.Namespaces = make(map[string]*Namespace)
	for nsName, nsData := range raw {
		ns, err := unmarshalNamespace(nsData, nsName)
		if err != nil {
			return err
		}
		s.Namespaces[nsName] = ns
	}
	return nil
}

func (s *Schema) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Namespaces)
}

func unmarshalNamespace(data json.RawMessage, nsName string) (*Namespace, error) {
	var raw struct {
		EntityTypes map[string]json.RawMessage `json:"entityTypes"`
		Actions     map[string]json.RawMessage `json:"actions"`
		CommonTypes map[string]json.RawMessage `json:"commonTypes"`
		Annotations map[string]string          `json:"annotations"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, &ParseError{Message: fmt.Sprintf("invalid namespace %q: %v", nsName, err)}
	}

	ns := newNamespace()
	if raw.Annotations != nil {
		ns.Annotations = raw.Annotations
	}

	for name, etData := range raw.EntityTypes {
		if isPrimitiveTypeName(name) {
			return nil, &ReservedNameError{Name: name, Kind: "entity type"}
		}
		et, enum, err := unmarshalEntityOrEnumType(etData)
		if err != nil {
			return nil, fmt.Errorf("entity type %q: %w", name, err)
		}
		if enum != nil {
			ns.EnumTypes[name] = enum
		} else {
			ns.EntityTypes[name] = et
		}
	}

	for name, actData := range raw.Actions {
		act, err := unmarshalAction(actData)
		if err != nil {
			return nil, fmt.Errorf("action %q: %w", name, err)
		}
		ns.Actions[name] = act
	}

	for name, ctData := range raw.CommonTypes {
		if isPrimitiveTypeName(name) {
			return nil, &ReservedNameError{Name: name, Kind: "common type"}
		}
		ct, err := unmarshalCommonType(ctData)
		if err != nil {
			return nil, fmt.Errorf("common type %q: %w", name, err)
		}
		ns.CommonTypes[name] = ct
	}

	return ns, nil
}

func unmarshalEntityOrEnumType(data json.RawMessage) (*EntityTypeDef, *EnumTypeDef, error) {
	var raw struct {
		MemberOfTypes []string          `json:"memberOfTypes"`
		Shape         json.RawMessage   `json:"shape"`
		Tags          json.RawMessage   `json:"tags"`
		Enum          []string          `json:"enum"`
		Annotations   map[string]string `json:"annotations"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, nil, err
	}

	if len(raw.Enum) > 0 {
		enum := &EnumTypeDef{Values: raw.Enum, Annotations: raw.Annotations}
		if enum.Annotations == nil {
			enum.Annotations = newAnnotations()
		}
		return nil, enum, nil
	}

	et := &EntityTypeDef{MemberOfTypes: raw.MemberOfTypes, Annotations: raw.Annotations}
	if et.Annotations == nil {
		et.Annotations = newAnnotations()
	}

	if len(raw.Shape) > 0 {
		shape, err := unmarshalTypeExpr(raw.Shape)
		if err != nil {
			return nil, nil, fmt.Errorf("shape: %w", err)
		}
		rt, ok := shape.(*RecordTypeExpr)
		if !ok {
			return nil, nil, fmt.Errorf("shape must be a Record type")
		}
		et.Shape = rt
	}

	if len(raw.Tags) > 0 {
		tags, err := unmarshalTypeExpr(raw.Tags)
		if err != nil {
			return nil, nil, fmt.Errorf("tags: %w", err)
		}
		et.Tags = tags
	}

	return et, nil, nil
}

func unmarshalAction(data json.RawMessage) (*ActionDef, error) {
	var raw struct {
		MemberOf    []actionRefJSON   `json:"memberOf"`
		AppliesTo   *appliesToJSON    `json:"appliesTo"`
		Annotations map[string]string `json:"annotations"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	act := &ActionDef{Annotations: raw.Annotations}
	if act.Annotations == nil {
		act.Annotations = newAnnotations()
	}

	for _, ref := range raw.MemberOf {
		act.MemberOf = append(act.MemberOf, &ActionRef{Type: ref.Type, ID: ref.ID})
	}

	if raw.AppliesTo != nil {
		at := &AppliesTo{
			PrincipalTypes: raw.AppliesTo.PrincipalTypes,
			ResourceTypes:  raw.AppliesTo.ResourceTypes,
		}
		if len(raw.AppliesTo.Context) > 0 {
			ctx, err := unmarshalTypeExpr(raw.AppliesTo.Context)
			if err != nil {
				return nil, fmt.Errorf("context: %w", err)
			}
			at.Context = ctx
		}
		act.AppliesTo = at
	}

	return act, nil
}

type actionRefJSON struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type appliesToJSON struct {
	PrincipalTypes []string        `json:"principalTypes"`
	ResourceTypes  []string        `json:"resourceTypes"`
	Context        json.RawMessage `json:"context"`
}

func unmarshalCommonType(data json.RawMessage) (*CommonTypeDef, error) {
	typ, err := unmarshalTypeExpr(data)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Annotations map[string]string `json:"annotations"`
	}
	_ = json.Unmarshal(data, &raw)

	ct := &CommonTypeDef{Type: typ, Annotations: raw.Annotations}
	if ct.Annotations == nil {
		ct.Annotations = newAnnotations()
	}
	return ct, nil
}

func unmarshalTypeExpr(data json.RawMessage) (TypeExpr, error) {
	var raw struct {
		Type       string                     `json:"type"`
		Name       string                     `json:"name"`
		Element    json.RawMessage            `json:"element"`
		Attributes map[string]json.RawMessage `json:"attributes"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	switch raw.Type {
	case "Long":
		return PrimitiveTypeExpr{Kind: PrimitiveLong}, nil
	case "String":
		return PrimitiveTypeExpr{Kind: PrimitiveString}, nil
	case "Bool", "Boolean":
		return PrimitiveTypeExpr{Kind: PrimitiveBool}, nil
	case "Set":
		if len(raw.Element) == 0 {
			return nil, fmt.Errorf("Set type requires element")
		}
		elem, err := unmarshalTypeExpr(raw.Element)
		if err != nil {
			return nil, fmt.Errorf("Set element: %w", err)
		}
		return SetTypeExpr{Element: elem}, nil
	case "Record":
		return parseRecordTypeExpr(raw.Attributes)
	case "Entity":
		if raw.Name == "" {
			return nil, fmt.Errorf("Entity type requires name")
		}
		return EntityRefExpr{Name: raw.Name}, nil
	case "Extension":
		if raw.Name == "" {
			return nil, fmt.Errorf("Extension type requires name")
		}
		return ExtensionTypeExpr{Name: raw.Name}, nil
	case "EntityOrCommon":
		if raw.Name == "" {
			return nil, fmt.Errorf("EntityOrCommon type requires name")
		}
		return TypeNameExpr{Name: raw.Name}, nil
	case "":
		if raw.Name != "" {
			return TypeNameExpr{Name: raw.Name}, nil
		}
		if len(raw.Attributes) > 0 {
			return parseRecordTypeExpr(raw.Attributes)
		}
		return nil, fmt.Errorf("unknown type format")
	default:
		return TypeNameExpr{Name: raw.Type}, nil
	}
}

func parseRecordTypeExpr(attrs map[string]json.RawMessage) (*RecordTypeExpr, error) {
	rt := &RecordTypeExpr{Attributes: make(map[string]*Attribute)}
	for attrName, attrData := range attrs {
		attr, err := unmarshalAttribute(attrData)
		if err != nil {
			return nil, fmt.Errorf("attribute %q: %w", attrName, err)
		}
		rt.Attributes[attrName] = attr
	}
	return rt, nil
}

func unmarshalAttribute(data json.RawMessage) (*Attribute, error) {
	var raw struct {
		Required    *bool             `json:"required"`
		Annotations map[string]string `json:"annotations"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	typ, err := unmarshalTypeExpr(data)
	if err != nil {
		return nil, err
	}

	attr := &Attribute{Type: typ, Required: true, Annotations: raw.Annotations}
	if attr.Annotations == nil {
		attr.Annotations = newAnnotations()
	}
	if raw.Required != nil {
		attr.Required = *raw.Required
	}
	return attr, nil
}

// Marshal helpers

func (ns *Namespace) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if len(ns.EntityTypes) > 0 || len(ns.EnumTypes) > 0 {
		entityTypes := make(map[string]any)
		for name, et := range ns.EntityTypes {
			entityTypes[name] = et
		}
		for name, enum := range ns.EnumTypes {
			entityTypes[name] = enum
		}
		m["entityTypes"] = entityTypes
	}
	if len(ns.Actions) > 0 {
		m["actions"] = ns.Actions
	}
	if len(ns.CommonTypes) > 0 {
		m["commonTypes"] = ns.CommonTypes
	}
	if len(ns.Annotations) > 0 {
		m["annotations"] = ns.Annotations
	}
	return json.Marshal(m)
}

func (et *EntityTypeDef) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if len(et.MemberOfTypes) > 0 {
		m["memberOfTypes"] = et.MemberOfTypes
	}
	if et.Shape != nil {
		m["shape"] = et.Shape
	}
	if et.Tags != nil {
		m["tags"] = marshalTypeExprValue(et.Tags)
	}
	if len(et.Annotations) > 0 {
		m["annotations"] = et.Annotations
	}
	return json.Marshal(m)
}

func (enum *EnumTypeDef) MarshalJSON() ([]byte, error) {
	m := map[string]any{"enum": enum.Values}
	if len(enum.Annotations) > 0 {
		m["annotations"] = enum.Annotations
	}
	return json.Marshal(m)
}

func (a *ActionDef) MarshalJSON() ([]byte, error) {
	m := make(map[string]any)
	if len(a.MemberOf) > 0 {
		refs := make([]actionRefJSON, len(a.MemberOf))
		for i, ref := range a.MemberOf {
			refs[i] = actionRefJSON{Type: ref.Type, ID: ref.ID}
		}
		m["memberOf"] = refs
	}
	if a.AppliesTo != nil {
		at := make(map[string]any)
		if len(a.AppliesTo.PrincipalTypes) > 0 {
			at["principalTypes"] = a.AppliesTo.PrincipalTypes
		}
		if len(a.AppliesTo.ResourceTypes) > 0 {
			at["resourceTypes"] = a.AppliesTo.ResourceTypes
		}
		if a.AppliesTo.Context != nil {
			at["context"] = marshalTypeExprValue(a.AppliesTo.Context)
		}
		m["appliesTo"] = at
	}
	if len(a.Annotations) > 0 {
		m["annotations"] = a.Annotations
	}
	return json.Marshal(m)
}

func (ct *CommonTypeDef) MarshalJSON() ([]byte, error) {
	if len(ct.Annotations) == 0 {
		return json.Marshal(marshalTypeExprValue(ct.Type))
	}
	m := marshalTypeExprToMap(ct.Type)
	m["annotations"] = ct.Annotations
	return json.Marshal(m)
}

func (rt *RecordTypeExpr) MarshalJSON() ([]byte, error) {
	return json.Marshal(marshalRecordTypeExpr(rt))
}

func marshalRecordTypeExpr(rt *RecordTypeExpr) map[string]any {
	attrs := make(map[string]any)
	for name, attr := range rt.Attributes {
		attrMap := marshalTypeExprToMap(attr.Type)
		if !attr.Required {
			attrMap["required"] = false
		}
		if len(attr.Annotations) > 0 {
			attrMap["annotations"] = attr.Annotations
		}
		attrs[name] = attrMap
	}
	return map[string]any{"type": "Record", "attributes": attrs}
}

func marshalTypeExprValue(t TypeExpr) any {
	switch v := t.(type) {
	case PrimitiveTypeExpr:
		return map[string]string{"type": v.Kind.String()}
	case SetTypeExpr:
		return map[string]any{"type": "Set", "element": marshalTypeExprValue(v.Element)}
	case *RecordTypeExpr:
		return marshalRecordTypeExpr(v)
	case EntityRefExpr:
		return map[string]string{"type": "Entity", "name": v.Name}
	case ExtensionTypeExpr:
		return map[string]string{"type": "Extension", "name": v.Name}
	case TypeNameExpr:
		return map[string]string{"type": v.Name}
	default: // EntityNameExpr
		return map[string]string{"type": "Entity", "name": t.(EntityNameExpr).Name}
	}
}

func marshalTypeExprToMap(t TypeExpr) map[string]any {
	switch v := t.(type) {
	case PrimitiveTypeExpr:
		return map[string]any{"type": v.Kind.String()}
	case SetTypeExpr:
		return map[string]any{"type": "Set", "element": marshalTypeExprValue(v.Element)}
	case *RecordTypeExpr:
		return marshalRecordTypeExpr(v)
	case EntityRefExpr:
		return map[string]any{"type": "Entity", "name": v.Name}
	case ExtensionTypeExpr:
		return map[string]any{"type": "Extension", "name": v.Name}
	case TypeNameExpr:
		return map[string]any{"type": v.Name}
	default: // EntityNameExpr
		return map[string]any{"type": "Entity", "name": t.(EntityNameExpr).Name}
	}
}
