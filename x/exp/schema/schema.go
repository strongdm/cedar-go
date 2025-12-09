// Package schema provides types and functions for working with Cedar schemas.
//
// Schemas can be authored programmatically using the builder pattern:
//
//	s := schema.New().
//		AddNamespace(schema.NewNamespace("MyApp").
//			AddEntityType(schema.Entity("User").
//				WithShape(schema.Record().
//					AddRequired("name", schema.String()).
//					AddOptional("email", schema.String()))).
//			AddAction(schema.NewAction("read").
//				WithAppliesTo(schema.NewAppliesTo().
//					WithPrincipals("User").
//					WithResources("Document"))))
//
// Schemas can also be parsed from JSON or Cedar format using [Schema.UnmarshalJSON]
// or [Schema.UnmarshalCedar], and serialized using [Schema.MarshalJSON] or [Schema.MarshalCedar].
package schema

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/types"
)

// Schema represents a Cedar schema containing namespaces with entity types,
// actions, and common types.
type Schema struct {
	// Namespaces contains all namespace definitions in the schema.
	// The first namespace with an empty Name is the anonymous namespace.
	Namespaces []*Namespace

	// Internal fields for parsing support
	filename    string
	jsonSchema  ast.JSONSchema
	humanSchema *ast.Schema
}

// New creates a new empty schema.
func New() *Schema {
	return &Schema{}
}

// AddNamespace adds a namespace to the schema.
func (s *Schema) AddNamespace(ns *Namespace) *Schema {
	s.Namespaces = append(s.Namespaces, ns)
	// Clear cached representations since we've modified the schema
	s.jsonSchema = nil
	s.humanSchema = nil
	return s
}

// GetNamespace returns the namespace with the given name, or nil if not found.
func (s *Schema) GetNamespace(name types.Path) *Namespace {
	for _, ns := range s.Namespaces {
		if ns.Name == name {
			return ns
		}
	}
	return nil
}

// UnmarshalCedar parses and stores the human-readable schema from src.
// Any errors returned will have file positions matching the filename set via [Schema.SetFilename].
func (s *Schema) UnmarshalCedar(src []byte) error {
	humanSchema, err := parser.ParseFile(s.filename, src)
	if err != nil {
		return err
	}

	// Convert parsed AST to our new types
	s.Namespaces = convertFromHumanAST(humanSchema)
	s.humanSchema = humanSchema
	s.jsonSchema = nil
	return nil
}

// MarshalCedar serializes the schema into the human-readable Cedar format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	// If we have namespaces, convert to human AST and format
	if len(s.Namespaces) > 0 {
		s.humanSchema = convertToHumanAST(s)
		s.jsonSchema = nil
	}

	// If we have a cached JSON schema but no human schema, convert it
	if s.humanSchema == nil && s.jsonSchema != nil {
		s.humanSchema = ast.ConvertJSON2Human(s.jsonSchema)
	}

	if s.humanSchema == nil {
		return nil, fmt.Errorf("schema is empty")
	}

	var buf bytes.Buffer
	err := ast.Format(s.humanSchema, &buf)
	return buf.Bytes(), err
}

// UnmarshalJSON deserializes a schema from Cedar JSON format.
func (s *Schema) UnmarshalJSON(src []byte) error {
	var jsonSchema ast.JSONSchema
	if err := json.Unmarshal(src, &jsonSchema); err != nil {
		return err
	}

	// Convert JSON schema to our new types
	s.Namespaces = convertFromJSONSchema(jsonSchema)
	s.jsonSchema = jsonSchema
	s.humanSchema = nil
	return nil
}

// MarshalJSON serializes the schema into Cedar JSON format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	// If we have namespaces, convert to JSON schema
	if len(s.Namespaces) > 0 {
		s.jsonSchema = convertToJSONSchema(s)
		s.humanSchema = nil
	}

	// If we have a cached human schema but no JSON schema, convert it
	if s.jsonSchema == nil && s.humanSchema != nil {
		s.jsonSchema = ast.ConvertHuman2JSON(s.humanSchema)
	}

	if s.jsonSchema == nil {
		return nil, nil
	}

	return json.Marshal(s.jsonSchema)
}

