package schema

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	internalast "github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/types"
)

// MarshalCedar serializes the schema to Cedar human-readable format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	var buf bytes.Buffer
	if err := s.WriteCedar(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// WriteCedar writes the schema to the writer in Cedar human-readable format.
func (s *Schema) WriteCedar(w io.Writer) error {
	f := &cedarFormatter{w: w, lastChar: '\n', tab: "  "}

	// Sort namespaces for deterministic output
	// Process anonymous namespace first (empty string), then others alphabetically
	nsNames := make([]string, 0, len(s.Namespaces))
	for name := range s.Namespaces {
		nsNames = append(nsNames, string(name))
	}
	sort.Strings(nsNames)

	// Move anonymous namespace to the front if it exists
	for i, name := range nsNames {
		if name == "" {
			nsNames = append([]string{""}, append(nsNames[:i], nsNames[i+1:]...)...)
			break
		}
	}

	for _, nsName := range nsNames {
		ns := s.Namespaces[types.Path(nsName)]
		if err := f.writeNamespace(ns); err != nil {
			return err
		}
	}

	return nil
}

// UnmarshalCedar parses the schema from Cedar human-readable format.
// If SetFilename was called, that filename will be used for error messages.
func (s *Schema) UnmarshalCedar(data []byte) error {
	return s.UnmarshalCedarWithFilename(s.filename, data)
}

// UnmarshalCedarWithFilename parses the schema from Cedar human-readable format
// with the given filename for error messages.
func (s *Schema) UnmarshalCedarWithFilename(filename string, data []byte) error {
	savedFilename := s.filename
	internalSchema, err := parser.ParseFile(filename, data)
	if err != nil {
		return err
	}

	return s.unmarshalFromInternal(savedFilename, internalSchema)
}

// unmarshalFromInternal converts an internal schema to a Schema.
// This is separated from UnmarshalCedarWithFilename for testability.
func (s *Schema) unmarshalFromInternal(savedFilename string, internalSchema *internalast.Schema) error {
	result := New()
	if err := convertInternalSchema(result, internalSchema); err != nil {
		return err
	}
	result.filename = savedFilename
	*s = *result
	return nil
}

// cedarFormatter formats a schema to Cedar human-readable format.
type cedarFormatter struct {
	w        io.Writer
	indent   int
	lastChar byte
	tab      string
}

func (f *cedarFormatter) write(s string) error {
	if len(s) == 0 {
		return nil
	}
	_, err := io.WriteString(f.w, s)
	if err != nil {
		return err
	}
	f.lastChar = s[len(s)-1]
	return nil
}

func (f *cedarFormatter) writef(format string, args ...any) error {
	return f.write(fmt.Sprintf(format, args...))
}

func (f *cedarFormatter) writeIndent(s string) error {
	for range f.indent {
		if err := f.write(f.tab); err != nil {
			return err
		}
	}
	return f.write(s)
}

func (f *cedarFormatter) writeIndentf(format string, args ...any) error {
	return f.writeIndent(fmt.Sprintf(format, args...))
}

func (f *cedarFormatter) writeNamespace(ns *Namespace) error {
	// Write annotations
	for _, ann := range ns.Annotations {
		if err := f.writeAnnotation(ann); err != nil {
			return err
		}
	}

	isAnonymous := ns.Name == ""

	if !isAnonymous {
		if err := f.writeIndentf("namespace %s {\n", ns.Name); err != nil {
			return err
		}
		f.indent++
	}

	// Write common types first (sorted)
	ctNames := make([]string, 0, len(ns.CommonTypes))
	for name := range ns.CommonTypes {
		ctNames = append(ctNames, string(name))
	}
	sort.Strings(ctNames)
	for _, ctName := range ctNames {
		ct := ns.CommonTypes[types.Ident(ctName)]
		if err := f.writeCommonType(ct); err != nil {
			return err
		}
	}

	// Write entities (sorted)
	entityNames := make([]string, 0, len(ns.Entities))
	for name := range ns.Entities {
		entityNames = append(entityNames, string(name))
	}
	sort.Strings(entityNames)
	for _, entityName := range entityNames {
		entity := ns.Entities[types.Ident(entityName)]
		if err := f.writeEntity(entity); err != nil {
			return err
		}
	}

	// Write actions (sorted)
	actionNames := make([]string, 0, len(ns.Actions))
	for name := range ns.Actions {
		actionNames = append(actionNames, string(name))
	}
	sort.Strings(actionNames)
	for _, actionName := range actionNames {
		action := ns.Actions[types.String(actionName)]
		if err := f.writeAction(action); err != nil {
			return err
		}
	}

	if !isAnonymous {
		f.indent--
		if err := f.writeIndent("}\n"); err != nil {
			return err
		}
	}

	return nil
}

func (f *cedarFormatter) writeAnnotation(ann Annotation) error {
	if err := f.writeIndentf("@%s", ann.Key); err != nil {
		return err
	}
	if ann.Value != "" {
		if err := f.writef("(%s)", strconv.Quote(string(ann.Value))); err != nil {
			return err
		}
	}
	return f.write("\n")
}

func (f *cedarFormatter) writeCommonType(ct *CommonTypeDecl) error {
	for _, ann := range ct.Annotations {
		if err := f.writeAnnotation(ann); err != nil {
			return err
		}
	}
	if err := f.writeIndentf("type %s = ", ct.Name); err != nil {
		return err
	}
	if err := f.writeType(ct.Type); err != nil {
		return err
	}
	return f.write(";\n")
}

func (f *cedarFormatter) writeEntity(e *EntityDecl) error {
	for _, ann := range e.Annotations {
		if err := f.writeAnnotation(ann); err != nil {
			return err
		}
	}
	if err := f.writeIndentf("entity %s", e.Name); err != nil {
		return err
	}

	if len(e.Enum) > 0 {
		// Enum entity
		if err := f.write(" enum ["); err != nil {
			return err
		}
		for i, val := range e.Enum {
			if i > 0 {
				if err := f.write(", "); err != nil {
					return err
				}
			}
			if err := f.write(strconv.Quote(string(val))); err != nil {
				return err
			}
		}
		if err := f.write("]"); err != nil {
			return err
		}
	} else {
		// Regular entity
		if len(e.MemberOfTypes) > 0 {
			if err := f.write(" in "); err != nil {
				return err
			}
			if err := f.writePathList(e.MemberOfTypes); err != nil {
				return err
			}
		}

		if len(e.Attributes) > 0 {
			if err := f.write(" "); err != nil {
				return err
			}
			if err := f.writeRecordType(e.Attributes); err != nil {
				return err
			}
		}

		if e.Tags.v != nil {
			if err := f.write(" tags "); err != nil {
				return err
			}
			if err := f.writeType(e.Tags); err != nil {
				return err
			}
		}
	}

	return f.write(";\n")
}

func (f *cedarFormatter) writeAction(a *ActionDecl) error {
	for _, ann := range a.Annotations {
		if err := f.writeAnnotation(ann); err != nil {
			return err
		}
	}

	// Determine if name needs quotes
	name := string(a.Name)
	needsQuotes := !isValidIdent(name)
	if needsQuotes {
		if err := f.writeIndentf("action %s", strconv.Quote(name)); err != nil {
			return err
		}
	} else {
		if err := f.writeIndentf("action %s", name); err != nil {
			return err
		}
	}

	if len(a.MemberOf) > 0 {
		if err := f.write(" in "); err != nil {
			return err
		}
		if len(a.MemberOf) > 1 {
			if err := f.write("["); err != nil {
				return err
			}
		}
		for i, ref := range a.MemberOf {
			if i > 0 {
				if err := f.write(", "); err != nil {
					return err
				}
			}
			if err := f.writeActionRef(ref); err != nil {
				return err
			}
		}
		if len(a.MemberOf) > 1 {
			if err := f.write("]"); err != nil {
				return err
			}
		}
	}

	if len(a.PrincipalTypes) > 0 || len(a.ResourceTypes) > 0 || a.Context.v != nil {
		if err := f.write(" appliesTo {\n"); err != nil {
			return err
		}
		f.indent++

		if len(a.PrincipalTypes) > 0 {
			if err := f.writeIndent("principal: "); err != nil {
				return err
			}
			if err := f.writePathList(a.PrincipalTypes); err != nil {
				return err
			}
			if err := f.write(",\n"); err != nil {
				return err
			}
		}

		if len(a.ResourceTypes) > 0 {
			if err := f.writeIndent("resource: "); err != nil {
				return err
			}
			if err := f.writePathList(a.ResourceTypes); err != nil {
				return err
			}
			if err := f.write(",\n"); err != nil {
				return err
			}
		}

		if a.Context.v != nil {
			if err := f.writeIndent("context: "); err != nil {
				return err
			}
			if err := f.writeType(a.Context); err != nil {
				return err
			}
			if err := f.write(",\n"); err != nil {
				return err
			}
		}

		f.indent--
		if err := f.writeIndent("}"); err != nil {
			return err
		}
	}

	return f.write(";\n")
}

func (f *cedarFormatter) writeActionRef(ref ActionRef) error {
	if ref.Namespace != "" {
		if err := f.writef("%s::", ref.Namespace); err != nil {
			return err
		}
	}
	name := string(ref.Name)
	needsQuotes := !isValidIdent(name)
	if needsQuotes {
		return f.write(strconv.Quote(name))
	}
	return f.write(name)
}

func (f *cedarFormatter) writePathList(paths []types.Path) error {
	if len(paths) > 1 {
		if err := f.write("["); err != nil {
			return err
		}
	}
	for i, path := range paths {
		if i > 0 {
			if err := f.write(", "); err != nil {
				return err
			}
		}
		if err := f.write(string(path)); err != nil {
			return err
		}
	}
	if len(paths) > 1 {
		if err := f.write("]"); err != nil {
			return err
		}
	}
	return nil
}

func (f *cedarFormatter) writeType(t Type) error {
	if t.v == nil {
		return nil
	}

	switch v := t.v.(type) {
	case TypeBoolean:
		return f.write("Bool")
	case TypeLong:
		return f.write("Long")
	case TypeString:
		return f.write("String")
	case TypeSet:
		if err := f.write("Set<"); err != nil {
			return err
		}
		if err := f.writeType(v.Element); err != nil {
			return err
		}
		return f.write(">")
	case TypeRecord:
		return f.writeRecordType(v.Attributes)
	case TypeEntity:
		return f.write(string(v.Name))
	case TypeExtension:
		return f.write(string(v.Name))
	case TypeRef:
		return f.write(string(v.Name))
	default:
		return fmt.Errorf("unknown type: %T", v)
	}
}

func (f *cedarFormatter) writeRecordType(attrs []Attribute) error {
	if err := f.write("{\n"); err != nil {
		return err
	}
	f.indent++
	for _, attr := range attrs {
		if err := f.writeIndent(""); err != nil {
			return err
		}
		// Attribute name - quote if necessary
		name := string(attr.Name)
		needsQuotes := !isValidIdent(name)
		if needsQuotes {
			if err := f.write(strconv.Quote(name)); err != nil {
				return err
			}
		} else {
			if err := f.write(name); err != nil {
				return err
			}
		}
		if !attr.Required {
			if err := f.write("?"); err != nil {
				return err
			}
		}
		if err := f.write(": "); err != nil {
			return err
		}
		if err := f.writeType(attr.Type); err != nil {
			return err
		}
		if err := f.write(",\n"); err != nil {
			return err
		}
	}
	f.indent--
	return f.writeIndent("}")
}

// isValidIdent checks if a string is a valid Cedar identifier
func isValidIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_') {
				return false
			}
		}
	}
	return true
}

