package schema

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strings"

	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/scan"
)

// UnmarshalCedar parses a Cedar schema from Cedar text format.
func (s *Schema) UnmarshalCedar(data []byte) error {
	p := &parser{scanner: scan.New(data), filename: s.filename}
	if err := p.advance(); err != nil {
		return err
	}
	schema, err := p.parseSchema()
	if err != nil {
		return err
	}
	s.Namespaces = schema.Namespaces
	return nil
}

// MarshalCedar serializes a Schema to Cedar text format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	var b strings.Builder
	nsNames := sortedKeys(s.Namespaces)
	first := true
	for _, nsName := range nsNames {
		ns := s.Namespaces[nsName]
		if !first {
			b.WriteByte('\n')
		}
		first = false
		if nsName == "" {
			writeNamespaceContents(&b, ns, "")
		} else {
			writeAnnotations(&b, ns.Annotations, "")
			fmt.Fprintf(&b, "namespace %s {\n", nsName)
			writeNamespaceContents(&b, ns, "  ")
			b.WriteString("}\n")
		}
	}
	return []byte(b.String()), nil
}

func writeNamespaceContents(b *strings.Builder, ns *Namespace, indent string) {
	entityNames := sortedKeys(ns.EntityTypes)
	for _, name := range entityNames {
		et := ns.EntityTypes[name]
		writeAnnotations(b, et.Annotations, indent)
		fmt.Fprintf(b, "%sentity %s", indent, name)
		if len(et.MemberOfTypes) > 0 {
			writeEntityParents(b, et.MemberOfTypes)
		}
		if et.Shape != nil && len(et.Shape.Attributes) > 0 {
			b.WriteString(" ")
			writeRecordType(b, et.Shape, indent)
		}
		if et.Tags != nil {
			b.WriteString(" tags ")
			writeTypeExpr(b, et.Tags)
		}
		b.WriteString(";\n")
	}

	enumNames := sortedKeys(ns.EnumTypes)
	for _, name := range enumNames {
		enum := ns.EnumTypes[name]
		writeAnnotations(b, enum.Annotations, indent)
		fmt.Fprintf(b, "%sentity %s enum [", indent, name)
		for i, v := range enum.Values {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(scan.Quote(v))
		}
		b.WriteString("];\n")
	}

	commonNames := sortedKeys(ns.CommonTypes)
	for _, name := range commonNames {
		ct := ns.CommonTypes[name]
		writeAnnotations(b, ct.Annotations, indent)
		fmt.Fprintf(b, "%stype %s = ", indent, name)
		writeTypeExpr(b, ct.Type)
		b.WriteString(";\n")
	}

	actionNames := sortedKeys(ns.Actions)
	for _, name := range actionNames {
		act := ns.Actions[name]
		writeAnnotations(b, act.Annotations, indent)
		fmt.Fprintf(b, "%saction %s", indent, quoteActionName(name))
		if len(act.MemberOf) > 0 {
			b.WriteString(" in [")
			for i, ref := range act.MemberOf {
				if i > 0 {
					b.WriteString(", ")
				}
				if ref.Type != "" {
					fmt.Fprintf(b, "%s::%s", ref.Type, scan.Quote(ref.ID))
				} else {
					b.WriteString(scan.Quote(ref.ID))
				}
			}
			b.WriteString("]")
		}
		if act.AppliesTo != nil {
			writeAppliesTo(b, act.AppliesTo, indent)
		}
		b.WriteString(";\n")
	}
}

func writeEntityParents(b *strings.Builder, parents []string) {
	if len(parents) == 1 {
		fmt.Fprintf(b, " in %s", parents[0])
	} else {
		b.WriteString(" in [")
		for i, p := range parents {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(p)
		}
		b.WriteString("]")
	}
}

