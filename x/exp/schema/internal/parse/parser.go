// Package parse provides the internal Cedar schema text parser.
package parse

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cedar-policy/cedar-go/internal/rust"
)

// ParseError provides details about a parse error.
type ParseError struct {
	Filename string
	Line     int
	Column   int
	Message  string
}

func (e *ParseError) Error() string {
	if e.Filename != "" {
		return fmt.Sprintf("parse error: %s:%d:%d: %s", e.Filename, e.Line, e.Column, e.Message)
	}
	if e.Line > 0 {
		return fmt.Sprintf("parse error: line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("parse error: %s", e.Message)
}

// Annotations are key-value metadata attached to schema elements.
type Annotations map[string]string

// Schema represents a parsed but unresolved Cedar schema.
type Schema struct {
	Namespaces map[string]*Namespace
}

// Namespace contains entity types, actions, common types, and enum types within a namespace.
type Namespace struct {
	EntityTypes map[string]*EntityTypeDef
	EnumTypes   map[string]*EnumTypeDef
	Actions     map[string]*ActionDef
	CommonTypes map[string]*CommonTypeDef
	Annotations Annotations
}

// EntityTypeDef describes an entity type in the schema.
type EntityTypeDef struct {
	MemberOfTypes []string
	Shape         *RecordType
	Tags          Type
	Annotations   Annotations
}

// EnumTypeDef describes an enumerated entity type in the schema.
type EnumTypeDef struct {
	Values      []string
	Annotations Annotations
}

// ActionDef describes an action in the schema.
type ActionDef struct {
	MemberOf    []*ActionRef
	AppliesTo   *AppliesTo
	Annotations Annotations
}

// ActionRef references an action, possibly in another namespace.
type ActionRef struct {
	Type string
	ID   string
}

// AppliesTo specifies what principals and resources an action applies to.
type AppliesTo struct {
	PrincipalTypes []string
	ResourceTypes  []string
	Context        *RecordType
	ContextRef     Type
}

// CommonTypeDef is a named type alias.
type CommonTypeDef struct {
	Type        Type
	Annotations Annotations
}

// Type represents a Cedar type in the schema.
type Type interface {
	isType()
}

// PrimitiveKind represents Cedar primitive types.
type PrimitiveKind int

const (
	PrimitiveLong PrimitiveKind = iota
	PrimitiveString
	PrimitiveBool
)

// PrimitiveType represents Long, String, or Bool.
type PrimitiveType struct {
	Kind PrimitiveKind
}

// isType is a marker method that indicates PrimitiveType implements Type.
func (PrimitiveType) isType() { _ = struct{}{} }

// SetType represents Set<T>.
type SetType struct {
	Element Type
}

// isType is a marker method that indicates SetType implements Type.
func (SetType) isType() { _ = struct{}{} }

// RecordType represents a record with named attributes.
type RecordType struct {
	Attributes map[string]*Attribute
}

// isType is a marker method that indicates RecordType implements Type.
func (*RecordType) isType() { _ = struct{}{} }

// Attribute is a named field in a record type.
type Attribute struct {
	Type        Type
	Required    bool
	Annotations Annotations
}

// EntityRef references an entity type by name.
type EntityRef struct {
	Name string
}

// isType is a marker method that indicates EntityRef implements Type.
func (EntityRef) isType() { _ = struct{}{} }

// ExtensionType represents an extension type like ipaddr or decimal.
type ExtensionType struct {
	Name string
}

// isType is a marker method that indicates ExtensionType implements Type.
func (ExtensionType) isType() { _ = struct{}{} }

// CommonTypeRef references a common type by name.
type CommonTypeRef struct {
	Name string
}

// isType is a marker method that indicates CommonTypeRef implements Type.
func (CommonTypeRef) isType() { _ = struct{}{} }

// EntityOrCommonRef is an ambiguous reference that could be an entity or common type.
type EntityOrCommonRef struct {
	Name string
}

// isType is a marker method that indicates EntityOrCommonRef implements Type.
func (EntityOrCommonRef) isType() { _ = struct{}{} }

// primitiveTypeNames are names that cannot be used as entity or common type names
// because they are reserved for Cedar's built-in primitive types.
var primitiveTypeNames = map[string]bool{
	"Bool":      true,
	"Boolean":   true,
	"Entity":    true,
	"Extension": true,
	"Long":      true,
	"Record":    true,
	"Set":       true,
	"String":    true,
}

// IsPrimitiveTypeName returns true if name is a built-in primitive type name
// (e.g., Bool, Long, String, Entity, etc.) that cannot be used as a custom type name.
func IsPrimitiveTypeName(name string) bool {
	return primitiveTypeNames[name]
}

// ReservedNameError provides details about use of a reserved name.
type ReservedNameError struct {
	Name string
	Kind string
}

func (e *ReservedNameError) Error() string {
	return fmt.Sprintf("reserved name: %q cannot be used as %s", e.Name, e.Kind)
}

// Parser is the Cedar schema text parser.
type Parser struct {
	src      []byte
	pos      int
	line     int
	col      int
	filename string
}

// New creates a new parser for the given source.
func New(src []byte, filename string) *Parser {
	return &Parser{
		src:      src,
		filename: filename,
		line:     1,
		col:      1,
	}
}

// Parse parses a Cedar schema from Cedar text format.
func (p *Parser) Parse() (*Schema, error) {
	p.line = 1
	p.col = 1

	schema := &Schema{
		Namespaces: make(map[string]*Namespace),
	}

	for {
		p.skipWhitespaceAndComments()
		if p.pos >= len(p.src) {
			break
		}

		// Check for annotations
		annotations := make(Annotations)
		for p.peek() == '@' {
			key, value, err := p.parseAnnotation()
			if err != nil {
				return nil, err
			}
			annotations[key] = value
			p.skipWhitespaceAndComments()
		}

		tok := p.peekToken()

		switch tok {
		case "namespace":
			ns, nsName, err := p.parseNamespace(annotations)
			if err != nil {
				return nil, err
			}
			if _, exists := schema.Namespaces[nsName]; exists {
				return nil, p.error("duplicate namespace %q", nsName)
			}
			schema.Namespaces[nsName] = ns

		case "entity", "action", "type":
			// Declaration in empty namespace
			ns, ok := schema.Namespaces[""]
			if !ok {
				ns = &Namespace{
					EntityTypes: make(map[string]*EntityTypeDef),
					EnumTypes:   make(map[string]*EnumTypeDef),
					Actions:     make(map[string]*ActionDef),
					CommonTypes: make(map[string]*CommonTypeDef),
					Annotations: make(Annotations),
				}
				schema.Namespaces[""] = ns
			}

			if err := p.parseDeclaration(ns, annotations); err != nil {
				return nil, err
			}

		default:
			return nil, p.error("expected 'namespace', 'entity', 'action', or 'type', got %q", tok)
		}
	}

	return schema, nil
}

func (p *Parser) parseNamespace(annotations Annotations) (*Namespace, string, error) {
	if err := p.expect("namespace"); err != nil {
		return nil, "", err
	}

	name, err := p.parsePath()
	if err != nil {
		return nil, "", err
	}

	if err := p.expect("{"); err != nil {
		return nil, "", err
	}

	ns := &Namespace{
		EntityTypes: make(map[string]*EntityTypeDef),
		EnumTypes:   make(map[string]*EnumTypeDef),
		Actions:     make(map[string]*ActionDef),
		CommonTypes: make(map[string]*CommonTypeDef),
		Annotations: annotations,
	}

	for {
		p.skipWhitespaceAndComments()
		if p.peek() == '}' {
			p.advance() // consume the closing brace we just confirmed is there
			break
		}

		// Check for annotations
		declAnnotations := make(Annotations)
		for p.peek() == '@' {
			key, value, err := p.parseAnnotation()
			if err != nil {
				return nil, "", err
			}
			declAnnotations[key] = value
			p.skipWhitespaceAndComments()
		}

		if err := p.parseDeclaration(ns, declAnnotations); err != nil {
			return nil, "", err
		}
	}

	return ns, name, nil
}

func (p *Parser) parseDeclaration(ns *Namespace, annotations Annotations) error {
	tok := p.peekToken()

	switch tok {
	case "entity":
		return p.parseEntity(ns, annotations)
	case "action":
		return p.parseAction(ns, annotations)
	case "type":
		return p.parseCommonType(ns, annotations)
	default:
		return p.error("expected 'entity', 'action', or 'type', got %q", tok)
	}
}

func (p *Parser) parseEntity(ns *Namespace, annotations Annotations) error {
	if err := p.expect("entity"); err != nil {
		return err
	}

	names, err := p.parseIdentList()
	if err != nil {
		return err
	}

	for _, name := range names {
		if IsPrimitiveTypeName(name) {
			return &ReservedNameError{Name: name, Kind: "entity type"}
		}
		if _, exists := ns.EntityTypes[name]; exists {
			return p.error("duplicate entity type %q", name)
		}
		if _, exists := ns.EnumTypes[name]; exists {
			return p.error("duplicate entity type %q", name)
		}
	}

	// Check for enum
	p.skipWhitespaceAndComments()
	if p.peekToken() == "enum" {
		p.consumeToken()
		enumVals, err := p.parseStringList()
		if err != nil {
			return err
		}
		if err := p.expect(";"); err != nil {
			return err
		}

		for _, name := range names {
			ns.EnumTypes[name] = &EnumTypeDef{
				Values:      enumVals,
				Annotations: copyAnnotations(annotations),
			}
		}
		return nil
	}

	var memberOf []string
	if p.peekToken() == "in" {
		p.consumeToken()
		memberOf, err = p.parseTypeList()
		if err != nil {
			return err
		}
	}

	var shape *RecordType
	p.skipWhitespaceAndComments()
	if p.peek() == '=' {
		p.advance()
		p.skipWhitespaceAndComments()
	}
	if p.peek() == '{' {
		shape, err = p.parseRecordType()
		if err != nil {
			return err
		}
	}

	var tags Type
	p.skipWhitespaceAndComments()
	if p.peekToken() == "tags" {
		p.consumeToken()
		tags, err = p.parseType()
		if err != nil {
			return err
		}
	}

	if err := p.expect(";"); err != nil {
		return err
	}

	for _, name := range names {
		ns.EntityTypes[name] = &EntityTypeDef{
			MemberOfTypes: memberOf,
			Shape:         shape,
			Tags:          tags,
			Annotations:   copyAnnotations(annotations),
		}
	}

	return nil
}

func (p *Parser) parseAction(ns *Namespace, annotations Annotations) error {
	if err := p.expect("action"); err != nil {
		return err
	}

	names, err := p.parseNameList()
	if err != nil {
		return err
	}

	for _, name := range names {
		if _, exists := ns.Actions[name]; exists {
			return p.error("duplicate action %q", name)
		}
	}

	var memberOf []*ActionRef
	p.skipWhitespaceAndComments()
	if p.peekToken() == "in" {
		p.consumeToken()
		memberOf, err = p.parseActionRefList()
		if err != nil {
			return err
		}
	}

	var appliesTo *AppliesTo
	p.skipWhitespaceAndComments()
	if p.peekToken() == "appliesTo" {
		p.consumeToken()
		appliesTo, err = p.parseAppliesTo()
		if err != nil {
			return err
		}
	}

	if err := p.expect(";"); err != nil {
		return err
	}

	for _, name := range names {
		ns.Actions[name] = &ActionDef{
			MemberOf:    memberOf,
			AppliesTo:   appliesTo,
			Annotations: copyAnnotations(annotations),
		}
	}

	return nil
}

func (p *Parser) parseCommonType(ns *Namespace, annotations Annotations) error {
	if err := p.expect("type"); err != nil {
		return err
	}

	name, err := p.parseIdent()
	if err != nil {
		return err
	}

	if IsPrimitiveTypeName(name) {
		return &ReservedNameError{Name: name, Kind: "common type"}
	}

	if _, exists := ns.CommonTypes[name]; exists {
		return p.error("duplicate common type %q", name)
	}

	if err := p.expect("="); err != nil {
		return err
	}

	typ, err := p.parseType()
	if err != nil {
		return err
	}

	if err := p.expect(";"); err != nil {
		return err
	}

	ns.CommonTypes[name] = &CommonTypeDef{
		Type:        typ,
		Annotations: annotations,
	}

	return nil
}

func (p *Parser) parseAppliesTo() (*AppliesTo, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}

	at := &AppliesTo{}

	for {
		p.skipWhitespaceAndComments()
		if p.peek() == '}' {
			p.advance() // consume the closing brace we just confirmed is there
			break
		}

		key, err := p.parseIdent()
		if err != nil {
			return nil, err
		}

		if err := p.expect(":"); err != nil {
			return nil, err
		}

		switch key {
		case "principal":
			types, err := p.parseTypeList()
			if err != nil {
				return nil, err
			}
			at.PrincipalTypes = types
		case "resource":
			types, err := p.parseTypeList()
			if err != nil {
				return nil, err
			}
			at.ResourceTypes = types
		case "context":
			// Can be a record type or a type reference
			p.skipWhitespaceAndComments()
			if p.peek() == '{' {
				rt, err := p.parseRecordType()
				if err != nil {
					return nil, err
				}
				at.Context = rt
			} else {
				// Type reference - store in ContextRef
				typ, err := p.parseType()
				if err != nil {
					return nil, err
				}
				at.ContextRef = typ
			}
		default:
			return nil, p.error("unexpected key in appliesTo: %q", key)
		}

		// Optional trailing comma
		p.skipWhitespaceAndComments()
		if p.peek() == ',' {
			p.advance()
		}
	}

	return at, nil
}

func (p *Parser) parseType() (Type, error) {
	p.skipWhitespaceAndComments()

	// Check for Set<...>
	if p.peekToken() == "Set" {
		p.consumeToken()
		if err := p.expect("<"); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if err := p.expect(">"); err != nil {
			return nil, err
		}
		return SetType{Element: elem}, nil
	}

	// Check for record type { ... }
	if p.peek() == '{' {
		return p.parseRecordType()
	}

	// Otherwise it's a type name (primitive, entity, extension, or common type)
	name, err := p.parsePath()
	if err != nil {
		return nil, err
	}

	// Check for primitives
	switch name {
	case "Long":
		return PrimitiveType{Kind: PrimitiveLong}, nil
	case "String":
		return PrimitiveType{Kind: PrimitiveString}, nil
	case "Bool":
		return PrimitiveType{Kind: PrimitiveBool}, nil
	}

	// Check for known extension types
	switch name {
	case "ipaddr", "decimal", "datetime", "duration":
		return ExtensionType{Name: name}, nil
	case "__cedar::ipaddr", "__cedar::decimal", "__cedar::datetime", "__cedar::duration":
		return ExtensionType{Name: strings.TrimPrefix(name, "__cedar::")}, nil
	case "__cedar::Long":
		return PrimitiveType{Kind: PrimitiveLong}, nil
	case "__cedar::String":
		return PrimitiveType{Kind: PrimitiveString}, nil
	case "__cedar::Bool":
		return PrimitiveType{Kind: PrimitiveBool}, nil
	}

	// Could be entity type or common type - ambiguous until resolution
	return EntityOrCommonRef{Name: name}, nil
}

func (p *Parser) parseRecordType() (*RecordType, error) {
	if err := p.expect("{"); err != nil {
		return nil, err
	}

	rt := &RecordType{
		Attributes: make(map[string]*Attribute),
	}

	for {
		p.skipWhitespaceAndComments()
		if p.peek() == '}' {
			p.advance() // consume the closing brace we just confirmed is there
			break
		}

		// Check for annotations
		annotations := make(Annotations)
		for p.peek() == '@' {
			key, value, err := p.parseAnnotation()
			if err != nil {
				return nil, err
			}
			annotations[key] = value
			p.skipWhitespaceAndComments()
		}

		// Parse attribute name (can be ident or string)
		name, err := p.parseName()
		if err != nil {
			return nil, err
		}

		// Check for optional marker
		required := true
		p.skipWhitespaceAndComments()
		if p.peek() == '?' {
			p.advance()
			required = false
		}

		if err := p.expect(":"); err != nil {
			return nil, err
		}

		typ, err := p.parseType()
		if err != nil {
			return nil, err
		}

		rt.Attributes[name] = &Attribute{
			Type:        typ,
			Required:    required,
			Annotations: annotations,
		}

		// Optional trailing comma
		p.skipWhitespaceAndComments()
		if p.peek() == ',' {
			p.advance()
		}
	}

	return rt, nil
}

func (p *Parser) parseAnnotation() (string, string, error) {
	if err := p.expect("@"); err != nil {
		return "", "", err
	}

	key, err := p.parseIdent()
	if err != nil {
		return "", "", err
	}

	p.skipWhitespaceAndComments()
	if p.peek() != '(' {
		return key, "", nil
	}

	p.advance() // consume (

	value, err := p.parseString()
	if err != nil {
		return "", "", err
	}

	if err := p.expect(")"); err != nil {
		return "", "", err
	}

	return key, value, nil
}

func (p *Parser) parseIdentList() ([]string, error) {
	var names []string

	name, err := p.parseIdent()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	for {
		p.skipWhitespaceAndComments()
		if p.peek() != ',' {
			break
		}
		p.advance()

		// Check if next token is still an ident (not 'in', 'enum', etc.)
		p.skipWhitespaceAndComments()
		tok := p.peekToken()
		if tok == "in" || tok == "enum" {
			// Put back the comma conceptually - but we've already consumed it
			// This is fine, we just stop parsing the list
			break
		}

		name, err = p.parseIdent()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (p *Parser) parseNameList() ([]string, error) {
	var names []string

	name, err := p.parseName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	for {
		p.skipWhitespaceAndComments()
		if p.peek() != ',' {
			break
		}
		p.advance()

		// Check if next token indicates end of name list
		p.skipWhitespaceAndComments()
		tok := p.peekToken()
		if tok == "in" || tok == "appliesTo" {
			break
		}

		name, err = p.parseName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	return names, nil
}

func (p *Parser) parseTypeList() ([]string, error) {
	p.skipWhitespaceAndComments()

	// Can be single type or [type1, type2, ...]
	if p.peek() == '[' {
		p.advance()
		var types []string
		for {
			p.skipWhitespaceAndComments()
			if p.peek() == ']' {
				p.advance() // consume the closing bracket we just confirmed is there
				break
			}

			typ, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			types = append(types, typ)

			p.skipWhitespaceAndComments()
			if p.peek() == ',' {
				p.advance()
			}
		}
		return types, nil
	}

	// Single type
	typ, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	return []string{typ}, nil
}

func (p *Parser) parseActionRefList() ([]*ActionRef, error) {
	p.skipWhitespaceAndComments()

	// Can be single ref or [ref1, ref2, ...]
	if p.peek() == '[' {
		p.advance()
		var refs []*ActionRef
		for {
			p.skipWhitespaceAndComments()
			if p.peek() == ']' {
				p.advance() // consume the closing bracket we just confirmed is there
				break
			}

			ref, err := p.parseActionRef()
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)

			p.skipWhitespaceAndComments()
			if p.peek() == ',' {
				p.advance()
			}
		}
		return refs, nil
	}

	// Single ref
	ref, err := p.parseActionRef()
	if err != nil {
		return nil, err
	}
	return []*ActionRef{ref}, nil
}

func (p *Parser) parseActionRef() (*ActionRef, error) {
	p.skipWhitespaceAndComments()

	if p.peek() == '"' {
		// Just a string ID
		id, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &ActionRef{ID: id}, nil
	}

	// Parse path
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}

	p.skipWhitespaceAndComments()
	if p.hasPrefix("::\"") {
		p.advance() // skip first :
		p.advance() // skip second :
		id, err := p.parseString()
		if err != nil {
			return nil, err
		}
		return &ActionRef{Type: path, ID: id}, nil
	}

	// Just an identifier
	return &ActionRef{ID: path}, nil
}

func (p *Parser) parseStringList() ([]string, error) {
	if err := p.expect("["); err != nil {
		return nil, err
	}

	var strs []string
	for {
		p.skipWhitespaceAndComments()
		if p.peek() == ']' {
			p.advance() // consume the closing bracket we just confirmed is there
			break
		}

		s, err := p.parseString()
		if err != nil {
			return nil, err
		}
		strs = append(strs, s)

		p.skipWhitespaceAndComments()
		if p.peek() == ',' {
			p.advance()
		}
	}

	return strs, nil
}

func (p *Parser) parseIdent() (string, error) {
	p.skipWhitespaceAndComments()

	start := p.pos
	if p.pos >= len(p.src) {
		return "", p.error("expected identifier, got EOF")
	}

	r, size := utf8.DecodeRune(p.src[p.pos:])
	if !unicode.IsLetter(r) && r != '_' {
		return "", p.error("expected identifier, got %q", string(r))
	}
	p.pos += size
	p.col++

	for p.pos < len(p.src) {
		r, size = utf8.DecodeRune(p.src[p.pos:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		p.pos += size
		p.col++
	}

	return string(p.src[start:p.pos]), nil
}

func (p *Parser) parseName() (string, error) {
	p.skipWhitespaceAndComments()

	if p.peek() == '"' {
		return p.parseString()
	}
	return p.parseIdent()
}

func (p *Parser) parsePath() (string, error) {
	var parts []string

	part, err := p.parseIdent()
	if err != nil {
		return "", err
	}
	parts = append(parts, part)

	for p.hasPrefix("::") && !p.hasPrefix("::\"") {
		// Skip :: but not ::\" (action entity UID)
		p.advance()
		p.advance()

		part, err = p.parseIdent()
		if err != nil {
			return "", err
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "::"), nil
}

func (p *Parser) parseString() (string, error) {
	if err := p.expect("\""); err != nil {
		return "", err
	}

	start := p.pos
	for p.pos < len(p.src) {
		if p.src[p.pos] == '"' {
			break
		}
		if p.src[p.pos] == '\\' && p.pos+1 < len(p.src) {
			p.pos += 2 // skip escape sequence
			p.col += 2
			continue
		}
		if p.src[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}

	content := p.src[start:p.pos]
	if err := p.expect("\""); err != nil {
		return "", err
	}

	unescaped, _, err := rust.Unquote(content, false)
	if err != nil {
		return "", p.error("invalid string: %v", err)
	}

	return unescaped, nil
}

func (p *Parser) expect(s string) error {
	p.skipWhitespaceAndComments()

	if !p.hasPrefix(s) {
		if p.pos >= len(p.src) {
			return p.error("expected %q, got EOF", s)
		}
		return p.error("expected %q, got %q", s, string(p.src[p.pos]))
	}

	for range s {
		if p.src[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}

	return nil
}

func (p *Parser) hasPrefix(s string) bool {
	return bytes.HasPrefix(p.src[p.pos:], []byte(s))
}

func (p *Parser) peek() byte {
	if p.pos >= len(p.src) {
		return 0
	}
	return p.src[p.pos]
}

func (p *Parser) advance() {
	if p.pos < len(p.src) {
		if p.src[p.pos] == '\n' {
			p.line++
			p.col = 1
		} else {
			p.col++
		}
		p.pos++
	}
}

func (p *Parser) peekToken() string {
	savedPos := p.pos
	savedLine := p.line
	savedCol := p.col

	p.skipWhitespaceAndComments()

	if p.pos >= len(p.src) {
		p.pos = savedPos
		p.line = savedLine
		p.col = savedCol
		return ""
	}

	start := p.pos
	r, size := utf8.DecodeRune(p.src[p.pos:])
	if !unicode.IsLetter(r) && r != '_' {
		p.pos = savedPos
		p.line = savedLine
		p.col = savedCol
		return ""
	}

	p.pos += size
	for p.pos < len(p.src) {
		r, size = utf8.DecodeRune(p.src[p.pos:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		p.pos += size
	}

	tok := string(p.src[start:p.pos])

	p.pos = savedPos
	p.line = savedLine
	p.col = savedCol

	return tok
}

func (p *Parser) consumeToken() {
	p.skipWhitespaceAndComments()

	if p.pos >= len(p.src) {
		return
	}

	r, size := utf8.DecodeRune(p.src[p.pos:])
	if !unicode.IsLetter(r) && r != '_' {
		return
	}

	p.pos += size
	p.col++

	for p.pos < len(p.src) {
		r, size = utf8.DecodeRune(p.src[p.pos:])
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			break
		}
		p.pos += size
		p.col++
	}
}

func (p *Parser) skipWhitespaceAndComments() {
	for p.pos < len(p.src) {
		// Skip whitespace
		if p.src[p.pos] == ' ' || p.src[p.pos] == '\t' || p.src[p.pos] == '\r' {
			p.pos++
			p.col++
			continue
		}

		if p.src[p.pos] == '\n' {
			p.pos++
			p.line++
			p.col = 1
			continue
		}

		// Skip // comments
		if p.hasPrefix("//") {
			p.pos += 2
			for p.pos < len(p.src) && p.src[p.pos] != '\n' {
				p.pos++
			}
			continue
		}

		// Skip /* */ comments
		if p.hasPrefix("/*") {
			p.pos += 2
			for p.pos < len(p.src)-1 {
				if p.src[p.pos] == '*' && p.src[p.pos+1] == '/' {
					p.pos += 2
					break
				}
				if p.src[p.pos] == '\n' {
					p.line++
					p.col = 1
				} else {
					p.col++
				}
				p.pos++
			}
			continue
		}

		break
	}
}

func (p *Parser) error(format string, args ...any) error {
	return &ParseError{
		Filename: p.filename,
		Line:     p.line,
		Column:   p.col,
		Message:  fmt.Sprintf(format, args...),
	}
}

func copyAnnotations(a Annotations) Annotations {
	if a == nil {
		return make(Annotations)
	}
	c := make(Annotations, len(a))
	for k, v := range a {
		c[k] = v
	}
	return c
}
