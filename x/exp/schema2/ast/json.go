package ast

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
)

// MarshalJSON serializes the schema to Cedar JSON schema format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	// First pass: collect all entity names for type resolution
	entityNames := collectEntityNames(s)

	namespaces := make(map[string]*jsonNamespace)

	// Second pass: group nodes by namespace
	for _, node := range s.Nodes {
		switch n := node.(type) {
		case *NamespaceNode:
			ns := getOrCreateNamespace(namespaces, string(n.Name))
			for _, decl := range n.Declarations {
				addDeclToNamespace(ns, decl, entityNames)
			}
		case *EntityNode:
			ns := getOrCreateNamespace(namespaces, "")
			addEntityToNamespace(ns, n, entityNames)
		case *EnumNode:
			ns := getOrCreateNamespace(namespaces, "")
			addEnumToNamespace(ns, n)
		case *ActionNode:
			ns := getOrCreateNamespace(namespaces, "")
			addActionToNamespace(ns, n, entityNames)
		case *CommonTypeNode:
			ns := getOrCreateNamespace(namespaces, "")
			addCommonTypeToNamespace(ns, n, entityNames)
		}
	}

	return json.MarshalIndent(namespaces, "", "    ")
}

// collectEntityNames collects all entity names from the schema.
func collectEntityNames(s *Schema) map[string]bool {
	names := make(map[string]bool)
	for _, node := range s.Nodes {
		switch n := node.(type) {
		case *NamespaceNode:
			for _, decl := range n.Declarations {
				switch d := decl.(type) {
				case *EntityNode:
					names[string(d.Name)] = true
				case *EnumNode:
					names[string(d.Name)] = true
				}
			}
		case *EntityNode:
			names[string(n.Name)] = true
		case *EnumNode:
			names[string(n.Name)] = true
		}
	}
	return names
}

// UnmarshalJSON deserializes a Cedar JSON schema into an AST.
func (s *Schema) UnmarshalJSON(data []byte) error {
	namespaces := make(map[string]*jsonNamespace)
	if err := json.Unmarshal(data, &namespaces); err != nil {
		return err
	}

	s.Nodes = nil

	// Sort namespace names for deterministic output
	nsNames := make([]string, 0, len(namespaces))
	for name := range namespaces {
		nsNames = append(nsNames, name)
	}
	sort.Strings(nsNames)

	for _, nsName := range nsNames {
		ns := namespaces[nsName]
		if nsName == "" {
			// Default namespace - add declarations at top level
			nodes, err := parseNamespaceContents(ns)
			if err != nil {
				return err
			}
			s.Nodes = append(s.Nodes, nodes...)
		} else {
			// Named namespace
			decls, err := parseNamespaceContents(ns)
			if err != nil {
				return err
			}
			declarations := make([]IsDeclaration, 0, len(decls))
			for _, d := range decls {
				if decl, ok := d.(IsDeclaration); ok {
					declarations = append(declarations, decl)
				}
			}
			s.Nodes = append(s.Nodes, &NamespaceNode{
				Name:         types.Path(nsName),
				Declarations: declarations,
			})
		}
	}

	return nil
}

// JSON schema types

type jsonNamespace struct {
	CommonTypes map[string]*jsonType   `json:"commonTypes,omitempty"`
	EntityTypes map[string]*jsonEntity `json:"entityTypes"`
	Actions     map[string]*jsonAction `json:"actions"`
}

type jsonEntity struct {
	MemberOfTypes []string  `json:"memberOfTypes,omitempty"`
	Shape         *jsonType `json:"shape,omitempty"`
	Tags          *jsonType `json:"tags,omitempty"`
	Enum          []string  `json:"enum,omitempty"`
}

type jsonAction struct {
	MemberOf  []jsonEntityUID `json:"memberOf,omitempty"`
	AppliesTo *jsonAppliesTo  `json:"appliesTo,omitempty"`
}

type jsonEntityUID struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type jsonAppliesTo struct {
	PrincipalTypes []string  `json:"principalTypes,omitempty"`
	ResourceTypes  []string  `json:"resourceTypes,omitempty"`
	Context        *jsonType `json:"context,omitempty"`
}

type jsonType struct {
	Type       string               `json:"type,omitempty"`
	Name       string               `json:"name,omitempty"`
	Element    *jsonType            `json:"element,omitempty"`
	Attributes map[string]*jsonAttr `json:"attributes,omitempty"`
}

type jsonAttr struct {
	Type       string               `json:"type,omitempty"`
	Name       string               `json:"name,omitempty"`
	Element    *jsonType            `json:"element,omitempty"`
	Attributes map[string]*jsonAttr `json:"attributes,omitempty"`
	Required   *bool                `json:"required,omitempty"`
}

