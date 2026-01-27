// Package json provides JSON marshalling and unmarshalling for Cedar schemas.
package json

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// Schema is a type alias to ast.Schema that provides JSON marshaling methods.
type Schema ast.Schema

// MarshalJSON serializes the schema to Cedar JSON schema format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	// First pass: collect all entity names for type resolution
	astSchema := (*ast.Schema)(s)
	entityNames := collectEntityNames(astSchema)

	namespaces := make(map[string]*Namespace)

	// Second pass: group nodes by namespace
	for _, node := range astSchema.Nodes {
		switch n := node.(type) {
		case ast.NamespaceNode:
			ns := getOrCreateNamespace(namespaces, string(n.Name), n.Annotations)
			for _, decl := range n.Declarations {
				addDeclToNamespace(ns, decl, entityNames)
			}
		case ast.EntityNode:
			ns := getOrCreateNamespace(namespaces, "", nil)
			addEntityToNamespace(ns, n, entityNames)
		case ast.EnumNode:
			ns := getOrCreateNamespace(namespaces, "", nil)
			addEnumToNamespace(ns, n)
		case ast.ActionNode:
			ns := getOrCreateNamespace(namespaces, "", nil)
			addActionToNamespace(ns, n, entityNames)
		case ast.CommonTypeNode:
			ns := getOrCreateNamespace(namespaces, "", nil)
			addCommonTypeToNamespace(ns, n, entityNames)
		}
	}

	return json.MarshalIndent(namespaces, "", "    ")
}

// UnmarshalJSON deserializes Cedar JSON schema format into the schema.
func (s *Schema) UnmarshalJSON(data []byte) error {
	namespaces := make(map[string]*Namespace)
	if err := json.Unmarshal(data, &namespaces); err != nil {
		return err
	}

	astSchema := &ast.Schema{}

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
			astSchema.Nodes = append(astSchema.Nodes, nodes...)
		} else {
			// Named namespace
			decls, err := parseNamespaceContents(ns)
			if err != nil {
				return err
			}
			declarations := make([]ast.IsDeclaration, 0, len(decls))
			for _, d := range decls {
				if decl, ok := d.(ast.IsDeclaration); ok {
					declarations = append(declarations, decl)
				}
			}
			astSchema.Nodes = append(astSchema.Nodes, ast.NamespaceNode{
				Name:         types.Path(nsName),
				Declarations: declarations,
				Annotations:  mapToAnnotations(ns.Annotations),
			})
		}
	}

	*s = Schema(*astSchema)
	return nil
}