// Conversion from internal schema AST to our schema types

func convertInternalSchema(s *Schema, internal *internalast.Schema) error {
	// The internal schema has a flat list of declarations
	// Some are namespaced, some are in the anonymous namespace

	for _, decl := range internal.Decls {
		switch d := decl.(type) {
		case *internalast.Namespace:
			ns, err := convertInternalNamespace(d)
			if err != nil {
				return err
			}
			s.Namespaces[ns.Name] = ns
		case *internalast.Entity:
			// Anonymous namespace entity
			ns := s.ensureNamespace("")
			entities, err := convertInternalEntity(d)
			if err != nil {
				return err
			}
			for _, e := range entities {
				ns.Entities[e.Name] = e
			}
		case *internalast.Action:
			// Anonymous namespace action
			ns := s.ensureNamespace("")
			actions, err := convertInternalAction(d)
			if err != nil {
				return err
			}
			for _, a := range actions {
				ns.Actions[a.Name] = a
			}
		case *internalast.CommonTypeDecl:
			// Anonymous namespace common type
			ns := s.ensureNamespace("")
			ct, err := convertInternalCommonType(d)
			if err != nil {
				return err
			}
			ns.CommonTypes[ct.Name] = ct
		case *internalast.CommentBlock:
			// Ignore comments
		}
	}

	return nil
}