// SetFilename sets the filename for error messages from [Schema.UnmarshalCedar].
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// convertFromHumanAST converts the internal parsed AST to our new schema types.
func convertFromHumanAST(h *ast.Schema) []*Namespace {
	var namespaces []*Namespace

	// Anonymous namespace holds top-level declarations
	anonNS := &Namespace{}

	for _, decl := range h.Decls {
		switch d := decl.(type) {
		case *ast.Namespace:
			ns := convertHumanNamespace(d)
			namespaces = append(namespaces, ns)
		case *ast.Entity:
			entities := convertHumanEntity(d)
			for _, e := range entities {
				anonNS.EntityTypes = append(anonNS.EntityTypes, e)
			}
		case *ast.Action:
			actions := convertHumanAction(d)
			for _, a := range actions {
				anonNS.Actions = append(anonNS.Actions, a)
			}
		case *ast.CommonTypeDecl:
			ct := convertHumanCommonType(d)
			anonNS.CommonTypes = append(anonNS.CommonTypes, ct)
		}
	}

	// Only include anonymous namespace if it has content
	if len(anonNS.EntityTypes) > 0 || len(anonNS.Actions) > 0 || len(anonNS.CommonTypes) > 0 {
		namespaces = append([]*Namespace{anonNS}, namespaces...)
	}

	return namespaces
}

func convertHumanNamespace(n *ast.Namespace) *Namespace {
	ns := &Namespace{
		Name:        types.Path(n.Name.String()),
		Annotations: convertHumanAnnotations(n.Annotations),
	}

	for _, decl := range n.Decls {
		switch d := decl.(type) {
		case *ast.Entity:
			entities := convertHumanEntity(d)
			for _, e := range entities {
				ns.EntityTypes = append(ns.EntityTypes, e)
			}
		case *ast.Action:
			actions := convertHumanAction(d)
			for _, a := range actions {
				ns.Actions = append(ns.Actions, a)
			}
		case *ast.CommonTypeDecl:
			ct := convertHumanCommonType(d)
			ns.CommonTypes = append(ns.CommonTypes, ct)
		}
	}

	return ns
}

func convertHumanEntity(e *ast.Entity) []*EntityType {
	// Handle enum entities
	if len(e.Enum) > 0 {
		var entities []*EntityType
		for _, name := range e.Names {
			et := &EntityType{
				Name:        types.Ident(name.Value),
				Annotations: convertHumanAnnotations(e.Annotations),
			}
			// Note: enum entities are a different concept - we'll handle them as regular entities for now
			entities = append(entities, et)
		}
		return entities
	}

	var entities []*EntityType
	for _, name := range e.Names {
		et := &EntityType{
			Name:        types.Ident(name.Value),
			Annotations: convertHumanAnnotations(e.Annotations),
		}

		// Convert memberOf
		for _, path := range e.In {
			et.MemberOf = append(et.MemberOf, types.Path(path.String()))
		}

		// Convert shape
		if e.Shape != nil {
			shape := convertHumanRecordType(e.Shape)
			et.Shape = &shape
		}

		// Convert tags
		if e.Tags != nil {
			et.Tags = convertHumanType(e.Tags)
		}

		entities = append(entities, et)
	}

	return entities
}

func convertHumanAction(a *ast.Action) []*Action {
	var actions []*Action

	for _, name := range a.Names {
		act := &Action{
			Name:        types.Ident(name.String()),
			Annotations: convertHumanAnnotations(a.Annotations),
		}

		// Convert memberOf
		for _, ref := range a.In {
			actRef := ActionRef{
				Name: types.String(ref.Name.String()),
			}
			if len(ref.Namespace) > 0 {
				parts := make([]string, len(ref.Namespace))
				for i, ident := range ref.Namespace {
					parts[i] = ident.Value
				}
				actRef.Namespace = types.Path(joinPath(parts))
			}
			act.MemberOf = append(act.MemberOf, actRef)
		}

		// Convert appliesTo
		if a.AppliesTo != nil {
			appliesTo := AppliesTo{}

			for _, p := range a.AppliesTo.Principal {
				appliesTo.PrincipalTypes = append(appliesTo.PrincipalTypes, types.Path(p.String()))
			}

			for _, r := range a.AppliesTo.Resource {
				appliesTo.ResourceTypes = append(appliesTo.ResourceTypes, types.Path(r.String()))
			}

			if a.AppliesTo.ContextRecord != nil {
				rec := convertHumanRecordType(a.AppliesTo.ContextRecord)
				appliesTo.Context = rec
			} else if a.AppliesTo.ContextPath != nil {
				appliesTo.Context = TypeName{Name: types.Path(a.AppliesTo.ContextPath.String())}
			}

			act.AppliesTo = &appliesTo
		}

		actions = append(actions, act)
	}

	return actions
}

