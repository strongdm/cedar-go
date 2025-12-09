package schema

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
)

// MarshalJSON serializes the schema to JSON format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	js := make(jsonSchema)
	for name, ns := range s.Namespaces {
		jsNS, err := namespaceToJSON(ns)
		if err != nil {
			return nil, err
		}
		js[string(name)] = jsNS
	}
	return json.Marshal(js)
}

// UnmarshalJSON deserializes a schema from JSON format.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var js jsonSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return err
	}

	result := New()
	for name, jsNS := range js {
		ns, err := namespaceFromJSON(types.Path(name), jsNS)
		if err != nil {
			return err
		}
		result.Namespaces[types.Path(name)] = ns
	}
	*s = *result
	return nil
}

// JSON schema types

type jsonSchema map[string]*jsonNamespace

type jsonNamespace struct {
	EntityTypes map[string]*jsonEntity     `json:"entityTypes,omitempty"`
	Actions     map[string]*jsonAction     `json:"actions,omitempty"`
	CommonTypes map[string]*jsonCommonType `json:"commonTypes,omitempty"`
	Annotations map[string]string          `json:"annotations,omitempty"`
}

type jsonEntity struct {
	MemberOfTypes []string          `json:"memberOfTypes,omitempty"`
	Shape         *jsonType         `json:"shape,omitempty"`
	Tags          *jsonType         `json:"tags,omitempty"`
	Enum          []string          `json:"enum,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type jsonAction struct {
	MemberOf    []*jsonMember     `json:"memberOf,omitempty"`
	AppliesTo   *jsonAppliesTo    `json:"appliesTo,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type jsonMember struct {
	ID   string `json:"id"`
	Type string `json:"type,omitempty"`
}

type jsonAppliesTo struct {
	PrincipalTypes []string  `json:"principalTypes,omitempty"`
	ResourceTypes  []string  `json:"resourceTypes,omitempty"`
	Context        *jsonType `json:"context,omitempty"`
}

type jsonType struct {
	Type       string                  `json:"type"`
	Element    *jsonType               `json:"element,omitempty"`
	Name       string                  `json:"name,omitempty"`
	Attributes map[string]*jsonAttr    `json:"attributes,omitempty"`
}

type jsonAttr struct {
	Type       string                  `json:"type"`
	Required   bool                    `json:"required,omitempty"`
	Element    *jsonType               `json:"element,omitempty"`
	Name       string                  `json:"name,omitempty"`
	Attributes map[string]*jsonAttr    `json:"attributes,omitempty"`
}

type jsonCommonType struct {
	Type       string                  `json:"type"`
	Element    *jsonType               `json:"element,omitempty"`
	Name       string                  `json:"name,omitempty"`
	Attributes map[string]*jsonAttr    `json:"attributes,omitempty"`
}

// Conversion functions: Schema -> JSON

func namespaceToJSON(ns *Namespace) (*jsonNamespace, error) {
	jsNS := &jsonNamespace{
		EntityTypes: make(map[string]*jsonEntity),
		Actions:     make(map[string]*jsonAction),
		CommonTypes: make(map[string]*jsonCommonType),
		Annotations: make(map[string]string),
	}

	for _, ann := range ns.Annotations {
		jsNS.Annotations[string(ann.Key)] = string(ann.Value)
	}

	for name, entity := range ns.Entities {
		jsEntity, err := entityToJSON(entity)
		if err != nil {
			return nil, err
		}
		jsNS.EntityTypes[string(name)] = jsEntity
	}

	for name, action := range ns.Actions {
		jsAction, err := actionToJSON(action)
		if err != nil {
			return nil, err
		}
		jsNS.Actions[string(name)] = jsAction
	}

	for name, ct := range ns.CommonTypes {
		jsCT, err := commonTypeToJSON(ct)
		if err != nil {
			return nil, err
		}
		jsNS.CommonTypes[string(name)] = jsCT
	}

	return jsNS, nil
}