func writeAppliesTo(b *strings.Builder, at *AppliesTo, indent string) {
	b.WriteString(" appliesTo {\n")
	inner := indent + "  "
	if len(at.PrincipalTypes) > 0 {
		fmt.Fprintf(b, "%sprincipal: ", inner)
		writeEntityTypeList(b, at.PrincipalTypes)
		b.WriteString(",\n")
	}
	if len(at.ResourceTypes) > 0 {
		fmt.Fprintf(b, "%sresource: ", inner)
		writeEntityTypeList(b, at.ResourceTypes)
		b.WriteString(",\n")
	}
	if at.Context != nil {
		fmt.Fprintf(b, "%scontext: ", inner)
		writeTypeExpr(b, at.Context)
		b.WriteString(",\n")
	}
	fmt.Fprintf(b, "%s}", indent)
}

func writeEntityTypeList(b *strings.Builder, types []string) {
	if len(types) == 1 {
		b.WriteString(types[0])
	} else {
		b.WriteString("[")
		for i, t := range types {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(t)
		}
		b.WriteString("]")
	}
}

func writeTypeExpr(b *strings.Builder, t TypeExpr) {
	switch v := t.(type) {
	case PrimitiveTypeExpr:
		b.WriteString(v.Kind.String())
	case SetTypeExpr:
		b.WriteString("Set<")
		writeTypeExpr(b, v.Element)
		b.WriteString(">")
	case *RecordTypeExpr:
		writeRecordType(b, v, "")
	case EntityRefExpr:
		b.WriteString(v.Name)
	case ExtensionTypeExpr:
		b.WriteString(v.Name)
	case TypeNameExpr:
		b.WriteString(v.Name)
	case EntityNameExpr:
		b.WriteString(v.Name)
	}
}

func writeRecordType(b *strings.Builder, rt *RecordTypeExpr, indent string) {
	if len(rt.Attributes) == 0 {
		b.WriteString("{}")
		return
	}
	b.WriteString("{\n")
	inner := indent + "  "
	names := sortedKeys(rt.Attributes)
	for _, name := range names {
		attr := rt.Attributes[name]
		writeAnnotations(b, attr.Annotations, inner)
		fmt.Fprintf(b, "%s%s", inner, quoteAttrName(name))
		if !attr.Required {
			b.WriteString("?")
		}
		b.WriteString(": ")
		writeTypeExpr(b, attr.Type)
		b.WriteString(",\n")
	}
	fmt.Fprintf(b, "%s}", indent)
}

func writeAnnotations(b *strings.Builder, ann Annotations, indent string) {
	if len(ann) == 0 {
		return
	}
	keys := sortedKeys(ann)
	for _, key := range keys {
		if v := ann[key]; v != "" {
			fmt.Fprintf(b, "%s@%s(%s)\n", indent, key, scan.Quote(v))
		} else {
			fmt.Fprintf(b, "%s@%s\n", indent, key)
		}
	}
}

func quoteActionName(name string) string {
	if isIdent(name) {
		return name
	}
	return scan.Quote(name)
}

func quoteAttrName(name string) string {
	if isIdent(name) {
		return name
	}
	return scan.Quote(name)
}

func isIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 {
			if ch != '_' && !isLetter(ch) {
				return false
			}
		} else {
			if ch != '_' && !isLetter(ch) && !isDigit(ch) {
				return false
			}
		}
	}
	return true
}

func isLetter(ch rune) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') }
func isDigit(ch rune) bool  { return ch >= '0' && ch <= '9' }

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Parser

type parser struct {
	scanner  *scan.Scanner
	tok      scan.Token
	filename string
}

func (p *parser) advance() error {
	tok, err := p.scanner.Next()
	if err != nil {
		return p.wrapErr(err.Error())
	}
	p.tok = tok
	return nil
}

func (p *parser) expect(kind scan.TokenKind) (scan.Token, error) {
	if p.tok.Kind != kind {
		return scan.Token{}, p.errorf("expected %s, got %s", kind, p.tok.Kind)
	}
	tok := p.tok
	if err := p.advance(); err != nil {
		return scan.Token{}, err
	}
	return tok, nil
}

func (p *parser) errorf(format string, args ...any) error {
	return &ParseError{
		Filename: p.filename,
		Line:     p.tok.Pos.Line,
		Column:   p.tok.Pos.Column,
		Message:  fmt.Sprintf(format, args...),
	}
}