// getAnnotationValue safely gets the annotation value, returning empty string if nil.
func getAnnotationValue(ann *internalast.Annotation) types.String {
	if ann.Value == nil {
		return ""
	}
	return types.String(ann.Value.String())
}

func convertInternalNamespace(internal *internalast.Namespace) (*Namespace, error) {
	ns := NewNamespace(types.Path(internal.Name.String()))

	// Convert annotations
	for _, ann := range internal.Annotations {
		ns.Annotations = ns.Annotations.Set(
			types.Ident(ann.Key.String()),
			getAnnotationValue(ann),
		)
	}

	for _, decl := range internal.Decls {
		switch d := decl.(type) {
		case *internalast.Entity:
			entities, err := convertInternalEntity(d)
			if err != nil {
				return nil, err
			}
			for _, e := range entities {
				ns.Entities[e.Name] = e
			}
		case *internalast.Action:
			actions, err := convertInternalAction(d)
			if err != nil {
				return nil, err
			}
			for _, a := range actions {
				ns.Actions[a.Name] = a
			}
		case *internalast.CommonTypeDecl:
			ct, err := convertInternalCommonType(d)
			if err != nil {
				return nil, err
			}
			ns.CommonTypes[ct.Name] = ct
		case *internalast.CommentBlock:
			// Ignore comments
		}
	}

	return ns, nil
}

