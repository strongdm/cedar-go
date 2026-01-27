// Package parser provides a parser for Cedar human-readable schema format.
package parser

import (
	"fmt"
	"io"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
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

// Parse parses a complete Cedar schema.
func (p *Parser) Parse() (*ast.Schema, error) {
	var nodes []ast.IsNode

	for !p.peek().isEOF() {
		var annotations []ast.Annotation

		// Parse annotations
		for p.peek().Text == "@" {
			ann, err := p.parseAnnotation()
			if err != nil {
				return nil, err
			}
			annotations = append(annotations, ann)
		}

		// Parse declaration
		tok := p.peek()
		switch tok.Text {
		case "namespace":
			ns, err := p.parseNamespace(annotations)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ns)
		case "entity":
			entities, err := p.parseEntity(annotations)
			if err != nil {
				return nil, err
			}
			for _, entity := range entities {
				nodes = append(nodes, entity)
			}
		case "action":
			actions, err := p.parseAction(annotations)
			if err != nil {
				return nil, err
			}
			for _, action := range actions {
				nodes = append(nodes, action)
			}
		case "type":
			ct, err := p.parseCommonType(annotations)
			if err != nil {
				return nil, err
			}
			nodes = append(nodes, ct)
		default:
			return nil, fmt.Errorf("unexpected token %q at %s", tok.Text, fmtPos(tok.Pos))
		}
	}

	return ast.NewSchema(nodes...), nil
}

func (p *Parser) parseAnnotation() (ast.Annotation, error) {
	if _, err := p.expect("@"); err != nil {
		return ast.Annotation{}, err
	}

	key, err := p.expectIdent()
	if err != nil {
		return ast.Annotation{}, err
	}

	var value string
	if p.peek().Text == "(" {
		p.advance()
		value, err = p.expectString()
		if err != nil {
			return ast.Annotation{}, err
		}
		if _, err := p.expect(")"); err != nil {
			return ast.Annotation{}, err
		}
	}

	return ast.Annotation{Key: types.Ident(key), Value: types.String(value)}, nil
}

func (p *Parser) parseNamespace(annotations []ast.Annotation) (ast.NamespaceNode, error) {
	if _, err := p.expect("namespace"); err != nil {
		return ast.NamespaceNode{}, err
	}

	path, err := p.parsePath()
	if err != nil {
		return ast.NamespaceNode{}, err
	}

	if _, err := p.expect("{"); err != nil {
		return ast.NamespaceNode{}, err
	}

	var decls []ast.IsDeclaration

	for p.peek().Text != "}" && !p.peek().isEOF() {
		var declAnnotations []ast.Annotation

		for p.peek().Text == "@" {
			ann, err := p.parseAnnotation()
			if err != nil {
				return ast.NamespaceNode{}, err
			}
			declAnnotations = append(declAnnotations, ann)
		}

		tok := p.peek()
		switch tok.Text {
		case "entity":
			entities, err := p.parseEntity(declAnnotations)
			if err != nil {
				return ast.NamespaceNode{}, err
			}
			decls = append(decls, entities...)
		case "action":
			actions, err := p.parseAction(declAnnotations)
			if err != nil {
				return ast.NamespaceNode{}, err
			}
			for _, action := range actions {
				decls = append(decls, action)
			}
		case "type":
			ct, err := p.parseCommonType(declAnnotations)
			if err != nil {
				return ast.NamespaceNode{}, err
			}
			decls = append(decls, ct)
		default:
			return ast.NamespaceNode{}, fmt.Errorf("unexpected token %q in namespace at %s", tok.Text, fmtPos(tok.Pos))
		}
	}

	if _, err := p.expect("}"); err != nil {
		return ast.NamespaceNode{}, err
	}

	ns := ast.Namespace(types.Path(path), decls...)
	ns.Annotations = annotations
	return ns, nil
}