func convertHumanCommonType(c *ast.CommonTypeDecl) *CommonType {
	return &CommonType{
		Name:        types.Ident(c.Name.Value),
		Annotations: convertHumanAnnotations(c.Annotations),
		Type:        convertHumanType(c.Value),
	}
}

func convertHumanType(t ast.Type) Type {
	switch t := t.(type) {
	case *ast.Path:
		name := t.String()
		switch name {
		case "String":
			return TypeString{}
		case "Long":
			return TypeLong{}
		case "Bool", "Boolean":
			return TypeBoolean{}
		default:
			return TypeName{Name: types.Path(name)}
		}
	case *ast.SetType:
		return TypeSet{Element: convertHumanType(t.Element)}
	case *ast.RecordType:
		return convertHumanRecordType(t)
	default:
		panic(fmt.Sprintf("unexpected type: %T", t))
	}
}

func convertHumanRecordType(r *ast.RecordType) TypeRecord {
	rec := TypeRecord{}
	for _, attr := range r.Attributes {
		rec.Attributes = append(rec.Attributes, Attribute{
			Name:        types.Ident(attr.Key.String()),
			Type:        convertHumanType(attr.Type),
			Required:    attr.IsRequired,
			Annotations: convertHumanAnnotations(attr.Annotations),
		})
	}
	return rec
}

func convertHumanAnnotations(anns []*ast.Annotation) Annotations {
	var result Annotations
	for _, a := range anns {
		ann := Annotation{Key: types.Ident(a.Key.Value)}
		if a.Value != nil {
			ann.Value = types.String(a.Value.Value())
		}
		result = append(result, ann)
	}
	return result
}

// convertToHumanAST converts our schema types to the internal AST format.
func convertToHumanAST(s *Schema) *ast.Schema {
	h := &ast.Schema{}

	for _, ns := range s.Namespaces {
		if ns.Name == "" {
			// Anonymous namespace - add declarations directly
			for _, ct := range ns.CommonTypes {
				h.Decls = append(h.Decls, convertToHumanCommonType(ct))
			}
			for _, et := range ns.EntityTypes {
				h.Decls = append(h.Decls, convertToHumanEntity(et))
			}
			for _, a := range ns.Actions {
				h.Decls = append(h.Decls, convertToHumanActionDecl(a))
			}
		} else {
			h.Decls = append(h.Decls, convertToHumanNamespace(ns))
		}
	}

	return h
}

func convertToHumanNamespace(n *Namespace) *ast.Namespace {
	ns := &ast.Namespace{
		Name:        pathToAST(string(n.Name)),
		Annotations: convertToHumanAnnotations(n.Annotations),
	}

	for _, ct := range n.CommonTypes {
		ns.Decls = append(ns.Decls, convertToHumanCommonType(ct))
	}
	for _, et := range n.EntityTypes {
		ns.Decls = append(ns.Decls, convertToHumanEntity(et))
	}
	for _, a := range n.Actions {
		ns.Decls = append(ns.Decls, convertToHumanActionDecl(a))
	}

	return ns
}

func convertToHumanEntity(e *EntityType) *ast.Entity {
	entity := &ast.Entity{
		Names:       []*ast.Ident{{Value: string(e.Name)}},
		Annotations: convertToHumanAnnotations(e.Annotations),
	}

	for _, m := range e.MemberOf {
		entity.In = append(entity.In, pathToAST(string(m)))
	}

	if e.Shape != nil {
		entity.Shape = convertToHumanRecordType(*e.Shape)
	}

	if e.Tags != nil {
		entity.Tags = convertToHumanTypeAST(e.Tags)
	}

	return entity
}

func convertToHumanActionDecl(a *Action) *ast.Action {
	action := &ast.Action{
		Names:       []ast.Name{&ast.Ident{Value: string(a.Name)}},
		Annotations: convertToHumanAnnotations(a.Annotations),
	}

	for _, ref := range a.MemberOf {
		r := &ast.Ref{
			Name: &ast.String{QuotedVal: fmt.Sprintf("%q", ref.Name)},
		}
		if ref.Namespace != "" {
			parts := splitPath(string(ref.Namespace))
			for _, p := range parts {
				r.Namespace = append(r.Namespace, &ast.Ident{Value: p})
			}
		}
		action.In = append(action.In, r)
	}

	if a.AppliesTo != nil {
		action.AppliesTo = &ast.AppliesTo{}
		for _, p := range a.AppliesTo.PrincipalTypes {
			action.AppliesTo.Principal = append(action.AppliesTo.Principal, pathToAST(string(p)))
		}
		for _, r := range a.AppliesTo.ResourceTypes {
			action.AppliesTo.Resource = append(action.AppliesTo.Resource, pathToAST(string(r)))
		}
		if a.AppliesTo.Context != nil {
			if rec, ok := a.AppliesTo.Context.(TypeRecord); ok {
				action.AppliesTo.ContextRecord = convertToHumanRecordType(rec)
			} else if name, ok := a.AppliesTo.Context.(TypeName); ok {
				action.AppliesTo.ContextPath = pathToAST(string(name.Name))
			}
		}
	}

	return action
}