func entityToJSON(e *EntityDecl) (*jsonEntity, error) {
	jsEntity := &jsonEntity{
		Annotations: make(map[string]string),
	}

	for _, ann := range e.Annotations {
		jsEntity.Annotations[string(ann.Key)] = string(ann.Value)
	}

	for _, parent := range e.MemberOfTypes {
		jsEntity.MemberOfTypes = append(jsEntity.MemberOfTypes, string(parent))
	}

	if len(e.Attributes) > 0 {
		jsType, err := typeToJSON(Record(e.Attributes...))
		if err != nil {
			return nil, err
		}
		jsEntity.Shape = jsType
	}

	if e.Tags.v != nil {
		jsType, err := typeToJSON(e.Tags)
		if err != nil {
			return nil, err
		}
		jsEntity.Tags = jsType
	}

	for _, enum := range e.Enum {
		jsEntity.Enum = append(jsEntity.Enum, string(enum))
	}

	return jsEntity, nil
}

func actionToJSON(a *ActionDecl) (*jsonAction, error) {
	jsAction := &jsonAction{
		Annotations: make(map[string]string),
	}

	for _, ann := range a.Annotations {
		jsAction.Annotations[string(ann.Key)] = string(ann.Value)
	}

	for _, ref := range a.MemberOf {
		jsMember := &jsonMember{
			ID: string(ref.Name),
		}
		if ref.Namespace != "" {
			jsMember.Type = string(ref.Namespace)
		}
		jsAction.MemberOf = append(jsAction.MemberOf, jsMember)
	}

	if len(a.PrincipalTypes) > 0 || len(a.ResourceTypes) > 0 || a.Context.v != nil {
		jsAction.AppliesTo = &jsonAppliesTo{}

		for _, p := range a.PrincipalTypes {
			jsAction.AppliesTo.PrincipalTypes = append(jsAction.AppliesTo.PrincipalTypes, string(p))
		}

		for _, r := range a.ResourceTypes {
			jsAction.AppliesTo.ResourceTypes = append(jsAction.AppliesTo.ResourceTypes, string(r))
		}

		if a.Context.v != nil {
			jsType, err := typeToJSON(a.Context)
			if err != nil {
				return nil, err
			}
			jsAction.AppliesTo.Context = jsType
		}
	}

	return jsAction, nil
}

func commonTypeToJSON(c *CommonTypeDecl) (*jsonCommonType, error) {
	jsType, err := typeToJSON(c.Type)
	if err != nil {
		return nil, err
	}
	if jsType == nil {
		return &jsonCommonType{}, nil
	}
	return &jsonCommonType{
		Type:       jsType.Type,
		Element:    jsType.Element,
		Name:       jsType.Name,
		Attributes: jsType.Attributes,
	}, nil
}

func typeToJSON(t Type) (*jsonType, error) {
	if t.v == nil {
		return nil, nil
	}

	switch v := t.v.(type) {
	case TypeBoolean:
		return &jsonType{Type: "Boolean"}, nil
	case TypeLong:
		return &jsonType{Type: "Long"}, nil
	case TypeString:
		return &jsonType{Type: "String"}, nil
	case TypeSet:
		elem, err := typeToJSON(v.Element)
		if err != nil {
			return nil, err
		}
		return &jsonType{Type: "Set", Element: elem}, nil
	case TypeRecord:
		attrs := make(map[string]*jsonAttr)
		for _, attr := range v.Attributes {
			jsAttr, err := attrToJSON(attr)
			if err != nil {
				return nil, err
			}
			attrs[string(attr.Name)] = jsAttr
		}
		return &jsonType{Type: "Record", Attributes: attrs}, nil
	case TypeEntity:
		return &jsonType{Type: "EntityOrCommon", Name: string(v.Name)}, nil
	case TypeExtension:
		return &jsonType{Type: "Extension", Name: string(v.Name)}, nil
	case TypeRef:
		return &jsonType{Type: "EntityOrCommon", Name: string(v.Name)}, nil
	default:
		return nil, fmt.Errorf("unknown type: %T", v)
	}
}

