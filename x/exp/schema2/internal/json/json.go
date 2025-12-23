// Package json provides JSON marshalling and unmarshalling for Cedar schemas.
package json

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// Marshal serializes an AST schema to Cedar JSON schema format.
func Marshal(s *ast.Schema) ([]byte, error) {
	// First pass: collect all entity names for type resolution
	entityNames := collectEntityNames(s)

	namespaces := make(map[string]*Namespace)

	// Second pass: group nodes by namespace
	for _, node := range s.Nodes {
		switch n := node.(type) {
		case *ast.NamespaceNode:
			ns := getOrCreateNamespace(namespaces, string(n.Name))
			for _, decl := range n.Declarations {
				addDeclToNamespace(ns, decl, entityNames)
			}
		case *ast.EntityNode:
			ns := getOrCreateNamespace(namespaces, "")
			addEntityToNamespace(ns, n, entityNames)
		case *ast.EnumNode:
			ns := getOrCreateNamespace(namespaces, "")
			addEnumToNamespace(ns, n)
		case *ast.ActionNode:
			ns := getOrCreateNamespace(namespaces, "")
			addActionToNamespace(ns, n, entityNames)
		case *ast.CommonTypeNode:
			ns := getOrCreateNamespace(namespaces, "")
			addCommonTypeToNamespace(ns, n, entityNames)
		}
	}

	return json.MarshalIndent(namespaces, "", "    ")
}

// Unmarshal deserializes Cedar JSON schema format into an AST.
func Unmarshal(data []byte) (*ast.Schema, error) {
	namespaces := make(map[string]*Namespace)
	if err := json.Unmarshal(data, &namespaces); err != nil {
		return nil, err
	}

	s := &ast.Schema{}

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
				return nil, err
			}
			s.Nodes = append(s.Nodes, nodes...)
		} else {
			// Named namespace
			decls, err := parseNamespaceContents(ns)
			if err != nil {
				return nil, err
			}
			declarations := make([]ast.IsDeclaration, 0, len(decls))
			for _, d := range decls {
				if decl, ok := d.(ast.IsDeclaration); ok {
					declarations = append(declarations, decl)
				}
			}
			s.Nodes = append(s.Nodes, &ast.NamespaceNode{
				Name:         types.Path(nsName),
				Declarations: declarations,
			})
		}
	}

	return s, nil
}

// collectEntityNames collects all entity names from the schema.
func collectEntityNames(s *ast.Schema) map[string]bool {
	names := make(map[string]bool)
	for _, node := range s.Nodes {
		switch n := node.(type) {
		case *ast.NamespaceNode:
			for _, decl := range n.Declarations {
				switch d := decl.(type) {
				case *ast.EntityNode:
					names[string(d.Name)] = true
				case *ast.EnumNode:
					names[string(d.Name)] = true
				}
			}
		case *ast.EntityNode:
			names[string(n.Name)] = true
		case *ast.EnumNode:
			names[string(n.Name)] = true
		}
	}
	return names
}

// JSON schema types

// Namespace represents a Cedar namespace in JSON format.
type Namespace struct {
	CommonTypes map[string]*Type   `json:"commonTypes,omitempty"`
	EntityTypes map[string]*Entity `json:"entityTypes"`
	Actions     map[string]*Action `json:"actions"`
}

// Entity represents a Cedar entity type in JSON format.
type Entity struct {
	MemberOfTypes []string `json:"memberOfTypes,omitempty"`
	Shape         *Type    `json:"shape,omitempty"`
	Tags          *Type    `json:"tags,omitempty"`
	Enum          []string `json:"enum,omitempty"`
}

// Action represents a Cedar action in JSON format.
type Action struct {
	MemberOf  []EntityUID `json:"memberOf,omitempty"`
	AppliesTo *AppliesTo  `json:"appliesTo,omitempty"`
}

// EntityUID represents an entity reference in JSON format.
type EntityUID struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// AppliesTo represents action constraints in JSON format.
type AppliesTo struct {
	PrincipalTypes []string `json:"principalTypes,omitempty"`
	ResourceTypes  []string `json:"resourceTypes,omitempty"`
	Context        *Type    `json:"context,omitempty"`
}

// Type represents a Cedar type in JSON format.
type Type struct {
	TypeName   string           `json:"type,omitempty"`
	Name       string           `json:"name,omitempty"`
	Element    *Type            `json:"element,omitempty"`
	Attributes map[string]*Attr `json:"attributes,omitempty"`
}

// Attr represents a record attribute in JSON format.
type Attr struct {
	TypeName   string           `json:"type,omitempty"`
	Name       string           `json:"name,omitempty"`
	Element    *Type            `json:"element,omitempty"`
	Attributes map[string]*Attr `json:"attributes,omitempty"`
	Required   *bool            `json:"required,omitempty"`
}