func convertInternalEntity(internal *internalast.Entity) ([]*EntityDecl, error) {
	// Internal entity can define multiple entities with the same shape
	var entities []*EntityDecl

	for _, name := range internal.Names {
		entity := NewEntity(types.Ident(name.String()))

		// Convert annotations
		for _, ann := range internal.Annotations {
			entity.Annotations = entity.Annotations.Set(
				types.Ident(ann.Key.String()),
				getAnnotationValue(ann),
			)
		}

		// Convert memberOf
		for _, path := range internal.In {
			entity.MemberOfTypes = append(entity.MemberOfTypes, types.Path(path.String()))
		}

		// Convert shape
		if internal.Shape != nil {
			attrs, err := convertInternalRecordAttrs(internal.Shape)
			if err != nil {
				return nil, err
			}
			entity.Attributes = attrs
		}

		// Convert tags
		if internal.Tags != nil {
			t, err := convertInternalType(internal.Tags)
			if err != nil {
				return nil, err
			}
			entity.Tags = t
		}

		// Convert enum
		for _, val := range internal.Enum {
			entity.Enum = append(entity.Enum, types.String(val.String()))
		}

		entities = append(entities, entity)
	}

	return entities, nil
}

func convertInternalAction(internal *internalast.Action) ([]*ActionDecl, error) {
	// Internal action can define multiple actions
	var actions []*ActionDecl

	for _, name := range internal.Names {
		action := NewAction(types.String(name.String()))

		// Convert annotations
		for _, ann := range internal.Annotations {
			action.Annotations = action.Annotations.Set(
				types.Ident(ann.Key.String()),
				getAnnotationValue(ann),
			)
		}

		// Convert memberOf
		for _, ref := range internal.In {
			actionRef := ActionRef{
				Name: types.String(ref.Name.String()),
			}
			if len(ref.Namespace) > 0 {
				var nsParts []string
				for _, part := range ref.Namespace {
					nsParts = append(nsParts, part.String())
				}
				actionRef.Namespace = types.Path(strings.Join(nsParts, "::"))
			}
			action.MemberOf = append(action.MemberOf, actionRef)
		}

		// Convert appliesTo
		if internal.AppliesTo != nil {
			for _, path := range internal.AppliesTo.Principal {
				action.PrincipalTypes = append(action.PrincipalTypes, types.Path(path.String()))
			}
			for _, path := range internal.AppliesTo.Resource {
				action.ResourceTypes = append(action.ResourceTypes, types.Path(path.String()))
			}
			if internal.AppliesTo.ContextRecord != nil {
				attrs, err := convertInternalRecordAttrs(internal.AppliesTo.ContextRecord)
				if err != nil {
					return nil, err
				}
				action.Context = Record(attrs...)
			} else if internal.AppliesTo.ContextPath != nil {
				action.Context = Ref(types.Path(internal.AppliesTo.ContextPath.String()))
			}
		}

		actions = append(actions, action)
	}

	return actions, nil
}