// Helper functions for marshalling

func getOrCreateNamespace(namespaces map[string]*jsonNamespace, name string) *jsonNamespace {
	if ns, ok := namespaces[name]; ok {
		return ns
	}
	ns := &jsonNamespace{
		CommonTypes: make(map[string]*jsonType),
		EntityTypes: make(map[string]*jsonEntity),
		Actions:     make(map[string]*jsonAction),
	}
	namespaces[name] = ns
	return ns
}

func addDeclToNamespace(ns *jsonNamespace, decl IsDeclaration, entityNames map[string]bool) {
	switch d := decl.(type) {
	case *EntityNode:
		addEntityToNamespace(ns, d, entityNames)
	case *EnumNode:
		addEnumToNamespace(ns, d)
	case *ActionNode:
		addActionToNamespace(ns, d, entityNames)
	case *CommonTypeNode:
		addCommonTypeToNamespace(ns, d, entityNames)
	}
}

func addEntityToNamespace(ns *jsonNamespace, e *EntityNode, entityNames map[string]bool) {
	je := &jsonEntity{}

	if len(e.MemberOfVal) > 0 {
		je.MemberOfTypes = make([]string, len(e.MemberOfVal))
		for i, ref := range e.MemberOfVal {
			je.MemberOfTypes[i] = string(ref.Name)
		}
	}

	if e.ShapeVal != nil && len(e.ShapeVal.Pairs) > 0 {
		je.Shape = typeToJSON(*e.ShapeVal, entityNames)
	}

	if e.TagsVal != nil {
		je.Tags = typeToJSON(e.TagsVal, entityNames)
	}

	ns.EntityTypes[string(e.Name)] = je
}

func addEnumToNamespace(ns *jsonNamespace, e *EnumNode) {
	je := &jsonEntity{
		Enum: make([]string, len(e.Values)),
	}
	for i, v := range e.Values {
		je.Enum[i] = string(v)
	}
	ns.EntityTypes[string(e.Name)] = je
}

func addActionToNamespace(ns *jsonNamespace, a *ActionNode, entityNames map[string]bool) {
	ja := &jsonAction{}

	if len(a.MemberOfVal) > 0 {
		ja.MemberOf = make([]jsonEntityUID, len(a.MemberOfVal))
		for i, ref := range a.MemberOfVal {
			ja.MemberOf[i] = jsonEntityUID{
				Type: string(ref.Type.Name),
				ID:   string(ref.ID),
			}
		}
	}

	if a.AppliesToVal != nil {
		ja.AppliesTo = &jsonAppliesTo{}
		if len(a.AppliesToVal.PrincipalTypes) > 0 {
			ja.AppliesTo.PrincipalTypes = make([]string, len(a.AppliesToVal.PrincipalTypes))
			for i, ref := range a.AppliesToVal.PrincipalTypes {
				ja.AppliesTo.PrincipalTypes[i] = string(ref.Name)
			}
		}
		if len(a.AppliesToVal.ResourceTypes) > 0 {
			ja.AppliesTo.ResourceTypes = make([]string, len(a.AppliesToVal.ResourceTypes))
			for i, ref := range a.AppliesToVal.ResourceTypes {
				ja.AppliesTo.ResourceTypes[i] = string(ref.Name)
			}
		}
		if a.AppliesToVal.Context != nil {
			ja.AppliesTo.Context = typeToJSON(a.AppliesToVal.Context, entityNames)
		}
	}

	ns.Actions[string(a.Name)] = ja
}

func addCommonTypeToNamespace(ns *jsonNamespace, ct *CommonTypeNode, entityNames map[string]bool) {
	ns.CommonTypes[string(ct.Name)] = typeToJSON(ct.Type, entityNames)
}

func typeToJSON(t IsType, entityNames map[string]bool) *jsonType {
	switch v := t.(type) {
	case StringType:
		return &jsonType{Type: "String"}
	case LongType:
		return &jsonType{Type: "Long"}
	case BoolType:
		return &jsonType{Type: "Boolean"}
	case ExtensionType:
		return &jsonType{Type: "Extension", Name: string(v.Name)}
	case SetType:
		return &jsonType{Type: "Set", Element: typeToJSON(v.Element, entityNames)}
	case RecordType:
		jt := &jsonType{
			Type:       "Record",
			Attributes: make(map[string]*jsonAttr),
		}
		for _, pair := range v.Pairs {
			attr := attrToJSON(pair.Type, entityNames)
			if pair.Optional {
				f := false
				attr.Required = &f
			}
			jt.Attributes[string(pair.Key)] = attr
		}
		return jt
	case EntityTypeRef:
		return &jsonType{Type: "Entity", Name: string(v.Name)}
	case TypeRef:
		// Check if this type name refers to an entity
		name := string(v.Name)
		if entityNames[name] {
			return &jsonType{Type: "Entity", Name: name}
		}
		return &jsonType{Type: name}
	default:
		return nil
	}
}