// Helper functions for marshalling

func getOrCreateNamespace(namespaces map[string]*Namespace, name string) *Namespace {
	if ns, ok := namespaces[name]; ok {
		return ns
	}
	ns := &Namespace{
		CommonTypes: make(map[string]*Type),
		EntityTypes: make(map[string]*Entity),
		Actions:     make(map[string]*Action),
	}
	namespaces[name] = ns
	return ns
}

func addDeclToNamespace(ns *Namespace, decl ast.IsDeclaration, entityNames map[string]bool) {
	switch d := decl.(type) {
	case *ast.EntityNode:
		addEntityToNamespace(ns, d, entityNames)
	case *ast.EnumNode:
		addEnumToNamespace(ns, d)
	case *ast.ActionNode:
		addActionToNamespace(ns, d, entityNames)
	case *ast.CommonTypeNode:
		addCommonTypeToNamespace(ns, d, entityNames)
	}
}

func addEntityToNamespace(ns *Namespace, e *ast.EntityNode, entityNames map[string]bool) {
	je := &Entity{}

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

func addEnumToNamespace(ns *Namespace, e *ast.EnumNode) {
	je := &Entity{
		Enum: make([]string, len(e.Values)),
	}
	for i, v := range e.Values {
		je.Enum[i] = string(v)
	}
	ns.EntityTypes[string(e.Name)] = je
}

func addActionToNamespace(ns *Namespace, a *ast.ActionNode, entityNames map[string]bool) {
	ja := &Action{}

	if len(a.MemberOfVal) > 0 {
		ja.MemberOf = make([]EntityUID, len(a.MemberOfVal))
		for i, ref := range a.MemberOfVal {
			ja.MemberOf[i] = EntityUID{
				Type: string(ref.Type.Name),
				ID:   string(ref.ID),
			}
		}
	}

	if a.AppliesToVal != nil {
		ja.AppliesTo = &AppliesTo{}
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

func addCommonTypeToNamespace(ns *Namespace, ct *ast.CommonTypeNode, entityNames map[string]bool) {
	ns.CommonTypes[string(ct.Name)] = typeToJSON(ct.Type, entityNames)
}

func typeToJSON(t ast.IsType, entityNames map[string]bool) *Type {
	switch v := t.(type) {
	case ast.StringType:
		return &Type{TypeName: "String"}
	case ast.LongType:
		return &Type{TypeName: "Long"}
	case ast.BoolType:
		return &Type{TypeName: "Boolean"}
	case ast.ExtensionType:
		return &Type{TypeName: "Extension", Name: string(v.Name)}
	case ast.SetType:
		return &Type{TypeName: "Set", Element: typeToJSON(v.Element, entityNames)}
	case ast.RecordType:
		jt := &Type{
			TypeName:   "Record",
			Attributes: make(map[string]*Attr),
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
	case ast.EntityTypeRef:
		return &Type{TypeName: "Entity", Name: string(v.Name)}
	case ast.TypeRef:
		// Check if this type name refers to an entity
		name := string(v.Name)
		if entityNames[name] {
			return &Type{TypeName: "Entity", Name: name}
		}
		return &Type{TypeName: name}
	default:
		return nil
	}
}

func attrToJSON(t ast.IsType, entityNames map[string]bool) *Attr {
	switch v := t.(type) {
	case ast.StringType:
		return &Attr{TypeName: "String"}
	case ast.LongType:
		return &Attr{TypeName: "Long"}
	case ast.BoolType:
		return &Attr{TypeName: "Boolean"}
	case ast.ExtensionType:
		return &Attr{TypeName: "Extension", Name: string(v.Name)}
	case ast.SetType:
		return &Attr{TypeName: "Set", Element: typeToJSON(v.Element, entityNames)}
	case ast.RecordType:
		ja := &Attr{
			TypeName:   "Record",
			Attributes: make(map[string]*Attr),
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
	case ast.EntityTypeRef:
		return &Attr{TypeName: "Entity", Name: string(v.Name)}
	case ast.TypeRef:
		// Check if this type name refers to an entity
		name := string(v.Name)
		if entityNames[name] {
			return &Attr{TypeName: "Entity", Name: name}
		}
		return &Attr{TypeName: name}
	default:
		return nil
	}
}

// Helper functions for unmarshalling

func parseNamespaceContents(ns *Namespace) ([]ast.IsNode, error) {
	var nodes []ast.IsNode

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
		nodes = append(nodes, &ast.CommonTypeNode{
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

func parseEntity(name string, je *Entity) (ast.IsDeclaration, error) {
	// Check if it's an enum
	if len(je.Enum) > 0 {
		values := make([]types.String, len(je.Enum))
		for i, v := range je.Enum {
			values[i] = types.String(v)
		}
		return &ast.EnumNode{
			Name:   types.EntityType(name),
			Values: values,
		}, nil
	}

	e := &ast.EntityNode{
		Name: types.EntityType(name),
	}

	if len(je.MemberOfTypes) > 0 {
		e.MemberOfVal = make([]ast.EntityTypeRef, len(je.MemberOfTypes))
		for i, ref := range je.MemberOfTypes {
			e.MemberOfVal[i] = ast.EntityTypeRef{Name: types.EntityType(ref)}
		}
	}

	if je.Shape != nil {
		t, err := jsonToType(je.Shape)
		if err != nil {
			return nil, err
		}
		if rt, ok := t.(ast.RecordType); ok {
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

func parseAction(name string, ja *Action) (ast.IsDeclaration, error) {
	a := &ast.ActionNode{
		Name: types.String(name),
	}

	if len(ja.MemberOf) > 0 {
		a.MemberOfVal = make([]ast.EntityRef, len(ja.MemberOf))
		for i, ref := range ja.MemberOf {
			a.MemberOfVal[i] = ast.EntityRef{
				Type: ast.EntityTypeRef{Name: types.EntityType(ref.Type)},
				ID:   types.String(ref.ID),
			}
		}
	}

	if ja.AppliesTo != nil {
		a.AppliesToVal = &ast.AppliesTo{}
		if len(ja.AppliesTo.PrincipalTypes) > 0 {
			a.AppliesToVal.PrincipalTypes = make([]ast.EntityTypeRef, len(ja.AppliesTo.PrincipalTypes))
			for i, ref := range ja.AppliesTo.PrincipalTypes {
				a.AppliesToVal.PrincipalTypes[i] = ast.EntityTypeRef{Name: types.EntityType(ref)}
			}
		}
		if len(ja.AppliesTo.ResourceTypes) > 0 {
			a.AppliesToVal.ResourceTypes = make([]ast.EntityTypeRef, len(ja.AppliesTo.ResourceTypes))
			for i, ref := range ja.AppliesTo.ResourceTypes {
				a.AppliesToVal.ResourceTypes[i] = ast.EntityTypeRef{Name: types.EntityType(ref)}
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

func jsonToType(jt *Type) (ast.IsType, error) {
	switch jt.TypeName {
	case "String":
		return ast.StringType{}, nil
	case "Long":
		return ast.LongType{}, nil
	case "Boolean":
		return ast.BoolType{}, nil
	case "Extension":
		return ast.ExtensionType{Name: types.Ident(jt.Name)}, nil
	case "Set":
		if jt.Element == nil {
			return nil, fmt.Errorf("set type missing element")
		}
		elem, err := jsonToType(jt.Element)
		if err != nil {
			return nil, err
		}
		return ast.SetType{Element: elem}, nil
	case "Record":
		pairs := make([]ast.Pair, 0, len(jt.Attributes))
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
			pairs = append(pairs, ast.Pair{
				Key:      types.String(name),
				Type:     t,
				Optional: optional,
			})
		}
		return ast.RecordType{Pairs: pairs}, nil
	case "Entity":
		return ast.EntityTypeRef{Name: types.EntityType(jt.Name)}, nil
	default:
		// Assume it's a type reference
		if jt.TypeName != "" {
			return ast.TypeRef{Name: types.Path(jt.TypeName)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", jt)
	}
}

func jsonAttrToType(ja *Attr) (ast.IsType, error) {
	switch ja.TypeName {
	case "String":
		return ast.StringType{}, nil
	case "Long":
		return ast.LongType{}, nil
	case "Boolean":
		return ast.BoolType{}, nil
	case "Extension":
		return ast.ExtensionType{Name: types.Ident(ja.Name)}, nil
	case "Set":
		if ja.Element == nil {
			return nil, fmt.Errorf("set type missing element")
		}
		elem, err := jsonToType(ja.Element)
		if err != nil {
			return nil, err
		}
		return ast.SetType{Element: elem}, nil
	case "Record":
		pairs := make([]ast.Pair, 0, len(ja.Attributes))
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
			pairs = append(pairs, ast.Pair{
				Key:      types.String(name),
				Type:     t,
				Optional: optional,
			})
		}
		return ast.RecordType{Pairs: pairs}, nil
	case "Entity":
		return ast.EntityTypeRef{Name: types.EntityType(ja.Name)}, nil
	default:
		// Assume it's a type reference
		if ja.TypeName != "" {
			return ast.TypeRef{Name: types.Path(ja.TypeName)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", ja)
	}
}