func (p *parser) wrapErr(msg string) error {
	return &ParseError{Filename: p.filename, Message: msg}
}

func (p *parser) parseSchema() (*Schema, error) {
	s := &Schema{Namespaces: make(map[string]*Namespace)}
	for p.tok.Kind != scan.TokenEOF {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		if p.tok.Kind == scan.TokenIdent && p.tok.Text == "namespace" {
			if err := p.advance(); err != nil {
				return nil, err
			}
			nsName, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(scan.TokenLBrace); err != nil {
				return nil, err
			}
			ns, err := p.parseNamespaceBody()
			if err != nil {
				return nil, err
			}
			ns.Annotations = annotations
			if _, exists := s.Namespaces[nsName]; exists {
				return nil, p.errorf("duplicate namespace %q", nsName)
			}
			s.Namespaces[nsName] = ns
			if _, err := p.expect(scan.TokenRBrace); err != nil {
				return nil, err
			}
		} else {
			ns := s.Namespaces[""]
			if ns == nil {
				ns = newNamespace()
				s.Namespaces[""] = ns
			}
			if err := p.parseDecl(ns, annotations); err != nil {
				return nil, err
			}
		}
	}
	return s, nil
}

func (p *parser) parseNamespaceBody() (*Namespace, error) {
	ns := newNamespace()
	for p.tok.Kind != scan.TokenRBrace && p.tok.Kind != scan.TokenEOF {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		if err := p.parseDecl(ns, annotations); err != nil {
			return nil, err
		}
	}
	return ns, nil
}

func (p *parser) parseDecl(ns *Namespace, annotations Annotations) error {
	if p.tok.Kind != scan.TokenIdent {
		return p.errorf("expected declaration, got %s", p.tok.Kind)
	}
	switch p.tok.Text {
	case "entity":
		return p.parseEntity(ns, annotations)
	case "action":
		return p.parseAction(ns, annotations)
	case "type":
		return p.parseTypeDecl(ns, annotations)
	default:
		return p.errorf("expected 'entity', 'action', or 'type', got %q", p.tok.Text)
	}
}

func (p *parser) parseEntity(ns *Namespace, annotations Annotations) error {
	if err := p.advance(); err != nil { // skip "entity"
		return err
	}

	names, err := p.parseIdentList()
	if err != nil {
		return err
	}

	if p.tok.Kind == scan.TokenIdent && p.tok.Text == "enum" {
		if err := p.advance(); err != nil {
			return err
		}
		values, err := p.parseStringList()
		if err != nil {
			return err
		}
		if _, err := p.expect(scan.TokenSemicolon); err != nil {
			return err
		}
		for _, name := range names {
			ns.EnumTypes[name] = &EnumTypeDef{
				Values:      slices.Clone(values),
				Annotations: cloneAnnotations(annotations),
			}
		}
		return nil
	}

	var memberOf []string
	if p.tok.Kind == scan.TokenIdent && p.tok.Text == "in" {
		if err := p.advance(); err != nil {
			return err
		}
		memberOf, err = p.parseEntityTypeList()
		if err != nil {
			return err
		}
	}

	var shape *RecordTypeExpr
	if p.tok.Kind == scan.TokenEquals {
		if err := p.advance(); err != nil {
			return err
		}
	}
	if p.tok.Kind == scan.TokenLBrace {
		shape, err = p.parseRecordType()
		if err != nil {
			return err
		}
	}

	var tags TypeExpr
	if p.tok.Kind == scan.TokenIdent && p.tok.Text == "tags" {
		if err := p.advance(); err != nil {
			return err
		}
		tags, err = p.parseType()
		if err != nil {
			return err
		}
	}

	if _, err := p.expect(scan.TokenSemicolon); err != nil {
		return err
	}

	for _, name := range names {
		et := &EntityTypeDef{
			MemberOfTypes: slices.Clone(memberOf),
			Annotations:   cloneAnnotations(annotations),
		}
		if shape != nil {
			et.Shape = cloneRecordType(shape)
		}
		if tags != nil {
			et.Tags = cloneTypeExpr(tags)
		}
		ns.EntityTypes[name] = et
	}
	return nil
}

