// Package parser provides a parser for Cedar human-readable schema format.
package parser

import (
	"fmt"
	"io"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// Parser parses Cedar schema from tokens.
type Parser struct {
	tokens []Token
	pos    int
}

// New creates a new parser for the given source.
func New(filename string, src []byte) (*Parser, error) {
	tokens, err := Tokenize(filename, src)
	if err != nil {
		return nil, err
	}
	return &Parser{tokens: tokens}, nil
}

// NewFromReader creates a new parser from an io.Reader.
func NewFromReader(filename string, r io.Reader) (*Parser, error) {
	tokens, err := TokenizeReader(filename, r)
	if err != nil {
		return nil, err
	}
	return &Parser{tokens: tokens}, nil
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) peekAhead(n int) Token {
	pos := p.pos + n - 1
	if pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

// fmtPos formats a position for error messages.
func fmtPos(pos Position) string {
	if pos.Filename != "" {
		return fmt.Sprintf("%s:%d:%d", pos.Filename, pos.Line, pos.Column)
	}
	return fmt.Sprintf("%d:%d", pos.Line, pos.Column)
}

func (p *Parser) expect(text string) (Token, error) {
	tok := p.advance()
	if tok.Text != text {
		return tok, fmt.Errorf("expected %q, got %q at %s", text, tok.Text, fmtPos(tok.Pos))
	}
	return tok, nil
}

// reservedIdents are identifiers that cannot be used as names
var reservedIdents = map[string]bool{
	"in": true, // reserved in Cedar schema
}

func (p *Parser) expectIdent() (string, error) {
	tok := p.advance()
	if tok.Type != TokenIdent {
		return "", fmt.Errorf("expected identifier, got %q at %s", tok.Text, fmtPos(tok.Pos))
	}
	if reservedIdents[tok.Text] {
		return "", fmt.Errorf("%q is a reserved identifier at %s", tok.Text, fmtPos(tok.Pos))
	}
	return tok.Text, nil
}

func (p *Parser) expectAnyIdent() (string, error) {
	tok := p.advance()
	if tok.Type != TokenIdent {
		return "", fmt.Errorf("expected identifier, got %q at %s", tok.Text, fmtPos(tok.Pos))
	}
	return tok.Text, nil
}

func (p *Parser) expectString() (string, error) {
	tok := p.advance()
	if tok.Type != TokenString {
		return "", fmt.Errorf("expected string, got %q at %s", tok.Text, fmtPos(tok.Pos))
	}
	return tok.stringValue()
}

// consumeString advances past a string token and returns its value.
// It panics if the current token is not a string - only call after checking tok.Type == TokenString.
func (p *Parser) consumeString() string {
	tok := p.advance()
	val, _ := tok.stringValue()
	return val
}

func (p *Parser) parseAnnotations() (ast.Annotations, error) {
	var annotations ast.Annotations

	// Parse annotations
	for p.peek().Text == "@" {
		ann, err := p.parseAnnotation()
		if err != nil {
			return nil, err
		}
		// Lazily initialize map on first annotation
		if annotations == nil {
			annotations = make(ast.Annotations)
		}
		// Check for duplicate annotation keys
		if _, exists := annotations[ann.Key]; exists {
			return nil, fmt.Errorf("duplicate annotation %q at %s", ann.Key, fmtPos(p.peek().Pos))
		}
		annotations[ann.Key] = ann.Value
	}
	return annotations, nil
}

// Parse parses a complete Cedar schema.
func (p *Parser) Parse() (*ast.Schema, error) {
	schema := &ast.Schema{}

	for !p.peek().isEOF() {
		annotations, err := p.parseAnnotations()
		if err != nil {
			return nil, err
		}

		// Parse declaration
		tok := p.peek()
		switch tok.Text {
		case "namespace":
			ns, name, err := p.parseNamespace(annotations)
			if err != nil {
				return nil, err
			}
			if schema.Namespaces == nil {
				schema.Namespaces = make(ast.Namespaces)
			}
			if _, exists := schema.Namespaces[name]; exists {
				return nil, fmt.Errorf("duplicate namespace %q at %s", name, fmtPos(tok.Pos))
			}
			schema.Namespaces[name] = ns
		case "entity":
			// Check if it's an enum by peeking ahead
			// After "entity", the next token is the name, and the token after that might be "enum"
			// We need to peek at position+2 (skip "entity" at pos, skip name at pos+1, check pos+2)
			if p.peekAhead(3).Text == "enum" {
				enum, name, err := p.parseEnum(annotations)
				if err != nil {
					return nil, err
				}
				if schema.Enums == nil {
					schema.Enums = make(ast.Enums)
				}
				if _, exists := schema.Enums[name]; exists {
					return nil, fmt.Errorf("duplicate enum %q at %s", name, fmtPos(tok.Pos))
				}
				if _, exists := schema.Entities[name]; exists {
					return nil, fmt.Errorf("enum %q conflicts with entity at %s", name, fmtPos(tok.Pos))
				}
				schema.Enums[name] = enum
			} else {
				entities, err := p.parseEntity(annotations)
				if err != nil {
					return nil, err
				}
				if schema.Entities == nil {
					schema.Entities = make(ast.Entities)
				}
				for name, entity := range entities {
					if _, exists := schema.Entities[name]; exists {
						return nil, fmt.Errorf("duplicate entity %q at %s", name, fmtPos(tok.Pos))
					}
					if _, exists := schema.Enums[name]; exists {
						return nil, fmt.Errorf("entity %q conflicts with enum at %s", name, fmtPos(tok.Pos))
					}
					schema.Entities[name] = entity
				}
			}
		case "action":
			actions, err := p.parseAction(annotations)
			if err != nil {
				return nil, err
			}
			if schema.Actions == nil {
				schema.Actions = make(ast.Actions)
			}
			for name, action := range actions {
				if _, exists := schema.Actions[name]; exists {
					return nil, fmt.Errorf("duplicate action %q at %s", name, fmtPos(tok.Pos))
				}
				schema.Actions[name] = action
			}
		case "type":
			ct, name, err := p.parseCommonType(annotations)
			if err != nil {
				return nil, err
			}
			if schema.CommonTypes == nil {
				schema.CommonTypes = make(ast.CommonTypes)
			}
			if _, exists := schema.CommonTypes[name]; exists {
				return nil, fmt.Errorf("duplicate common type %q at %s", name, fmtPos(tok.Pos))
			}
			schema.CommonTypes[name] = ct
		default:
			return nil, fmt.Errorf("unexpected token %q at %s", tok.Text, fmtPos(tok.Pos))
		}
	}

	return schema, nil
}

type Annotation struct {
	Key   types.Ident
	Value types.String
}

func (p *Parser) parseAnnotation() (Annotation, error) {
	if _, err := p.expect("@"); err != nil {
		return Annotation{}, err
	}

	key, err := p.expectAnyIdent()
	if err != nil {
		return Annotation{}, err
	}

	var value string
	if p.peek().Text == "(" {
		p.advance()
		value, err = p.expectString()
		if err != nil {
			return Annotation{}, err
		}
		if _, err := p.expect(")"); err != nil {
			return Annotation{}, err
		}
	}

	return Annotation{Key: types.Ident(key), Value: types.String(value)}, nil
}

func (p *Parser) parseNamespace(annotations ast.Annotations) (ast.Namespace, types.Path, error) {
	if _, err := p.expect("namespace"); err != nil {
		return ast.Namespace{}, "", err
	}

	path, err := p.parsePath()
	if err != nil {
		return ast.Namespace{}, "", err
	}

	if _, err := p.expect("{"); err != nil {
		return ast.Namespace{}, "", err
	}

	ns := ast.Namespace{
		Annotations: annotations,
	}

	for p.peek().Text != "}" && !p.peek().isEOF() {
		declAnnotations, err := p.parseAnnotations()
		if err != nil {
			return ast.Namespace{}, "", err
		}

		tok := p.peek()
		switch tok.Text {
		case "entity":
			// Check if it's an enum by peeking ahead
			// After "entity", the next token is the name, and the token after that might be "enum"
			if p.peekAhead(3).Text == "enum" {
				enum, name, err := p.parseEnum(declAnnotations)
				if err != nil {
					return ast.Namespace{}, "", err
				}
				if ns.Enums == nil {
					ns.Enums = make(ast.Enums)
				}
				if _, exists := ns.Enums[name]; exists {
					return ast.Namespace{}, "", fmt.Errorf("duplicate enum %q in namespace at %s", name, fmtPos(tok.Pos))
				}
				if _, exists := ns.Entities[name]; exists {
					return ast.Namespace{}, "", fmt.Errorf("enum %q conflicts with entity in namespace at %s", name, fmtPos(tok.Pos))
				}
				ns.Enums[name] = enum
			} else {
				entities, err := p.parseEntity(declAnnotations)
				if err != nil {
					return ast.Namespace{}, "", err
				}
				if ns.Entities == nil {
					ns.Entities = make(ast.Entities)
				}
				for name, entity := range entities {
					if _, exists := ns.Entities[name]; exists {
						return ast.Namespace{}, "", fmt.Errorf("duplicate entity %q in namespace at %s", name, fmtPos(tok.Pos))
					}
					if _, exists := ns.Enums[name]; exists {
						return ast.Namespace{}, "", fmt.Errorf("entity %q conflicts with enum in namespace at %s", name, fmtPos(tok.Pos))
					}
					ns.Entities[name] = entity
				}
			}
		case "action":
			actions, err := p.parseAction(declAnnotations)
			if err != nil {
				return ast.Namespace{}, "", err
			}
			if ns.Actions == nil {
				ns.Actions = make(ast.Actions)
			}
			for name, action := range actions {
				if _, exists := ns.Actions[name]; exists {
					return ast.Namespace{}, "", fmt.Errorf("duplicate action %q in namespace at %s", name, fmtPos(tok.Pos))
				}
				ns.Actions[name] = action
			}
		case "type":
			ct, name, err := p.parseCommonType(declAnnotations)
			if err != nil {
				return ast.Namespace{}, "", err
			}
			if ns.CommonTypes == nil {
				ns.CommonTypes = make(ast.CommonTypes)
			}
			if _, exists := ns.CommonTypes[name]; exists {
				return ast.Namespace{}, "", fmt.Errorf("duplicate common type %q in namespace at %s", name, fmtPos(tok.Pos))
			}
			ns.CommonTypes[name] = ct
		default:
			return ast.Namespace{}, "", fmt.Errorf("unexpected token %q in namespace at %s", tok.Text, fmtPos(tok.Pos))
		}
	}

	if _, err := p.expect("}"); err != nil {
		return ast.Namespace{}, "", err
	}

	return ns, types.Path(path), nil
}

func (p *Parser) parseEnum(annotations ast.Annotations) (ast.Enum, types.EntityType, error) {
	if _, err := p.expect("entity"); err != nil {
		return ast.Enum{}, "", err
	}

	name, err := p.expectIdent()
	if err != nil {
		return ast.Enum{}, "", err
	}

	if _, err := p.expect("enum"); err != nil {
		return ast.Enum{}, "", err
	}

	values, err := p.parseEnumValues()
	if err != nil {
		return ast.Enum{}, "", err
	}

	if _, err := p.expect(";"); err != nil {
		return ast.Enum{}, "", err
	}

	return ast.Enum{
		Annotations: annotations,
		Values:      values,
	}, types.EntityType(name), nil
}

func (p *Parser) parseEntity(annotations ast.Annotations) (map[types.EntityType]ast.Entity, error) {
	if _, err := p.expect("entity"); err != nil {
		return nil, err
	}

	// Parse comma-separated entity names
	var names []string
	name, err := p.expectIdent()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	for p.peek().Text == "," {
		p.advance()
		name, err := p.expectIdent()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	// Parse shared modifiers for all entities
	var parents []ast.EntityTypeRef
	var attrs ast.Attributes
	var tagsType ast.IsType

	// Parse "in" clause
	if p.peek().Text == "in" {
		p.advance()
		parents, err = p.parseEntityTypeRefs()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional "=" before shape
	if p.peek().Text == "=" {
		p.advance()
	}

	// Parse shape
	if p.peek().Text == "{" {
		attrs, err = p.parseAttributes()
		if err != nil {
			return nil, err
		}
	}

	// Parse tags
	if p.peek().Text == "tags" {
		p.advance()
		tagsType, err = p.parseType()
		if err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(";"); err != nil {
		return nil, err
	}

	// Create entity nodes for each name
	entities := make(map[types.EntityType]ast.Entity)
	for _, n := range names {
		entity := ast.Entity{
			Annotations: annotations,
			MemberOf:    parents,
			Tags:        tagsType,
		}
		if attrs != nil {
			entity.Shape = &ast.RecordType{Attributes: attrs}
		}
		entities[types.EntityType(n)] = entity
	}

	return entities, nil
}

func (p *Parser) parseEnumValues() ([]types.String, error) {
	if _, err := p.expect("["); err != nil {
		return nil, err
	}

	var values []types.String
	for p.peek().Text != "]" && !p.peek().isEOF() {
		val, err := p.expectString()
		if err != nil {
			return nil, err
		}
		values = append(values, types.String(val))

		if p.peek().Text == "," {
			p.advance()
		} else {
			break
		}
	}

	if _, err := p.expect("]"); err != nil {
		return nil, err
	}

	return values, nil
}

func (p *Parser) parseEntityTypeRefs() ([]ast.EntityTypeRef, error) {
	if p.peek().Text == "[" {
		p.advance()
		var refs []ast.EntityTypeRef
		for p.peek().Text != "]" && !p.peek().isEOF() {
			path, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			refs = append(refs, ast.EntityType(types.EntityType(path)))

			if p.peek().Text == "," {
				p.advance()
				// Reject trailing commas
				if p.peek().Text == "]" {
					return nil, fmt.Errorf("unexpected trailing comma at %s", fmtPos(p.peek().Pos))
				}
			} else {
				break
			}
		}
		if _, err := p.expect("]"); err != nil {
			return nil, err
		}
		return refs, nil
	}

	// Single entity type
	path, err := p.parsePath()
	if err != nil {
		return nil, err
	}
	return []ast.EntityTypeRef{ast.EntityType(types.EntityType(path))}, nil
}

func (p *Parser) parseAction(annotations ast.Annotations) (map[types.String]ast.Action, error) {
	if _, err := p.expect("action"); err != nil {
		return nil, err
	}

	// Parse comma-separated action names
	var names []string
	name, err := p.parseActionName()
	if err != nil {
		return nil, err
	}
	names = append(names, name)

	for p.peek().Text == "," {
		p.advance()
		name, err := p.parseActionName()
		if err != nil {
			return nil, err
		}
		names = append(names, name)
	}

	// Parse shared modifiers for all actions
	var memberOf []ast.EntityRef
	var principals []ast.EntityTypeRef
	var resources []ast.EntityTypeRef
	var contextType ast.IsType
	hasAppliesTo := false

	// Parse "in" clause for action groups
	if p.peek().Text == "in" {
		p.advance()
		memberOf, err = p.parseEntityRefs()
		if err != nil {
			return nil, err
		}
	}

	// Parse "appliesTo" clause
	if p.peek().Text == "appliesTo" {
		hasAppliesTo = true
		p.advance()
		if _, err := p.expect("{"); err != nil {
			return nil, err
		}

		for p.peek().Text != "}" && !p.peek().isEOF() {
			tok := p.peek()
			switch tok.Text {
			case "principal":
				p.advance()
				if _, err := p.expect(":"); err != nil {
					return nil, err
				}
				principals, err = p.parseEntityTypeRefs()
				if err != nil {
					return nil, err
				}
			case "resource":
				p.advance()
				if _, err := p.expect(":"); err != nil {
					return nil, err
				}
				resources, err = p.parseEntityTypeRefs()
				if err != nil {
					return nil, err
				}
			case "context":
				p.advance()
				if _, err := p.expect(":"); err != nil {
					return nil, err
				}
				contextType, err = p.parseType()
				if err != nil {
					return nil, err
				}
			default:
				return nil, fmt.Errorf("unexpected %q in appliesTo at %s", tok.Text, fmtPos(tok.Pos))
			}

			if p.peek().Text == "," {
				p.advance()
			}
		}

		if _, err := p.expect("}"); err != nil {
			return nil, err
		}
	}

	if _, err := p.expect(";"); err != nil {
		return nil, err
	}

	// Create action nodes for each name
	actions := make(map[types.String]ast.Action)
	for _, n := range names {
		action := ast.Action{
			Annotations: annotations,
			MemberOf:    memberOf,
		}
		if hasAppliesTo {
			action.AppliesTo = &ast.AppliesTo{
				PrincipalTypes: principals,
				ResourceTypes:  resources,
				Context:        contextType,
			}
		}
		actions[types.String(n)] = action
	}

	return actions, nil
}

func (p *Parser) parseActionName() (string, error) {
	tok := p.peek()
	if tok.Type == TokenString {
		return p.expectString()
	}
	return p.expectIdent()
}

func (p *Parser) parseEntityRefs() ([]ast.EntityRef, error) {
	if p.peek().Text == "[" {
		p.advance()
		var refs []ast.EntityRef
		for p.peek().Text != "]" && !p.peek().isEOF() {
			ref, err := p.parseEntityRef()
			if err != nil {
				return nil, err
			}
			refs = append(refs, ref)

			if p.peek().Text == "," {
				p.advance()
				// Reject trailing commas
				if p.peek().Text == "]" {
					return nil, fmt.Errorf("unexpected trailing comma at %s", fmtPos(p.peek().Pos))
				}
			} else {
				break
			}
		}
		if _, err := p.expect("]"); err != nil {
			return nil, err
		}
		return refs, nil
	}

	ref, err := p.parseEntityRef()
	if err != nil {
		return nil, err
	}
	return []ast.EntityRef{ref}, nil
}

func (p *Parser) parseEntityRef() (ast.EntityRef, error) {
	// Could be either Path::"id" or just "id" (implies Action type)
	tok := p.peek()
	if tok.Type == TokenString {
		// We've verified it's a string, so use consumeString
		id := p.consumeString()
		return ast.EntityRefFromID(types.String(id)), nil
	}

	// Parse path components manually so we can stop before a string ID
	name, err := p.expectIdent()
	if err != nil {
		return ast.EntityRef{}, err
	}

	for p.peek().Text == "::" {
		p.advance()
		// Check if the next token is a string (entity ID)
		if p.peek().Type == TokenString {
			// We've verified it's a string, so use consumeString
			id := p.consumeString()
			return ast.NewEntityRef(types.EntityType(name), types.String(id)), nil
		}
		// Otherwise it's another path component
		next, err := p.expectIdent()
		if err != nil {
			return ast.EntityRef{}, err
		}
		name += "::" + next
	}

	// Just a name implies it's an Action::"name"
	return ast.EntityRefFromID(types.String(name)), nil
}

func (p *Parser) parseCommonType(annotations ast.Annotations) (ast.CommonType, types.Ident, error) {
	if _, err := p.expect("type"); err != nil {
		return ast.CommonType{}, "", err
	}

	name, err := p.expectIdent()
	if err != nil {
		return ast.CommonType{}, "", err
	}

	if _, err := p.expect("="); err != nil {
		return ast.CommonType{}, "", err
	}

	t, err := p.parseType()
	if err != nil {
		return ast.CommonType{}, "", err
	}

	if _, err := p.expect(";"); err != nil {
		return ast.CommonType{}, "", err
	}

	ct := ast.CommonType{
		Annotations: annotations,
		Type:        t,
	}
	return ct, types.Ident(name), nil
}

func (p *Parser) parseType() (ast.IsType, error) {
	tok := p.peek()

	switch tok.Text {
	case "{":
		attrs, err := p.parseAttributes()
		if err != nil {
			return nil, err
		}
		return ast.Record(attrs), nil
	case "String":
		p.advance()
		return ast.String(), nil
	case "Long":
		p.advance()
		return ast.Long(), nil
	case "Bool":
		p.advance()
		return ast.Bool(), nil
	case "Set":
		p.advance()
		if _, err := p.expect("<"); err != nil {
			return nil, err
		}
		elem, err := p.parseType()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(">"); err != nil {
			return nil, err
		}
		return ast.Set(elem), nil
	// case "__cedar":
	// 	// Extension type: __cedar::ipaddr, __cedar::decimal, etc.
	// 	p.advance()
	// 	if _, err := p.expect("::"); err != nil {
	// 		return nil, err
	// 	}
	// 	name, err := p.expectIdent()
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	return ast.ExtensionType{Name: types.Ident(name)}, nil
	default:
		// Could be an entity reference or a type reference
		if tok.Type == TokenIdent {
			path, err := p.parsePath()
			if err != nil {
				return nil, err
			}
			// If it contains "::", treat as entity type ref, otherwise common type ref
			// if strings.Contains(path, "::") {
			// 	return ast.EntityType(types.EntityType(path)), nil
			// }
			return ast.Type(types.Path(path)), nil
		}
		return nil, fmt.Errorf("expected type, got %q at %s", tok.Text, fmtPos(tok.Pos))
	}
}

func (p *Parser) parseAttributes() (ast.Attributes, error) {
	if _, err := p.expect("{"); err != nil {
		return nil, err
	}

	attrs := make(ast.Attributes)
	for p.peek().Text != "}" && !p.peek().isEOF() {
		key, attr, err := p.parseAttribute()
		if err != nil {
			return nil, err
		}

		// Check for duplicate attribute keys
		if _, exists := attrs[key]; exists {
			return nil, fmt.Errorf("duplicate attribute %q at %s", key, fmtPos(p.peek().Pos))
		}
		attrs[key] = attr

		if p.peek().Text == "," {
			p.advance()
		}
	}

	if _, err := p.expect("}"); err != nil {
		return nil, err
	}

	return attrs, nil
}

func (p *Parser) parseAttribute() (types.String, ast.Attribute, error) {
	var err error
	annotations, err := p.parseAnnotations()
	if err != nil {
		return "", ast.Attribute{}, err
	}

	tok := p.peek()
	var key string

	if tok.Type == TokenString {
		key, err = p.expectString()
	} else {
		key, err = p.expectIdent()
	}
	if err != nil {
		return "", ast.Attribute{}, err
	}

	optional := false
	if p.peek().Text == "?" {
		p.advance()
		optional = true
	}

	if _, err := p.expect(":"); err != nil {
		return "", ast.Attribute{}, err
	}

	t, err := p.parseType()
	if err != nil {
		return "", ast.Attribute{}, err
	}

	return types.String(key), ast.Attribute{
		Type:        t,
		Optional:    optional,
		Annotations: annotations,
	}, nil
}

func (p *Parser) parsePath() (string, error) {
	name, err := p.expectIdent()
	if err != nil {
		return "", err
	}

	for p.peek().Text == "::" {
		p.advance()
		next, err := p.expectIdent()
		if err != nil {
			return "", err
		}
		name += "::" + next
	}

	return name, nil
}

// ParseSchema parses Cedar schema from the given source bytes.
func ParseSchema(filename string, src []byte) (*ast.Schema, error) {
	p, err := New(filename, src)
	if err != nil {
		return nil, err
	}
	return p.Parse()
}