func attrToJSON(a Attribute) (*jsonAttr, error) {
	jsAttr := &jsonAttr{
		Required: a.Required,
	}

	if a.Type.v == nil {
		return jsAttr, nil
	}

	switch v := a.Type.v.(type) {
	case TypeBoolean:
		jsAttr.Type = "Boolean"
	case TypeLong:
		jsAttr.Type = "Long"
	case TypeString:
		jsAttr.Type = "String"
	case TypeSet:
		jsAttr.Type = "Set"
		elem, err := typeToJSON(v.Element)
		if err != nil {
			return nil, err
		}
		jsAttr.Element = elem
	case TypeRecord:
		jsAttr.Type = "Record"
		jsAttr.Attributes = make(map[string]*jsonAttr)
		for _, attr := range v.Attributes {
			nested, err := attrToJSON(attr)
			if err != nil {
				return nil, err
			}
			jsAttr.Attributes[string(attr.Name)] = nested
		}
	case TypeEntity:
		jsAttr.Type = "EntityOrCommon"
		jsAttr.Name = string(v.Name)
	case TypeExtension:
		jsAttr.Type = "Extension"
		jsAttr.Name = string(v.Name)
	case TypeRef:
		jsAttr.Type = "EntityOrCommon"
		jsAttr.Name = string(v.Name)
	default:
		return nil, fmt.Errorf("unknown type: %T", v)
	}

	return jsAttr, nil
}

// Conversion functions: JSON -> Schema

func namespaceFromJSON(name types.Path, jsNS *jsonNamespace) (*Namespace, error) {
	ns := NewNamespace(name)

	for key, value := range jsNS.Annotations {
		ns.Annotations = ns.Annotations.Set(types.Ident(key), types.String(value))
	}

	// Sort entity names for deterministic iteration
	entityNames := make([]string, 0, len(jsNS.EntityTypes))
	for name := range jsNS.EntityTypes {
		entityNames = append(entityNames, name)
	}
	sort.Strings(entityNames)

	for _, entityName := range entityNames {
		jsEntity := jsNS.EntityTypes[entityName]
		entity, err := entityFromJSON(types.Ident(entityName), jsEntity)
		if err != nil {
			return nil, err
		}
		ns.Entities[types.Ident(entityName)] = entity
	}

	// Sort action names for deterministic iteration
	actionNames := make([]string, 0, len(jsNS.Actions))
	for name := range jsNS.Actions {
		actionNames = append(actionNames, name)
	}
	sort.Strings(actionNames)

	for _, actionName := range actionNames {
		jsAction := jsNS.Actions[actionName]
		action, err := actionFromJSON(types.String(actionName), jsAction)
		if err != nil {
			return nil, err
		}
		ns.Actions[types.String(actionName)] = action
	}

	// Sort common type names for deterministic iteration
	ctNames := make([]string, 0, len(jsNS.CommonTypes))
	for name := range jsNS.CommonTypes {
		ctNames = append(ctNames, name)
	}
	sort.Strings(ctNames)

	for _, ctName := range ctNames {
		jsCT := jsNS.CommonTypes[ctName]
		ct, err := commonTypeFromJSON(types.Ident(ctName), jsCT)
		if err != nil {
			return nil, err
		}
		ns.CommonTypes[types.Ident(ctName)] = ct
	}

	return ns, nil
}

func entityFromJSON(name types.Ident, jsEntity *jsonEntity) (*EntityDecl, error) {
	entity := NewEntity(name)

	for key, value := range jsEntity.Annotations {
		entity.Annotations = entity.Annotations.Set(types.Ident(key), types.String(value))
	}

	for _, parent := range jsEntity.MemberOfTypes {
		entity.MemberOfTypes = append(entity.MemberOfTypes, types.Path(parent))
	}

	if jsEntity.Shape != nil {
		t, err := typeFromJSON(jsEntity.Shape)
		if err != nil {
			return nil, err
		}
		if rt, ok := t.v.(TypeRecord); ok {
			entity.Attributes = rt.Attributes
		}
	}

	if jsEntity.Tags != nil {
		t, err := typeFromJSON(jsEntity.Tags)
		if err != nil {
			return nil, err
		}
		entity.Tags = t
	}

	for _, enumVal := range jsEntity.Enum {
		entity.Enum = append(entity.Enum, types.String(enumVal))
	}

	return entity, nil
}