// collectEntityNames collects all entity names from the schema.
func collectEntityNames(s *ast.Schema) map[string]bool {
	names := make(map[string]bool)
	for _, node := range s.Nodes {
		switch n := node.(type) {
		case ast.NamespaceNode:
			for _, decl := range n.Declarations {
				switch d := decl.(type) {
				case ast.EntityNode:
					names[string(d.Name)] = true
				case ast.EnumNode:
					names[string(d.Name)] = true
				}
			}
		case ast.EntityNode:
			names[string(n.Name)] = true
		case ast.EnumNode:
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
	Annotations map[string]string  `json:"annotations,omitempty"`
}

// Entity represents a Cedar entity type in JSON format.
type Entity struct {
	MemberOfTypes []string          `json:"memberOfTypes,omitempty"`
	Shape         *Type             `json:"shape,omitempty"`
	Tags          *Type             `json:"tags,omitempty"`
	Enum          []string          `json:"enum,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

// Action represents a Cedar action in JSON format.
type Action struct {
	MemberOf    []EntityUID       `json:"memberOf,omitempty"`
	AppliesTo   *AppliesTo        `json:"appliesTo,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// EntityUID represents an entity reference in JSON format.
type EntityUID struct {
	Type string `json:"type,omitempty"`
	ID   string `json:"id"`
}

// AppliesTo represents action constraints in JSON format.
type AppliesTo struct {
	PrincipalTypes []string `json:"principalTypes"`
	ResourceTypes  []string `json:"resourceTypes"`
	Context        *Type    `json:"context,omitempty"`
}

// Type represents a Cedar type in JSON format.
type Type struct {
	TypeName    string            `json:"type,omitempty"`
	Name        string            `json:"name,omitempty"`
	Element     *Type             `json:"element,omitempty"`
	Attributes  map[string]*Attr  `json:"attributes,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// MarshalJSON ensures Record types always output the attributes field.
func (t Type) MarshalJSON() ([]byte, error) {
	if t.TypeName == "Record" {
		// Record types must always have an attributes field
		type recordType struct {
			TypeName    string            `json:"type"`
			Attributes  map[string]*Attr  `json:"attributes"`
			Annotations map[string]string `json:"annotations,omitempty"`
		}
		attrs := t.Attributes
		if attrs == nil {
			attrs = make(map[string]*Attr)
		}
		return json.Marshal(recordType{TypeName: t.TypeName, Attributes: attrs, Annotations: t.Annotations})
	}
	// Use an alias to avoid infinite recursion for non-Record types
	type TypeAlias Type
	return json.Marshal(TypeAlias(t))
}

// Attr represents a record attribute in JSON format.
type Attr struct {
	TypeName    string            `json:"type,omitempty"`
	Name        string            `json:"name,omitempty"`
	Element     *Type             `json:"element,omitempty"`
	Attributes  map[string]*Attr  `json:"attributes,omitempty"`
	Required    *bool             `json:"required,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

// MarshalJSON ensures Record types always output the attributes field.
func (a Attr) MarshalJSON() ([]byte, error) {
	if a.TypeName == "Record" {
		// Record types must always have an attributes field
		type recordAttr struct {
			TypeName    string            `json:"type"`
			Attributes  map[string]*Attr  `json:"attributes"`
			Required    *bool             `json:"required,omitempty"`
			Annotations map[string]string `json:"annotations,omitempty"`
		}
		attrs := a.Attributes
		if attrs == nil {
			attrs = make(map[string]*Attr)
		}
		return json.Marshal(recordAttr{TypeName: a.TypeName, Attributes: attrs, Required: a.Required, Annotations: a.Annotations})
	}
	// Use an alias to avoid infinite recursion for non-Record types
	type AttrAlias Attr
	return json.Marshal(AttrAlias(a))
}

// Helper functions for marshalling

func annotationsToMap(annotations []ast.Annotation) map[string]string {
	if len(annotations) == 0 {
		return nil
	}
	result := make(map[string]string, len(annotations))
	for _, ann := range annotations {
		result[string(ann.Key)] = string(ann.Value)
	}
	return result
}

func mapToAnnotations(m map[string]string) []ast.Annotation {
	if len(m) == 0 {
		return nil
	}
	// Sort keys for deterministic output
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	result := make([]ast.Annotation, 0, len(m))
	for _, k := range keys {
		result = append(result, ast.Annotation{
			Key:   types.Ident(k),
			Value: types.String(m[k]),
		})
	}
	return result
}

func getOrCreateNamespace(namespaces map[string]*Namespace, name string, annotations []ast.Annotation) *Namespace {
	if ns, ok := namespaces[name]; ok {
		return ns
	}
	ns := &Namespace{
		CommonTypes: make(map[string]*Type),
		EntityTypes: make(map[string]*Entity),
		Actions:     make(map[string]*Action),
		Annotations: make(map[string]string),
	}

	ns.Annotations = annotationsToMap(annotations)

	namespaces[name] = ns
	return ns
}

func addDeclToNamespace(ns *Namespace, decl ast.IsDeclaration, entityNames map[string]bool) {
	switch d := decl.(type) {
	case ast.EntityNode:
		addEntityToNamespace(ns, d, entityNames)
	case ast.EnumNode:
		addEnumToNamespace(ns, d)
	case ast.ActionNode:
		addActionToNamespace(ns, d, entityNames)
	case ast.CommonTypeNode:
		addCommonTypeToNamespace(ns, d, entityNames)
	}
}

func addEntityToNamespace(ns *Namespace, e ast.EntityNode, entityNames map[string]bool) {
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

	je.Annotations = annotationsToMap(e.Annotations)

	ns.EntityTypes[string(e.Name)] = je
}

func addEnumToNamespace(ns *Namespace, e ast.EnumNode) {
	je := &Entity{
		Enum: make([]string, len(e.Values)),
	}
	for i, v := range e.Values {
		je.Enum[i] = string(v)
	}

	je.Annotations = annotationsToMap(e.Annotations)

	ns.EntityTypes[string(e.Name)] = je
}

func addActionToNamespace(ns *Namespace, a ast.ActionNode, entityNames map[string]bool) {
	ja := &Action{}

	if len(a.MemberOfVal) > 0 {
		ja.MemberOf = make([]EntityUID, len(a.MemberOfVal))
		for i, ref := range a.MemberOfVal {
			// Action memberOf always refers to actions, use "Action" as type
			refType := string(ref.Type.Name)
			// if refType == "" {
			// 	refType = "Action"
			// }
			ja.MemberOf[i] = EntityUID{
				Type: refType,
				ID:   string(ref.ID),
			}
		}
	}

	if a.AppliesToVal != nil {
		// Only emit appliesTo if there's meaningful content
		hasPrincipals := len(a.AppliesToVal.PrincipalTypes) > 0
		hasResources := len(a.AppliesToVal.ResourceTypes) > 0
		hasContext := a.AppliesToVal.Context != nil

		if hasPrincipals || hasResources || hasContext {
			ja.AppliesTo = &AppliesTo{}
			if hasPrincipals {
				ja.AppliesTo.PrincipalTypes = make([]string, len(a.AppliesToVal.PrincipalTypes))
				for i, ref := range a.AppliesToVal.PrincipalTypes {
					ja.AppliesTo.PrincipalTypes[i] = string(ref.Name)
				}
			}
			if hasResources {
				ja.AppliesTo.ResourceTypes = make([]string, len(a.AppliesToVal.ResourceTypes))
				for i, ref := range a.AppliesToVal.ResourceTypes {
					ja.AppliesTo.ResourceTypes[i] = string(ref.Name)
				}
			}
			if hasContext {
				ja.AppliesTo.Context = typeToJSONFromContext(a.AppliesToVal.Context, entityNames, true)
			}
		}
	}

	ja.Annotations = annotationsToMap(a.Annotations)

	ns.Actions[string(a.Name)] = ja
}

func addCommonTypeToNamespace(ns *Namespace, ct ast.CommonTypeNode, entityNames map[string]bool) {
	jt := typeToJSON(ct.Type, entityNames)
	if jt == nil {
		jt = &Type{}
	}
	jt.Annotations = annotationsToMap(ct.Annotations)
	ns.CommonTypes[string(ct.Name)] = jt
}

func typeToJSON(t ast.IsType, entityNames map[string]bool) *Type {
	return typeToJSONFromContext(t, entityNames, false)
}

func typeToJSONFromContext(t ast.IsType, entityNames map[string]bool, fromContext bool) *Type {
	switch v := t.(type) {
	case ast.StringType:
		return &Type{TypeName: "EntityOrCommon", Name: "String"}
	case ast.LongType:
		return &Type{TypeName: "EntityOrCommon", Name: "Long"}
	case ast.BoolType:
		return &Type{TypeName: "EntityOrCommon", Name: "Bool"}
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
			if len(pair.Annotations) > 0 {
				attr.Annotations = map[string]string{}
				for _, a := range pair.Annotations {
					attr.Annotations[string(a.Key)] = string(a.Value)
				}
			}
			jt.Attributes[string(pair.Key)] = attr
		}
		return jt
	case ast.EntityTypeRef:
		return &Type{TypeName: "EntityOrCommon", Name: string(v.Name)}
	case ast.TypeRef:
		// Check if this type name refers to an entity
		name := string(v.Name)
		if !fromContext {
			return &Type{TypeName: "EntityOrCommon", Name: name}
		}
		// Common type reference
		return &Type{TypeName: name}
	default:
		return nil
	}
}

func attrToJSON(t ast.IsType, entityNames map[string]bool) *Attr {
	typ := typeToJSON(t, entityNames)
	return &Attr{
		TypeName:    typ.TypeName,
		Name:        typ.Name,
		Element:     typ.Element,
		Attributes:  typ.Attributes,
		Annotations: typ.Annotations,
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
		ct := ast.CommonTypeNode{
			Name:        types.Ident(name),
			Type:        t,
			Annotations: mapToAnnotations(jt.Annotations),
		}
		nodes = append(nodes, ct)
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
		en := ast.EnumNode{
			Name:        types.EntityType(name),
			Values:      values,
			Annotations: mapToAnnotations(je.Annotations),
		}
		return en, nil
	}

	e := ast.EntityNode{
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

	e.Annotations = mapToAnnotations(je.Annotations)

	return e, nil
}

func parseAction(name string, ja *Action) (ast.IsDeclaration, error) {
	a := ast.ActionNode{
		Name: types.String(name),
	}

	if len(ja.MemberOf) > 0 {
		a.MemberOfVal = make([]ast.EntityRef, len(ja.MemberOf))
		for i, ref := range ja.MemberOf {
			// Action memberOf always refers to actions, default to "Action" if type is empty
			refType := ref.Type
			// if refType == "" {
			// 	refType = "Action"
			// }
			a.MemberOfVal[i] = ast.EntityRef{
				Type: ast.EntityTypeRef{Name: types.EntityType(refType)},
				ID:   types.String(ref.ID),
			}
		}
	}

	if ja.AppliesTo != nil {
		// Only create AppliesToVal if there's meaningful content
		hasPrincipals := len(ja.AppliesTo.PrincipalTypes) > 0
		hasResources := len(ja.AppliesTo.ResourceTypes) > 0
		hasContext := ja.AppliesTo.Context != nil

		if hasPrincipals || hasResources || hasContext {
			a.AppliesToVal = &ast.AppliesTo{}
			if hasPrincipals {
				a.AppliesToVal.PrincipalTypes = make([]ast.EntityTypeRef, len(ja.AppliesTo.PrincipalTypes))
				for i, ref := range ja.AppliesTo.PrincipalTypes {
					a.AppliesToVal.PrincipalTypes[i] = ast.EntityTypeRef{Name: types.EntityType(ref)}
				}
			}
			if hasResources {
				a.AppliesToVal.ResourceTypes = make([]ast.EntityTypeRef, len(ja.AppliesTo.ResourceTypes))
				for i, ref := range ja.AppliesTo.ResourceTypes {
					a.AppliesToVal.ResourceTypes[i] = ast.EntityTypeRef{Name: types.EntityType(ref)}
				}
			}
			if hasContext {
				t, err := jsonToType(ja.AppliesTo.Context)
				if err != nil {
					return nil, err
				}
				a.AppliesToVal.Context = t
			}
		}
	}

	if len(ja.Annotations) > 0 {
		a.Annotations = make([]ast.Annotation, 0, len(ja.Annotations))
		for k, v := range ja.Annotations {
			a.Annotations = append(a.Annotations, ast.Annotation{
				Key:   types.Ident(k),
				Value: types.String(v),
			})
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
	case "Bool":
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
			var annotations []ast.Annotation
			annKeys := slices.Sorted(maps.Keys(attr.Annotations))
			for _, key := range annKeys {
				annotations = append(annotations, ast.Annotation{Key: types.Ident(key), Value: types.String(attr.Annotations[key])})
			}
			pairs = append(pairs, ast.Pair{
				Key:         types.String(name),
				Type:        t,
				Optional:    optional,
				Annotations: annotations,
			})
		}
		return ast.RecordType{Pairs: pairs}, nil
	case "Entity":
		return ast.EntityTypeRef{Name: types.EntityType(jt.Name)}, nil
	case "EntityOrCommon":
		// EntityOrCommon is used by Rust CLI v4.8+ for all type references
		// The actual type is in the "name" field
		return resolveEntityOrCommon(jt.Name)
	default:
		// Assume it's a type reference
		if jt.TypeName != "" {
			return ast.TypeRef{Name: types.Path(jt.TypeName)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", jt)
	}
}

// resolveEntityOrCommon converts an EntityOrCommon name to the appropriate type.
// Rust CLI v4.8+ uses EntityOrCommon for all types including primitives.
// Rust may output primitives as "__cedar::String" etc.
func resolveEntityOrCommon(name string) (ast.IsType, error) {
	// Handle __cedar:: prefix for all types
	// if len(name) > 9 && name[:9] == "__cedar::" {
	// 	extName := name[9:]
	// 	// Check if it's a primitive type with __cedar:: prefix
	// 	switch extName {
	// 	case "String":
	// 		return ast.StringType{}, nil
	// 	case "Long":
	// 		return ast.LongType{}, nil
	// 	case "Bool":
	// 		return ast.BoolType{}, nil
	// 	default:
	// 		// It's a real extension type (ipaddr, datetime, decimal, duration)
	// 		return ast.ExtensionType{Name: types.Ident(extName)}, nil
	// 	}
	// }

	// Handle unprefixed primitive types
	switch name {
	case "String":
		return ast.StringType{}, nil
	case "Long":
		return ast.LongType{}, nil
	case "Bool":
		return ast.BoolType{}, nil
	default:
		// Otherwise it's a type reference (could be entity or common type)
		return ast.TypeRef{Name: types.Path(name)}, nil
	}
}

func jsonAttrToType(ja *Attr) (ast.IsType, error) {
	switch ja.TypeName {
	case "String":
		return ast.StringType{}, nil
	case "Long":
		return ast.LongType{}, nil
	case "Bool":
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
			var annotations []ast.Annotation
			annKeys := slices.Sorted(maps.Keys(attr.Annotations))
			for _, key := range annKeys {
				annotations = append(annotations, ast.Annotation{Key: types.Ident(key), Value: types.String(attr.Annotations[key])})
			}
			pairs = append(pairs, ast.Pair{
				Key:         types.String(name),
				Type:        t,
				Optional:    optional,
				Annotations: annotations,
			})
		}
		return ast.RecordType{Pairs: pairs}, nil
	case "Entity":
		return ast.EntityTypeRef{Name: types.EntityType(ja.Name)}, nil
	case "EntityOrCommon":
		// EntityOrCommon is used by Rust CLI v4.8+ for all type references
		return resolveEntityOrCommon(ja.Name)
	default:
		// Assume it's a type reference
		if ja.TypeName != "" {
			return ast.TypeRef{Name: types.Path(ja.TypeName)}, nil
		}
		return nil, fmt.Errorf("unknown type: %v", ja)
	}
}