func (p *parser) parseAction(ns *Namespace, annotations Annotations) error {
	if err := p.advance(); err != nil { // skip "action"
		return err
	}

	names, err := p.parseNameList()
	if err != nil {
		return err
	}

	var memberOf []*ActionRef
	if p.tok.Kind == scan.TokenIdent && p.tok.Text == "in" {
		if err := p.advance(); err != nil {
			return err
		}
		memberOf, err = p.parseActionRefList()
		if err != nil {
			return err
		}
	}

	var appliesTo *AppliesTo
	if p.tok.Kind == scan.TokenIdent && p.tok.Text == "appliesTo" {
		if err := p.advance(); err != nil {
			return err
		}
		appliesTo, err = p.parseAppliesTo()
		if err != nil {
			return err
		}
	}

	if _, err := p.expect(scan.TokenSemicolon); err != nil {
		return err
	}

	for _, name := range names {
		act := &ActionDef{
			Annotations: cloneAnnotations(annotations),
		}
		if len(memberOf) > 0 {
			act.MemberOf = cloneActionRefs(memberOf)
		}
		if appliesTo != nil {
			act.AppliesTo = cloneAppliesTo(appliesTo)
		}
		ns.Actions[name] = act
	}
	return nil
}

func (p *parser) parseTypeDecl(ns *Namespace, annotations Annotations) error {
	if err := p.advance(); err != nil { // skip "type"
		return err
	}
	name, err := p.expect(scan.TokenIdent)
	if err != nil {
		return err
	}
	if _, err := p.expect(scan.TokenEquals); err != nil {
		return err
	}
	typ, err := p.parseType()
	if err != nil {
		return err
	}
	if _, err := p.expect(scan.TokenSemicolon); err != nil {
		return err
	}
	ns.CommonTypes[name.Text] = &CommonTypeDef{
		Type:        typ,
		Annotations: annotations,
	}
	return nil
}

func (p *parser) parseAnnotations() (Annotations, error) {
	ann := newAnnotations()
	for p.tok.Kind == scan.TokenAt {
		if err := p.advance(); err != nil {
			return nil, err
		}
		key, err := p.expect(scan.TokenIdent)
		if err != nil {
			return nil, err
		}
		if p.tok.Kind == scan.TokenLParen {
			if err := p.advance(); err != nil {
				return nil, err
			}
			val, err := p.expect(scan.TokenString)
			if err != nil {
				return nil, err
			}
			if _, err := p.expect(scan.TokenRParen); err != nil {
				return nil, err
			}
			ann[key.Text] = val.Value
		} else {
			ann[key.Text] = ""
		}
	}
	return ann, nil
}

func (p *parser) parsePath() (string, error) {
	ident, err := p.expect(scan.TokenIdent)
	if err != nil {
		return "", err
	}
	var parts []string
	parts = append(parts, ident.Text)
	for p.tok.Kind == scan.TokenDoubleColon {
		if err := p.advance(); err != nil {
			return "", err
		}
		next, err := p.expect(scan.TokenIdent)
		if err != nil {
			return "", err
		}
		parts = append(parts, next.Text)
	}
	return strings.Join(parts, "::"), nil
}

func (p *parser) parseType() (TypeExpr, error) {
	if p.tok.Kind == scan.TokenLBrace {
		return p.parseRecordType()
	}
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	if path == "Set" {
		if _, err := p.expect(scan.TokenLAngle); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(scan.TokenRAngle); err != nil {
			return nil, err
		}
		return SetTypeExpr{Element: elem}, nil
	}
	return TypeNameExpr{Name: path}, nil
}

func (p *parser) parseRecordType() (*RecordTypeExpr, error) {
	if _, err := p.expect(scan.TokenLBrace); err != nil {
		return nil, err
	}
	rt := &RecordTypeExpr{Attributes: make(map[string]*Attribute)}
	for p.tok.Kind != scan.TokenRBrace && p.tok.Kind != scan.TokenEOF {
		ann, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}
		name, err := p.parseName()
		if err != nil {
			return nil, err
		}
		required := true
		if p.tok.Kind == scan.TokenQuestion {
			required = false
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
		if _, err := p.expect(scan.TokenColon); err != nil {
			return nil, err
		}
		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}
		rt.Attributes[name] = &Attribute{
			Type:        typ,
			Required:    required,
			Annotations: ann,
		}
		if p.tok.Kind == scan.TokenComma {
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
	}
	if _, err := p.expect(scan.TokenRBrace); err != nil {
		return nil, err
	}
	return rt, nil
}