func convertToHumanCommonType(c *CommonType) *ast.CommonTypeDecl {
	return &ast.CommonTypeDecl{
		Name:        &ast.Ident{Value: string(c.Name)},
		Value:       convertToHumanTypeAST(c.Type),
		Annotations: convertToHumanAnnotations(c.Annotations),
	}
}

func convertToHumanTypeAST(t Type) ast.Type {
	switch t := t.(type) {
	case TypeString:
		return &ast.Path{Parts: []*ast.Ident{{Value: "String"}}}
	case TypeLong:
		return &ast.Path{Parts: []*ast.Ident{{Value: "Long"}}}
	case TypeBoolean:
		return &ast.Path{Parts: []*ast.Ident{{Value: "Bool"}}}
	case TypeSet:
		return &ast.SetType{Element: convertToHumanTypeAST(t.Element)}
	case TypeRecord:
		return convertToHumanRecordType(t)
	case TypeName:
		return pathToAST(string(t.Name))
	case TypeExtension:
		return pathToAST(string(t.Name))
	default:
		panic(fmt.Sprintf("unexpected type: %T", t))
	}
}

func convertToHumanRecordType(r TypeRecord) *ast.RecordType {
	rec := &ast.RecordType{}
	for _, attr := range r.Attributes {
		rec.Attributes = append(rec.Attributes, &ast.Attribute{
			Key:         &ast.Ident{Value: string(attr.Name)},
			Type:        convertToHumanTypeAST(attr.Type),
			IsRequired:  attr.Required,
			Annotations: convertToHumanAnnotations(attr.Annotations),
		})
	}
	return rec
}

func convertToHumanAnnotations(anns Annotations) []*ast.Annotation {
	var result []*ast.Annotation
	for _, a := range anns {
		ann := &ast.Annotation{
			Key: &ast.Ident{Value: string(a.Key)},
		}
		if a.Value != "" {
			ann.Value = &ast.String{QuotedVal: fmt.Sprintf("%q", a.Value)}
		}
		result = append(result, ann)
	}
	return result
}

// convertFromJSONSchema converts a JSON schema to our new types.
func convertFromJSONSchema(js ast.JSONSchema) []*Namespace {
	var namespaces []*Namespace

	for name, jsNS := range js {
		ns := &Namespace{
			Name:        types.Path(name),
			Annotations: convertJSONAnnotations(jsNS.Annotations),
		}

		// Convert common types
		for ctName, ct := range jsNS.CommonTypes {
			ns.CommonTypes = append(ns.CommonTypes, &CommonType{
				Name:        types.Ident(ctName),
				Annotations: convertJSONAnnotations(ct.Annotations),
				Type:        convertJSONType(ct.JSONType),
			})
		}

		// Convert entity types
		for etName, et := range jsNS.EntityTypes {
			entity := &EntityType{
				Name:        types.Ident(etName),
				Annotations: convertJSONAnnotations(et.Annotations),
			}

			for _, m := range et.MemberOfTypes {
				entity.MemberOf = append(entity.MemberOf, types.Path(m))
			}

			if et.Shape != nil {
				shape := convertJSONRecordType(et.Shape)
				entity.Shape = &shape
			}

			if et.Tags != nil {
				entity.Tags = convertJSONType(et.Tags)
			}

			ns.EntityTypes = append(ns.EntityTypes, entity)
		}

		// Convert actions
		for actName, act := range jsNS.Actions {
			action := &Action{
				Name:        types.Ident(actName),
				Annotations: convertJSONAnnotations(act.Annotations),
			}

			for _, m := range act.MemberOf {
				ref := ActionRef{
					Name: types.String(m.ID),
				}
				if m.Type != "" {
					ref.Namespace = types.Path(m.Type)
				}
				action.MemberOf = append(action.MemberOf, ref)
			}

			if act.AppliesTo != nil {
				appliesTo := AppliesTo{}
				for _, p := range act.AppliesTo.PrincipalTypes {
					appliesTo.PrincipalTypes = append(appliesTo.PrincipalTypes, types.Path(p))
				}
				for _, r := range act.AppliesTo.ResourceTypes {
					appliesTo.ResourceTypes = append(appliesTo.ResourceTypes, types.Path(r))
				}
				if act.AppliesTo.Context != nil {
					appliesTo.Context = convertJSONType(act.AppliesTo.Context)
				}
				action.AppliesTo = &appliesTo
			}

			ns.Actions = append(ns.Actions, action)
		}

		namespaces = append(namespaces, ns)
	}

	return namespaces
}