func actionFromJSON(name types.String, jsAction *jsonAction) (*ActionDecl, error) {
	action := NewAction(name)

	for key, value := range jsAction.Annotations {
		action.Annotations = action.Annotations.Set(types.Ident(key), types.String(value))
	}

	for _, jsMember := range jsAction.MemberOf {
		ref := ActionRef{
			Name: types.String(jsMember.ID),
		}
		if jsMember.Type != "" {
			ref.Namespace = types.Path(jsMember.Type)
		}
		action.MemberOf = append(action.MemberOf, ref)
	}

	if jsAction.AppliesTo != nil {
		for _, p := range jsAction.AppliesTo.PrincipalTypes {
			action.PrincipalTypes = append(action.PrincipalTypes, types.Path(p))
		}

		for _, r := range jsAction.AppliesTo.ResourceTypes {
			action.ResourceTypes = append(action.ResourceTypes, types.Path(r))
		}

		if jsAction.AppliesTo.Context != nil {
			t, err := typeFromJSON(jsAction.AppliesTo.Context)
			if err != nil {
				return nil, err
			}
			action.Context = t
		}
	}

	return action, nil
}

func commonTypeFromJSON(name types.Ident, jsCT *jsonCommonType) (*CommonTypeDecl, error) {
	jsType := &jsonType{
		Type:       jsCT.Type,
		Element:    jsCT.Element,
		Name:       jsCT.Name,
		Attributes: jsCT.Attributes,
	}
	t, err := typeFromJSON(jsType)
	if err != nil {
		return nil, err
	}
	return NewCommonType(name, t), nil
}

func typeFromJSON(jsType *jsonType) (Type, error) {
	if jsType == nil {
		return Type{}, nil
	}

	switch jsType.Type {
	case "Boolean":
		return Boolean(), nil
	case "Long":
		return Long(), nil
	case "String":
		return String(), nil
	case "Set":
		elem, err := typeFromJSON(jsType.Element)
		if err != nil {
			return Type{}, err
		}
		return SetOf(elem), nil
	case "Record":
		var attrs []Attribute
		// Sort attribute names for deterministic iteration
		attrNames := make([]string, 0, len(jsType.Attributes))
		for name := range jsType.Attributes {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)

		for _, name := range attrNames {
			jsAttr := jsType.Attributes[name]
			attr, err := attrFromJSON(types.Ident(name), jsAttr)
			if err != nil {
				return Type{}, err
			}
			attrs = append(attrs, attr)
		}
		return Record(attrs...), nil
	case "EntityOrCommon":
		// EntityOrCommon could be an entity reference or a common type reference
		// We treat it as a type reference which will be resolved later
		return Ref(types.Path(jsType.Name)), nil
	case "Extension":
		return Extension(types.Path(jsType.Name)), nil
	default:
		return Type{}, fmt.Errorf("unknown JSON type: %s", jsType.Type)
	}
}

func attrFromJSON(name types.Ident, jsAttr *jsonAttr) (Attribute, error) {
	var t Type
	var err error

	switch jsAttr.Type {
	case "Boolean":
		t = Boolean()
	case "Long":
		t = Long()
	case "String":
		t = String()
	case "Set":
		elem, err := typeFromJSON(jsAttr.Element)
		if err != nil {
			return Attribute{}, err
		}
		t = SetOf(elem)
	case "Record":
		var attrs []Attribute
		// Sort attribute names for deterministic iteration
		attrNames := make([]string, 0, len(jsAttr.Attributes))
		for n := range jsAttr.Attributes {
			attrNames = append(attrNames, n)
		}
		sort.Strings(attrNames)

		for _, n := range attrNames {
			nestedAttr := jsAttr.Attributes[n]
			attr, err := attrFromJSON(types.Ident(n), nestedAttr)
			if err != nil {
				return Attribute{}, err
			}
			attrs = append(attrs, attr)
		}
		t = Record(attrs...)
	case "EntityOrCommon":
		t = Ref(types.Path(jsAttr.Name))
	case "Extension":
		t = Extension(types.Path(jsAttr.Name))
	default:
		return Attribute{}, fmt.Errorf("unknown JSON attribute type: %s", jsAttr.Type)
	}

	if err != nil {
		return Attribute{}, err
	}

	return Attribute{
		Name:     name,
		Type:     t,
		Required: jsAttr.Required,
	}, nil
}