func (p *Parser) parseEntity(annotations []ast.Annotation) ([]ast.IsDeclaration, error) {
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

	// Check for enum syntax: entity Name enum ["val1", "val2"];
	// Note: enum only supports single entity name
	if p.peek().Text == "enum" {
		if len(names) > 1 {
			return nil, fmt.Errorf("enum entity cannot have multiple names")
		}
		p.advance()
		values, err := p.parseEnumValues()
		if err != nil {
			return nil, err
		}
		if _, err := p.expect(";"); err != nil {
			return nil, err
		}
		enum := ast.Enum(types.EntityType(names[0]), values...)
		enum.Annotations = annotations
		return []ast.IsDeclaration{enum}, nil
	}

	// Parse shared modifiers for all entities
	var parents []ast.EntityTypeRef
	var pairs []ast.Pair
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
		pairs, err = p.parseRecordPairs()
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
	var entities []ast.IsDeclaration
	for _, n := range names {
		entity := ast.Entity(types.EntityType(n))
		entity.Annotations = annotations
		if len(parents) > 0 {
			entity = entity.MemberOf(parents...)
		}
		if len(pairs) > 0 {
			entity = entity.Shape(pairs...)
		}
		if tagsType != nil {
			entity = entity.Tags(tagsType)
		}
		entities = append(entities, entity)
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

func (p *Parser) parseAction(annotations []ast.Annotation) ([]ast.ActionNode, error) {
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
	var actions []ast.ActionNode
	for _, n := range names {
		action := ast.Action(types.String(n))
		action.Annotations = annotations
		if len(memberOf) > 0 {
			action = action.MemberOf(memberOf...)
		}
		if hasAppliesTo {
			if len(principals) > 0 {
				action = action.Principal(principals...)
			}
			if len(resources) > 0 {
				action = action.Resource(resources...)
			}
			if contextType != nil {
				action = action.Context(contextType)
			}
		}
		actions = append(actions, action)
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
		return ast.UID(types.String(id)), nil
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
			return ast.EntityUID(types.EntityType(name), types.String(id)), nil
		}
		// Otherwise it's another path component
		next, err := p.expectIdent()
		if err != nil {
			return ast.EntityRef{}, err
		}
		name += "::" + next
	}

	// Just a name implies it's an Action::"name"
	return ast.UID(types.String(name)), nil
}

func (p *Parser) parseCommonType(annotations []ast.Annotation) (ast.CommonTypeNode, error) {
	if _, err := p.expect("type"); err != nil {
		return ast.CommonTypeNode{}, err
	}

	name, err := p.expectIdent()
	if err != nil {
		return ast.CommonTypeNode{}, err
	}

	if _, err := p.expect("="); err != nil {
		return ast.CommonTypeNode{}, err
	}

	t, err := p.parseType()
	if err != nil {
		return ast.CommonTypeNode{}, err
	}

	if _, err := p.expect(";"); err != nil {
		return ast.CommonTypeNode{}, err
	}

	ct := ast.CommonType(types.Ident(name), t)
	ct.Annotations = annotations
	return ct, nil
}

func (p *Parser) parseType() (ast.IsType, error) {
	tok := p.peek()

	switch tok.Text {
	case "{":
		pairs, err := p.parseRecordPairs()
		if err != nil {
			return nil, err
		}
		return ast.Record(pairs...), nil
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

func (p *Parser) parseRecordPairs() ([]ast.Pair, error) {
	if _, err := p.expect("{"); err != nil {
		return nil, err
	}

	var pairs []ast.Pair
	for p.peek().Text != "}" && !p.peek().isEOF() {
		pair, err := p.parseRecordPair()
		if err != nil {
			return nil, err
		}
		pairs = append(pairs, pair)

		if p.peek().Text == "," {
			p.advance()
		}
	}

	if _, err := p.expect("}"); err != nil {
		return nil, err
	}

	return pairs, nil
}

func (p *Parser) parseRecordPair() (ast.Pair, error) {
	tok := p.peek()
	var key string
	var err error

	if tok.Type == TokenString {
		key, err = p.expectString()
	} else {
		key, err = p.expectIdent()
	}
	if err != nil {
		return ast.Pair{}, err
	}

	optional := false
	if p.peek().Text == "?" {
		p.advance()
		optional = true
	}

	if _, err := p.expect(":"); err != nil {
		return ast.Pair{}, err
	}

	t, err := p.parseType()
	if err != nil {
		return ast.Pair{}, err
	}

	if optional {
		return ast.Optional(types.String(key), t), nil
	}
	return ast.Attribute(types.String(key), t), nil
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