func convertJSONType(jt *ast.JSONType) Type {
	if jt == nil {
		return nil
	}

	switch jt.Type {
	case "String":
		return TypeString{}
	case "Long":
		return TypeLong{}
	case "Boolean":
		return TypeBoolean{}
	case "Set":
		return TypeSet{Element: convertJSONType(jt.Element)}
	case "Record":
		return convertJSONRecordType(jt)
	case "Entity", "EntityOrCommon":
		return TypeName{Name: types.Path(jt.Name)}
	case "Extension":
		return TypeExtension{Name: types.Path(jt.Name)}
	default:
		// Unknown type - treat as entity/common reference
		return TypeName{Name: types.Path(jt.Type)}
	}
}

func convertJSONRecordType(jt *ast.JSONType) TypeRecord {
	rec := TypeRecord{}
	for name, attr := range jt.Attributes {
		rec.Attributes = append(rec.Attributes, Attribute{
			Name:        types.Ident(name),
			Type:        convertJSONAttrType(attr),
			Required:    attr.Required,
			Annotations: convertJSONAnnotations(attr.Annotations),
		})
	}
	return rec
}

func convertJSONAttrType(attr *ast.JSONAttribute) Type {
	switch attr.Type {
	case "String":
		return TypeString{}
	case "Long":
		return TypeLong{}
	case "Boolean":
		return TypeBoolean{}
	case "Set":
		return TypeSet{Element: convertJSONType(attr.Element)}
	case "Record":
		return convertJSONAttrRecord(attr)
	case "Entity", "EntityOrCommon":
		return TypeName{Name: types.Path(attr.Name)}
	case "Extension":
		return TypeExtension{Name: types.Path(attr.Name)}
	default:
		return TypeName{Name: types.Path(attr.Type)}
	}
}

func convertJSONAttrRecord(attr *ast.JSONAttribute) TypeRecord {
	rec := TypeRecord{}
	for name, inner := range attr.Attributes {
		rec.Attributes = append(rec.Attributes, Attribute{
			Name:        types.Ident(name),
			Type:        convertJSONAttrType(inner),
			Required:    inner.Required,
			Annotations: convertJSONAnnotations(inner.Annotations),
		})
	}
	return rec
}

func convertJSONAnnotations(anns map[string]string) Annotations {
	var result Annotations
	for k, v := range anns {
		result = append(result, Annotation{Key: types.Ident(k), Value: types.String(v)})
	}
	return result
}

