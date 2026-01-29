// Package json provides JSON marshalling and unmarshalling for Cedar schemas.
package json

import (
	"encoding/json"
	"fmt"
	"maps"
	"slices"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// Schema is a type alias to ast.Schema that provides JSON marshaling methods.
type Schema ast.Schema

// MarshalJSON serializes the schema to Cedar JSON schema format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	astSchema := (*ast.Schema)(s)
	namespaces := make(map[string]*Namespace)

	// Add default namespace for top-level declarations
	if len(astSchema.Entities) > 0 || len(astSchema.Enums) > 0 || len(astSchema.Actions) > 0 || len(astSchema.CommonTypes) > 0 {
		ns := getOrCreateNamespace(namespaces, "", nil)

		// Add common types
		for name, ct := range astSchema.CommonTypes {
			addCommonTypeToNamespace(ns, string(name), ct)
		}

		// Add entities
		for name, e := range astSchema.Entities {
			addEntityToNamespace(ns, string(name), e)
		}

		// Add enums
		for name, e := range astSchema.Enums {
			addEnumToNamespace(ns, string(name), e)
		}

		// Add actions
		for name, a := range astSchema.Actions {
			addActionToNamespace(ns, string(name), a)
		}
	}

	// Add named namespaces
	for nsPath, nsNode := range astSchema.Namespaces {
		ns := getOrCreateNamespace(namespaces, string(nsPath), annotationsToMap(nsNode.Annotations))

		// Add common types
		for name, ct := range nsNode.CommonTypes {
			addCommonTypeToNamespace(ns, string(name), ct)
		}

		// Add entities
		for name, e := range nsNode.Entities {
			addEntityToNamespace(ns, string(name), e)
		}

		// Add enums
		for name, e := range nsNode.Enums {
			addEnumToNamespace(ns, string(name), e)
		}

		// Add actions
		for name, a := range nsNode.Actions {
			addActionToNamespace(ns, string(name), a)
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

	for nsName := range namespaces {
		ns := namespaces[nsName]
		if nsName == "" {
			// Default namespace - add declarations at top level
			if err := parseNamespaceIntoSchema(astSchema, ns); err != nil {
				return err
			}
		} else {
			// Named namespace
			nsNode := ast.Namespace{
				Annotations: mapToAnnotations(ns.Annotations),
			}

			if err := parseNamespaceIntoNode(&nsNode, ns); err != nil {
				return err
			}

			if astSchema.Namespaces == nil {
				astSchema.Namespaces = make(ast.Namespaces)
			}
			astSchema.Namespaces[types.Path(nsName)] = nsNode
		}
	}

	*s = Schema(*astSchema)
	return nil
}

// Helper functions for marshalling

func annotationsToMap(annotations ast.Annotations) map[string]string {
	if len(annotations) == 0 {
		return nil
	}
	result := make(map[string]string, len(annotations))
	for key, value := range annotations {
		result[string(key)] = string(value)
	}
	return result
}

func mapToAnnotations(m map[string]string) ast.Annotations {
	if len(m) == 0 {
		return nil
	}
	result := ast.Annotations{}
	for k, v := range m {
		result[types.Ident(k)] = types.String(v)
	}
	return result
}

func getOrCreateNamespace(namespaces map[string]*Namespace, name string, annotations map[string]string) *Namespace {
	if ns, ok := namespaces[name]; ok {
		return ns
	}
	ns := &Namespace{
		CommonTypes: make(map[string]*Type),
		EntityTypes: make(map[string]*Entity),
		Actions:     make(map[string]*Action),
		Annotations: annotations,
	}

	namespaces[name] = ns
	return ns
}

func addEntityToNamespace(ns *Namespace, name string, e ast.Entity) {
	je := &Entity{}

	if len(e.MemberOf) > 0 {
		je.MemberOfTypes = make([]string, len(e.MemberOf))
		for i, ref := range e.MemberOf {
			je.MemberOfTypes[i] = string(ref)
		}
	}

	if e.Shape != nil && len(*e.Shape) > 0 {
		je.Shape = typeToJSON(*e.Shape)
	}

	if e.Tags != nil {
		je.Tags = typeToJSON(e.Tags)
	}

	je.Annotations = annotationsToMap(e.Annotations)

	ns.EntityTypes[name] = je
}

func addEnumToNamespace(ns *Namespace, name string, e ast.Enum) {
	je := &Entity{
		Enum: make([]string, len(e.Values)),
	}
	for i, v := range e.Values {
		je.Enum[i] = string(v)
	}

	je.Annotations = annotationsToMap(e.Annotations)

	ns.EntityTypes[name] = je
}

func addActionToNamespace(ns *Namespace, name string, a ast.Action) {
	ja := &Action{}

	if len(a.MemberOf) > 0 {
		ja.MemberOf = make([]EntityUID, len(a.MemberOf))
		for i, ref := range a.MemberOf {
			refType := string(ref.Type)
			ja.MemberOf[i] = EntityUID{
				Type: refType,
				ID:   string(ref.ID),
			}
		}
	}

	if a.AppliesTo != nil {
		// Only emit appliesTo if there's meaningful content
		hasPrincipals := len(a.AppliesTo.Principals) > 0
		hasResources := len(a.AppliesTo.Resources) > 0
		hasContext := a.AppliesTo.Context != nil

		if hasPrincipals || hasResources || hasContext {
			ja.AppliesTo = &AppliesTo{}
			if hasPrincipals {
				ja.AppliesTo.PrincipalTypes = make([]string, len(a.AppliesTo.Principals))
				for i, ref := range a.AppliesTo.Principals {
					ja.AppliesTo.PrincipalTypes[i] = string(ref)
				}
			}
			if hasResources {
				ja.AppliesTo.ResourceTypes = make([]string, len(a.AppliesTo.Resources))
				for i, ref := range a.AppliesTo.Resources {
					ja.AppliesTo.ResourceTypes[i] = string(ref)
				}
			}
			if hasContext {
				ja.AppliesTo.Context = typeToJSONFromContext(a.AppliesTo.Context, true)
			}
		}
	}

	ja.Annotations = annotationsToMap(a.Annotations)

	ns.Actions[name] = ja
}

func addCommonTypeToNamespace(ns *Namespace, name string, ct ast.CommonType) {
	jt := typeToJSON(ct.Type)
	if jt == nil {
		jt = &Type{}
	}
	jt.Annotations = annotationsToMap(ct.Annotations)
	ns.CommonTypes[name] = jt
}

func typeToJSON(t ast.IsType) *Type {
	return typeToJSONFromContext(t, false)
}

func typeToJSONFromContext(t ast.IsType, fromContext bool) *Type {
	switch v := t.(type) {
	case ast.StringType:
		return &Type{TypeName: "EntityOrCommon", Name: "String"}
	case ast.LongType:
		return &Type{TypeName: "EntityOrCommon", Name: "Long"}
	case ast.BoolType:
		return &Type{TypeName: "EntityOrCommon", Name: "Bool"}
	case ast.ExtensionType:
		return &Type{TypeName: "Extension", Name: string(v)}
	case ast.SetType:
		return &Type{TypeName: "Set", Element: typeToJSON(v.Element)}
	case ast.RecordType:
		jt := &Type{
			TypeName:   "Record",
			Attributes: make(map[string]*Attr),
		}
		for key, attr := range v {
			jattr := attrToJSON(attr.Type)
			if attr.Optional {
				f := false
				jattr.Required = &f
			}
			if len(attr.Annotations) > 0 {
				jattr.Annotations = annotationsToMap(attr.Annotations)
			}
			jt.Attributes[string(key)] = jattr
		}
		return jt
	case ast.EntityTypeRef:
		return &Type{TypeName: "EntityOrCommon", Name: string(v)}
	case ast.TypeRef:
		name := string(v)
		if !fromContext {
			return &Type{TypeName: "EntityOrCommon", Name: name}
		}
		// Common type reference
		return &Type{TypeName: name}
	default:
		return nil
	}
}

func attrToJSON(t ast.IsType) *Attr {
	typ := typeToJSON(t)
	return &Attr{
		TypeName:    typ.TypeName,
		Name:        typ.Name,
		Element:     typ.Element,
		Attributes:  typ.Attributes,
		Annotations: typ.Annotations,
	}
}

// Helper functions for unmarshalling

func parseNamespaceIntoSchema(schema *ast.Schema, ns *Namespace) error {
	// Parse common types first (sorted for determinism)
	ctNames := slices.Sorted(maps.Keys(ns.CommonTypes))
	for _, name := range ctNames {
		jt := ns.CommonTypes[name]
		t, err := jsonToType(jt)
		if err != nil {
			return fmt.Errorf("parsing common type %s: %w", name, err)
		}
		if schema.CommonTypes == nil {
			schema.CommonTypes = make(ast.CommonTypes)
		}
		schema.CommonTypes[types.Ident(name)] = ast.CommonType{
			Type:        t,
			Annotations: mapToAnnotations(jt.Annotations),
		}
	}

	// Parse entity types (sorted for determinism)
	etNames := slices.Sorted(maps.Keys(ns.EntityTypes))
	for _, name := range etNames {
		je := ns.EntityTypes[name]

		// Check if it's an enum
		if len(je.Enum) > 0 {
			values := make([]types.String, len(je.Enum))
			for i, v := range je.Enum {
				values[i] = types.String(v)
			}
			if schema.Enums == nil {
				schema.Enums = make(ast.Enums)
			}
			schema.Enums[types.EntityType(name)] = ast.Enum{
				Values:      values,
				Annotations: mapToAnnotations(je.Annotations),
			}
		} else {
			entity, err := parseEntity(je)
			if err != nil {
				return fmt.Errorf("parsing entity %s: %w", name, err)
			}
			if schema.Entities == nil {
				schema.Entities = make(ast.Entities)
			}
			schema.Entities[types.EntityType(name)] = entity
		}
	}

	// Parse actions (sorted for determinism)
	actNames := slices.Sorted(maps.Keys(ns.Actions))
	for _, name := range actNames {
		ja := ns.Actions[name]
		action, err := parseAction(ja)
		if err != nil {
			return fmt.Errorf("parsing action %s: %w", name, err)
		}
		if schema.Actions == nil {
			schema.Actions = make(ast.Actions)
		}
		schema.Actions[types.String(name)] = action
	}

	return nil
}

func parseNamespaceIntoNode(nsNode *ast.Namespace, ns *Namespace) error {
	// Parse common types first (sorted for determinism)
	ctNames := slices.Sorted(maps.Keys(ns.CommonTypes))
	for _, name := range ctNames {
		jt := ns.CommonTypes[name]
		t, err := jsonToType(jt)
		if err != nil {
			return fmt.Errorf("parsing common type %s: %w", name, err)
		}
		if nsNode.CommonTypes == nil {
			nsNode.CommonTypes = make(ast.CommonTypes)
		}
		nsNode.CommonTypes[types.Ident(name)] = ast.CommonType{
			Type:        t,
			Annotations: mapToAnnotations(jt.Annotations),
		}
	}

	// Parse entity types (sorted for determinism)
	etNames := slices.Sorted(maps.Keys(ns.EntityTypes))
	for _, name := range etNames {
		je := ns.EntityTypes[name]

		// Check if it's an enum
		if len(je.Enum) > 0 {
			values := make([]types.String, len(je.Enum))
			for i, v := range je.Enum {
				values[i] = types.String(v)
			}
			if nsNode.Enums == nil {
				nsNode.Enums = make(ast.Enums)
			}
			nsNode.Enums[types.EntityType(name)] = ast.Enum{
				Values:      values,
				Annotations: mapToAnnotations(je.Annotations),
			}
		} else {
			entity, err := parseEntity(je)
			if err != nil {
				return fmt.Errorf("parsing entity %s: %w", name, err)
			}
			if nsNode.Entities == nil {
				nsNode.Entities = make(ast.Entities)
			}
			nsNode.Entities[types.EntityType(name)] = entity
		}
	}

	// Parse actions (sorted for determinism)
	actNames := slices.Sorted(maps.Keys(ns.Actions))
	for _, name := range actNames {
		ja := ns.Actions[name]
		action, err := parseAction(ja)
		if err != nil {
			return fmt.Errorf("parsing action %s: %w", name, err)
		}
		if nsNode.Actions == nil {
			nsNode.Actions = make(ast.Actions)
		}
		nsNode.Actions[types.String(name)] = action
	}

	return nil
}

func parseEntity(je *Entity) (ast.Entity, error) {
	e := ast.Entity{}

	if len(je.MemberOfTypes) > 0 {
		e.MemberOf = make([]ast.EntityTypeRef, len(je.MemberOfTypes))
		for i, ref := range je.MemberOfTypes {
			e.MemberOf[i] = ast.EntityTypeRef(ref)
		}
	}

	if je.Shape != nil {
		t, err := jsonToType(je.Shape)
		if err != nil {
			return e, err
		}
		if rt, ok := t.(ast.RecordType); ok {
			e.Shape = &rt
		}
	}

	if je.Tags != nil {
		t, err := jsonToType(je.Tags)
		if err != nil {
			return e, err
		}
		e.Tags = t
	}

	e.Annotations = mapToAnnotations(je.Annotations)

	return e, nil
}

func parseAction(ja *Action) (ast.Action, error) {
	a := ast.Action{}

	if len(ja.MemberOf) > 0 {
		a.MemberOf = make([]ast.EntityRef, len(ja.MemberOf))
		for i, ref := range ja.MemberOf {
			refType := ref.Type
			a.MemberOf[i] = ast.EntityRef{
				Type: ast.EntityTypeRef(refType),
				ID:   types.String(ref.ID),
			}
		}
	}

	if ja.AppliesTo != nil {
		// Only create AppliesTo if there's meaningful content
		hasPrincipals := len(ja.AppliesTo.PrincipalTypes) > 0
		hasResources := len(ja.AppliesTo.ResourceTypes) > 0
		hasContext := ja.AppliesTo.Context != nil

		if hasPrincipals || hasResources || hasContext {
			a.AppliesTo = &ast.AppliesTo{}
			if hasPrincipals {
				a.AppliesTo.Principals = make([]ast.EntityTypeRef, len(ja.AppliesTo.PrincipalTypes))
				for i, ref := range ja.AppliesTo.PrincipalTypes {
					a.AppliesTo.Principals[i] = ast.EntityTypeRef(ref)
				}
			}
			if hasResources {
				a.AppliesTo.Resources = make([]ast.EntityTypeRef, len(ja.AppliesTo.ResourceTypes))
				for i, ref := range ja.AppliesTo.ResourceTypes {
					a.AppliesTo.Resources[i] = ast.EntityTypeRef(ref)
				}
			}
			if hasContext {
				t, err := jsonToType(ja.AppliesTo.Context)
				if err != nil {
					return a, err
				}
				a.AppliesTo.Context = t
			}
		}
	}

	a.Annotations = mapToAnnotations(ja.Annotations)

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
		return ast.ExtensionType(types.Ident(jt.Name)), nil
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
		var attrs ast.RecordType

		// Sort attribute names for determinism
		attrNames := slices.Sorted(maps.Keys(jt.Attributes))
		for _, name := range attrNames {
			attr := jt.Attributes[name]
			t, err := jsonAttrToType(attr)
			if err != nil {
				return nil, err
			}
			optional := attr.Required != nil && !*attr.Required
			if attrs == nil {
				attrs = make(ast.RecordType)
			}
			attrs[types.String(name)] = ast.Attribute{
				Type:        t,
				Optional:    optional,
				Annotations: mapToAnnotations(attr.Annotations),
			}
		}
		return ast.RecordType(attrs), nil
	case "Entity":
		return ast.EntityTypeRef(jt.Name), nil
	case "EntityOrCommon":
		// EntityOrCommon is used by Rust CLI v4.8+ for all type references
		// The actual type is in the "name" field
		return resolveEntityOrCommon(jt.Name)
	default:
		// Assume it's a type reference
		if jt.TypeName != "" {
			return ast.TypeRef(jt.TypeName), nil
		}
		return nil, fmt.Errorf("unknown type: %v", jt)
	}
}

// resolveEntityOrCommon converts an EntityOrCommon name to the appropriate type.
// Rust CLI v4.8+ uses EntityOrCommon for all types including primitives.
func resolveEntityOrCommon(name string) (ast.IsType, error) {
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
		return ast.TypeRef(name), nil
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
		return ast.ExtensionType(types.Ident(ja.Name)), nil
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
		var attrs ast.RecordType

		// Sort attribute names for determinism
		attrNames := slices.Sorted(maps.Keys(ja.Attributes))
		for _, name := range attrNames {
			attr := ja.Attributes[name]
			t, err := jsonAttrToType(attr)
			if err != nil {
				return nil, err
			}
			optional := attr.Required != nil && !*attr.Required
			if attrs == nil {
				attrs = make(ast.RecordType)
			}
			attrs[types.String(name)] = ast.Attribute{
				Type:        t,
				Optional:    optional,
				Annotations: mapToAnnotations(attr.Annotations),
			}
		}
		return ast.RecordType(attrs), nil
	case "Entity":
		return ast.EntityTypeRef(ja.Name), nil
	case "EntityOrCommon":
		// EntityOrCommon is used by Rust CLI v4.8+ for all type references
		return resolveEntityOrCommon(ja.Name)
	default:
		// Assume it's a type reference
		if ja.TypeName != "" {
			return ast.TypeRef(ja.TypeName), nil
		}
		return nil, fmt.Errorf("unknown type: %v", ja)
	}
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