func convertInternalCommonType(internal *internalast.CommonTypeDecl) (*CommonTypeDecl, error) {
	t, err := convertInternalType(internal.Value)
	if err != nil {
		return nil, err
	}

	ct := NewCommonType(types.Ident(internal.Name.String()), t)

	// Convert annotations
	for _, ann := range internal.Annotations {
		ct.Annotations = ct.Annotations.Set(
			types.Ident(ann.Key.String()),
			getAnnotationValue(ann),
		)
	}

	return ct, nil
}

func convertInternalType(internal internalast.Type) (Type, error) {
	if internal == nil {
		return Type{}, nil
	}

	switch t := internal.(type) {
	case *internalast.Path:
		name := t.String()
		switch name {
		case "Bool", "Boolean":
			return Boolean(), nil
		case "Long":
			return Long(), nil
		case "String":
			return String(), nil
		default:
			// Could be an entity reference or common type reference
			return Ref(types.Path(name)), nil
		}
	case *internalast.SetType:
		elem, err := convertInternalType(t.Element)
		if err != nil {
			return Type{}, err
		}
		return SetOf(elem), nil
	case *internalast.RecordType:
		attrs, err := convertInternalRecordAttrs(t)
		if err != nil {
			return Type{}, err
		}
		return Record(attrs...), nil
	default:
		return Type{}, fmt.Errorf("unknown internal type: %T", t)
	}
}

func convertInternalRecordAttrs(internal *internalast.RecordType) ([]Attribute, error) {
	var attrs []Attribute
	for _, attr := range internal.Attributes {
		t, err := convertInternalType(attr.Type)
		if err != nil {
			return nil, err
		}
		attrs = append(attrs, Attribute{
			Name:     types.Ident(attr.Key.String()),
			Type:     t,
			Required: attr.IsRequired,
		})
	}
	return attrs, nil
}