func attrToJSON(t IsType, entityNames map[string]bool) *jsonAttr {
	switch v := t.(type) {
	case StringType:
		return &jsonAttr{Type: "String"}
	case LongType:
		return &jsonAttr{Type: "Long"}
	case BoolType:
		return &jsonAttr{Type: "Boolean"}
	case ExtensionType:
		return &jsonAttr{Type: "Extension", Name: string(v.Name)}
	case SetType:
		return &jsonAttr{Type: "Set", Element: typeToJSON(v.Element, entityNames)}
	case RecordType:
		ja := &jsonAttr{
			Type:       "Record",
			Attributes: make(map[string]*jsonAttr),
		}
		for _, pair := range v.Pairs {
			attr := attrToJSON(pair.Type, entityNames)
			if pair.Optional {
				f := false
				attr.Required = &f
			}
			ja.Attributes[string(pair.Key)] = attr
		}
		return ja
	case EntityTypeRef:
		return &jsonAttr{Type: "Entity", Name: string(v.Name)}
	case TypeRef:
		// Check if this type name refers to an entity
		name := string(v.Name)
		if entityNames[name] {
			return &jsonAttr{Type: "Entity", Name: name}
		}
		return &jsonAttr{Type: name}
	default:
		return nil
	}
}

// Helper functions for unmarshalling

func parseNamespaceContents(ns *jsonNamespace) ([]IsNode, error) {
	var nodes []IsNode

	// Parse common types first (sorted for determinism)
	ctNames := make([]string, 0, len(ns.CommonTypes))
	for name := range ns.CommonTypes {
		ctNames = append(ctNames, name)
	}
	sort.Strings(ctNames)
	for _, name := range ctNames {
		jt := ns.CommonTypes[name]
		t, err := jsonToType(jt)
		if err != nil {
			return nil, fmt.Errorf("parsing common type %s: %w", name, err)
		}
		nodes = append(nodes, &CommonTypeNode{
			Name: types.Ident(name),
			Type: t,
		})
	}

	// Parse entity types (sorted for determinism)
	etNames := make([]string, 0, len(ns.EntityTypes))
	for name := range ns.EntityTypes {
		etNames = append(etNames, name)
	}
	sort.Strings(etNames)
	for _, name := range etNames {
		je := ns.EntityTypes[name]
		node, err := parseEntity(name, je)
		if err != nil {
			return nil, fmt.Errorf("parsing entity %s: %w", name, err)
		}
		nodes = append(nodes, node)
	}

	// Parse actions (sorted for determinism)
	actNames := make([]string, 0, len(ns.Actions))
	for name := range ns.Actions {
		actNames = append(actNames, name)
	}
	sort.Strings(actNames)
	for _, name := range actNames {
		ja := ns.Actions[name]
		node, err := parseAction(name, ja)
		if err != nil {
			return nil, fmt.Errorf("parsing action %s: %w", name, err)
		}
		nodes = append(nodes, node)
	}

	return nodes, nil
}

func parseEntity(name string, je *jsonEntity) (IsDeclaration, error) {
	// Check if it's an enum
	if len(je.Enum) > 0 {
		values := make([]types.String, len(je.Enum))
		for i, v := range je.Enum {
			values[i] = types.String(v)
		}
		return &EnumNode{
			Name:   types.EntityType(name),
			Values: values,
		}, nil
	}

	e := &EntityNode{
		Name: types.EntityType(name),
	}

	if len(je.MemberOfTypes) > 0 {
		e.MemberOfVal = make([]EntityTypeRef, len(je.MemberOfTypes))
		for i, ref := range je.MemberOfTypes {
			e.MemberOfVal[i] = EntityTypeRef{Name: types.EntityType(ref)}
		}
	}

	if je.Shape != nil {
		t, err := jsonToType(je.Shape)
		if err != nil {
			return nil, err
		}
		if rt, ok := t.(RecordType); ok {
			e.ShapeVal = &rt
		}
	}

	if je.Tags != nil {
		t, err := jsonToType(je.Tags)
		if err != nil {
			return nil, err
		}
		e.TagsVal = t
	}

	return e, nil
}