func (p *parser) parseIdentList() ([]string, error) {
	ident, err := p.expect(scan.TokenIdent)
	if err != nil {
		return nil, err
	}
	names := []string{ident.Text}
	for p.tok.Kind == scan.TokenComma {
		if err := p.advance(); err != nil {
			return nil, err
		}
		next, err := p.expect(scan.TokenIdent)
		if err != nil {
			return nil, err
		}
		names = append(names, next.Text)
	}
	return names, nil
}

func (p *parser) parseNameList() ([]string, error) {
	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	names := []string{name}
	for p.tok.Kind == scan.TokenComma {
		if err := p.advance(); err != nil {
			return nil, err
		}
		next, err := p.parseName()
		if err != nil {
			return nil, err
		}
		names = append(names, next)
	}
	return names, nil
}

func (p *parser) parseName() (string, error) {
	if p.tok.Kind == scan.TokenString {
		tok := p.tok
		if err := p.advance(); err != nil {
			return "", err
		}
		return tok.Value, nil
	}
	tok, err := p.expect(scan.TokenIdent)
	if err != nil {
		return "", err
	}
	return tok.Text, nil
}

func (p *parser) parseEntityTypeList() ([]string, error) {
	if p.tok.Kind == scan.TokenLBracket {
		if err := p.advance(); err != nil {
			return nil, err
		}
		var types []string
		if p.tok.Kind != scan.TokenRBracket {
			path, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			types = append(types, path)
			for p.tok.Kind == scan.TokenComma {
				if err := p.advance(); err != nil {
					return nil, err
				}
				path, err := p.parsePath()
				if err != nil {
					return nil, err
				}
				types = append(types, path)
			}
		}
		if _, err := p.expect(scan.TokenRBracket); err != nil {
			return nil, err
		}
		return types, nil
	}
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	return []string{path}, nil
}

func (p *parser) parseActionRefList() ([]*ActionRef, error) {
	if p.tok.Kind == scan.TokenLBracket {
		if err := p.advance(); err != nil {
			return nil, err
		}
		var refs []*ActionRef
		if p.tok.Kind != scan.TokenRBracket {
			ref, err := p.parseActionRef()
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)
			for p.tok.Kind == scan.TokenComma {
				if err := p.advance(); err != nil {
					return nil, err
				}
				ref, err := p.parseActionRef()
				if err != nil {
					return nil, err
				}
				refs = append(refs, ref)
			}
		}
		if _, err := p.expect(scan.TokenRBracket); err != nil {
			return nil, err
		}
		return refs, nil
	}
	ref, err := p.parseActionRef()
	if err != nil {
		return nil, err
	}
	return []*ActionRef{ref}, nil
}

// parseActionRef parses: STR | Path '::' STR | Ident
func (p *parser) parseActionRef() (*ActionRef, error) {
	if p.tok.Kind == scan.TokenString {
		tok := p.tok
		if err := p.advance(); err != nil {
			return nil, err
		}
		return &ActionRef{ID: tok.Value}, nil
	}
	ident, err := p.expect(scan.TokenIdent)
	if err != nil {
		return nil, err
	}
	parts := []string{ident.Text}
	for p.tok.Kind == scan.TokenDoubleColon {
		if err := p.advance(); err != nil {
			return nil, err
		}
		if p.tok.Kind == scan.TokenString {
			str := p.tok
			if err := p.advance(); err != nil {
				return nil, err
			}
			return &ActionRef{
				Type: strings.Join(parts, "::"),
				ID:   str.Value,
			}, nil
		}
		next, err := p.expect(scan.TokenIdent)
		if err != nil {
			return nil, err
		}
		parts = append(parts, next.Text)
	}
	return &ActionRef{ID: strings.Join(parts, "::")}, nil
}