// convertToJSONSchema converts our schema types to JSON schema format.
func convertToJSONSchema(s *Schema) ast.JSONSchema {
	js := make(ast.JSONSchema)

	for _, ns := range s.Namespaces {
		jsNS := &ast.JSONNamespace{
			EntityTypes: make(map[string]*ast.JSONEntity),
			Actions:     make(map[string]*ast.JSONAction),
			CommonTypes: make(map[string]*ast.JSONCommonType),
			Annotations: make(map[string]string),
		}

		for _, ann := range ns.Annotations {
			jsNS.Annotations[string(ann.Key)] = string(ann.Value)
		}

		for _, ct := range ns.CommonTypes {
			jsNS.CommonTypes[string(ct.Name)] = &ast.JSONCommonType{
				JSONType: convertToJSONType(ct.Type),
			}
			for _, ann := range ct.Annotations {
				jsNS.CommonTypes[string(ct.Name)].Annotations[string(ann.Key)] = string(ann.Value)
			}
		}

		for _, et := range ns.EntityTypes {
			jsEntity := &ast.JSONEntity{
				Annotations: make(map[string]string),
			}

			for _, ann := range et.Annotations {
				jsEntity.Annotations[string(ann.Key)] = string(ann.Value)
			}

			for _, m := range et.MemberOf {
				jsEntity.MemberOfTypes = append(jsEntity.MemberOfTypes, string(m))
			}

			if et.Shape != nil {
				jsEntity.Shape = convertToJSONType(*et.Shape)
			}

			if et.Tags != nil {
				jsEntity.Tags = convertToJSONType(et.Tags)
			}

			jsNS.EntityTypes[string(et.Name)] = jsEntity
		}

		for _, act := range ns.Actions {
			jsAction := &ast.JSONAction{
				Annotations: make(map[string]string),
			}

			for _, ann := range act.Annotations {
				jsAction.Annotations[string(ann.Key)] = string(ann.Value)
			}

			for _, m := range act.MemberOf {
				jsAction.MemberOf = append(jsAction.MemberOf, &ast.JSONMember{
					ID:   string(m.Name),
					Type: string(m.Namespace),
				})
			}

			if act.AppliesTo != nil {
				jsAction.AppliesTo = &ast.JSONAppliesTo{}
				for _, p := range act.AppliesTo.PrincipalTypes {
					jsAction.AppliesTo.PrincipalTypes = append(jsAction.AppliesTo.PrincipalTypes, string(p))
				}
				for _, r := range act.AppliesTo.ResourceTypes {
					jsAction.AppliesTo.ResourceTypes = append(jsAction.AppliesTo.ResourceTypes, string(r))
				}
				if act.AppliesTo.Context != nil {
					jsAction.AppliesTo.Context = convertToJSONType(act.AppliesTo.Context)
				}
			}

			jsNS.Actions[string(act.Name)] = jsAction
		}

		js[string(ns.Name)] = jsNS
	}

	return js
}

func convertToJSONType(t Type) *ast.JSONType {
	if t == nil {
		return nil
	}

	switch t := t.(type) {
	case TypeString:
		return &ast.JSONType{Type: "String"}
	case TypeLong:
		return &ast.JSONType{Type: "Long"}
	case TypeBoolean:
		return &ast.JSONType{Type: "Boolean"}
	case TypeSet:
		return &ast.JSONType{Type: "Set", Element: convertToJSONType(t.Element)}
	case TypeRecord:
		return convertToJSONRecordType(t)
	case TypeName:
		return &ast.JSONType{Type: "EntityOrCommon", Name: string(t.Name)}
	case TypeExtension:
		return &ast.JSONType{Type: "Extension", Name: string(t.Name)}
	default:
		panic(fmt.Sprintf("unexpected type: %T", t))
	}
}

func convertToJSONRecordType(r TypeRecord) *ast.JSONType {
	jt := &ast.JSONType{
		Type:       "Record",
		Attributes: make(map[string]*ast.JSONAttribute),
	}

	for _, attr := range r.Attributes {
		jsAttr := &ast.JSONAttribute{
			Required:    attr.Required,
			Annotations: make(map[string]string),
		}

		for _, ann := range attr.Annotations {
			jsAttr.Annotations[string(ann.Key)] = string(ann.Value)
		}

		switch t := attr.Type.(type) {
		case TypeString:
			jsAttr.Type = "String"
		case TypeLong:
			jsAttr.Type = "Long"
		case TypeBoolean:
			jsAttr.Type = "Boolean"
		case TypeSet:
			jsAttr.Type = "Set"
			jsAttr.Element = convertToJSONType(t.Element)
		case TypeRecord:
			jsAttr.Type = "Record"
			rec := convertToJSONRecordType(t)
			jsAttr.Attributes = rec.Attributes
		case TypeName:
			jsAttr.Type = "EntityOrCommon"
			jsAttr.Name = string(t.Name)
		case TypeExtension:
			jsAttr.Type = "Extension"
			jsAttr.Name = string(t.Name)
		}

		jt.Attributes[string(attr.Name)] = jsAttr
	}

	return jt
}

// Helper functions

func pathToAST(s string) *ast.Path {
	parts := splitPath(s)
	path := &ast.Path{}
	for _, p := range parts {
		path.Parts = append(path.Parts, &ast.Ident{Value: p})
	}
	return path
}

func splitPath(s string) []string {
	if s == "" {
		return nil
	}
	var parts []string
	current := ""
	for i := 0; i < len(s); i++ {
		if i+1 < len(s) && s[i] == ':' && s[i+1] == ':' {
			parts = append(parts, current)
			current = ""
			i++ // skip second colon
		} else {
			current += string(s[i])
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

func joinPath(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "::"
		}
		result += p
	}
	return result
}