func parseAction(name string, ja *jsonAction) (IsDeclaration, error) {
	a := &ActionNode{
		Name: types.String(name),
	}

	if len(ja.MemberOf) > 0 {
		a.MemberOfVal = make([]EntityRef, len(ja.MemberOf))
		for i, ref := range ja.MemberOf {
			a.MemberOfVal[i] = EntityRef{
				Type: EntityTypeRef{Name: types.EntityType(ref.Type)},
				ID:   types.String(ref.ID),
			}
		}
	}

	if ja.AppliesTo != nil {
		a.AppliesToVal = &AppliesTo{}
		if len(ja.AppliesTo.PrincipalTypes) > 0 {
			a.AppliesToVal.PrincipalTypes = make([]EntityTypeRef, len(ja.AppliesTo.PrincipalTypes))
			for i, ref := range ja.AppliesTo.PrincipalTypes {
				a.AppliesToVal.PrincipalTypes[i] = EntityTypeRef{Name: types.EntityType(ref)}
			}
		}
		if len(ja.AppliesTo.ResourceTypes) > 0 {
			a.AppliesToVal.ResourceTypes = make([]EntityTypeRef, len(ja.AppliesTo.ResourceTypes))
			for i, ref := range ja.AppliesTo.ResourceTypes {
				a.AppliesToVal.ResourceTypes[i] = EntityTypeRef{Name: types.EntityType(ref)}
			}
		}
		if ja.AppliesTo.Context != nil {
			t, err := jsonToType(ja.AppliesTo.Context)
			if err != nil {
				return nil, err
			}
			a.AppliesToVal.Context = t
		}
	}

	return a, nil
}

func jsonToType(jt *jsonType) (IsType, error) {
	switch jt.Type {
	case "String":
		return StringType{}, nil
	case "Long":
		return LongType{}, nil
	case "Boolean":
		return BoolType{}, nil
	case "Extension":
		return ExtensionType{Name: types.Ident(jt.Name)}, nil
	case "Set":
		if jt.Element == nil {
			return nil, fmt.Errorf("Set type missing element")
		}
		elem, err := jsonToType(jt.Element)
		if err != nil {
			return nil, err
		}
		return SetType{Element: elem}, nil
	case "Record":
		pairs := make([]Pair, 0, len(jt.Attributes))
		// Sort attribute names for determinism
		attrNames := make([]string, 0, len(jt.Attributes))
		for name := range jt.Attributes {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)
		for _, name := range attrNames {
			attr := jt.Attributes[name]
			t, err := jsonAttrToType(attr)
			if err != nil {
				return nil, err
			}
			optional := attr.Required != nil && !*attr.Required
			pairs = append(pairs, Pair{
				Key:      types.String(name),
				Type:     t,
				Optional: optional,
			})
		}
		return RecordType{Pairs: pairs}, nil
	case "Entity":
		return EntityTypeRef{Name: types.EntityType(jt.Name)}, nil
	default:
		// Assume it's a type reference
		if jt.Type != "" {
			return TypeRef{Name: types.Path(jt.Type)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", jt)
	}
}

func jsonAttrToType(ja *jsonAttr) (IsType, error) {
	switch ja.Type {
	case "String":
		return StringType{}, nil
	case "Long":
		return LongType{}, nil
	case "Boolean":
		return BoolType{}, nil
	case "Extension":
		return ExtensionType{Name: types.Ident(ja.Name)}, nil
	case "Set":
		if ja.Element == nil {
			return nil, fmt.Errorf("Set type missing element")
		}
		elem, err := jsonToType(ja.Element)
		if err != nil {
			return nil, err
		}
		return SetType{Element: elem}, nil
	case "Record":
		pairs := make([]Pair, 0, len(ja.Attributes))
		// Sort attribute names for determinism
		attrNames := make([]string, 0, len(ja.Attributes))
		for name := range ja.Attributes {
			attrNames = append(attrNames, name)
		}
		sort.Strings(attrNames)
		for _, name := range attrNames {
			attr := ja.Attributes[name]
			t, err := jsonAttrToType(attr)
			if err != nil {
				return nil, err
			}
			optional := attr.Required != nil && !*attr.Required
			pairs = append(pairs, Pair{
				Key:      types.String(name),
				Type:     t,
				Optional: optional,
			})
		}
		return RecordType{Pairs: pairs}, nil
	case "Entity":
		return EntityTypeRef{Name: types.EntityType(ja.Name)}, nil
	default:
		// Assume it's a type reference
		if ja.Type != "" {
			return TypeRef{Name: types.Path(ja.Type)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", ja)
	}
}