func (p *parser) parseAppliesTo() (*AppliesTo, error) {
	if _, err := p.expect(scan.TokenLBrace); err != nil {
		return nil, err
	}
	at := &AppliesTo{}
	for p.tok.Kind != scan.TokenRBrace && p.tok.Kind != scan.TokenEOF {
		if p.tok.Kind != scan.TokenIdent {
			return nil, p.errorf("expected 'principal', 'resource', or 'context', got %s", p.tok.Kind)
		}
		switch p.tok.Text {
		case "principal":
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(scan.TokenColon); err != nil {
				return nil, err
			}
			types, err := p.parseEntityTypeList()
			if err != nil {
				return nil, err
			}
			at.PrincipalTypes = types
		case "resource":
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(scan.TokenColon); err != nil {
				return nil, err
			}
			types, err := p.parseEntityTypeList()
			if err != nil {
				return nil, err
			}
			at.ResourceTypes = types
		case "context":
			if err := p.advance(); err != nil {
				return nil, err
			}
			if _, err := p.expect(scan.TokenColon); err != nil {
				return nil, err
			}
			typ, err := p.parseType()
			if err != nil {
				return nil, err
			}
			at.Context = typ
		default:
			return nil, p.errorf("expected 'principal', 'resource', or 'context', got %q", p.tok.Text)
		}
		if p.tok.Kind == scan.TokenComma {
			if err := p.advance(); err != nil {
				return nil, err
			}
		}
	}
	if _, err := p.expect(scan.TokenRBrace); err != nil {
		return nil, err
	}
	return at, nil
}

func (p *parser) parseStringList() ([]string, error) {
	if _, err := p.expect(scan.TokenLBracket); err != nil {
		return nil, err
	}
	var values []string
	if p.tok.Kind != scan.TokenRBracket {
		str, err := p.expect(scan.TokenString)
		if err != nil {
			return nil, err
		}
		values = append(values, str.Value)
		for p.tok.Kind == scan.TokenComma {
			if err := p.advance(); err != nil {
				return nil, err
			}
			if p.tok.Kind == scan.TokenRBracket {
				break
			}
			str, err := p.expect(scan.TokenString)
			if err != nil {
				return nil, err
			}
			values = append(values, str.Value)
		}
	}
	if _, err := p.expect(scan.TokenRBracket); err != nil {
		return nil, err
	}
	return values, nil
}

func cloneAnnotations(ann Annotations) Annotations {
	if len(ann) == 0 {
		return newAnnotations()
	}
	c := make(Annotations, len(ann))
	maps.Copy(c, ann)
	return c
}

func cloneRecordType(rt *RecordTypeExpr) *RecordTypeExpr {
	attrs := make(map[string]*Attribute, len(rt.Attributes))
	for k, v := range rt.Attributes {
		attrs[k] = &Attribute{
			Type:        cloneTypeExpr(v.Type),
			Required:    v.Required,
			Annotations: cloneAnnotations(v.Annotations),
		}
	}
	return &RecordTypeExpr{Attributes: attrs}
}

func cloneActionRefs(refs []*ActionRef) []*ActionRef {
	c := make([]*ActionRef, len(refs))
	for i, ref := range refs {
		c[i] = &ActionRef{Type: ref.Type, ID: ref.ID}
	}
	return c
}

func cloneAppliesTo(at *AppliesTo) *AppliesTo {
	c := &AppliesTo{
		PrincipalTypes: slices.Clone(at.PrincipalTypes),
		ResourceTypes:  slices.Clone(at.ResourceTypes),
		Context:        cloneTypeExpr(at.Context),
	}
	return c
}

func cloneTypeExpr(expr TypeExpr) TypeExpr {
	if expr == nil {
		return nil
	}
	switch v := expr.(type) {
	case *RecordTypeExpr:
		return cloneRecordType(v)
	case SetTypeExpr:
		return SetTypeExpr{Element: cloneTypeExpr(v.Element)}
	default:
		return expr
	}
}
