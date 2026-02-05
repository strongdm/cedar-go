package scan

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/cedar-policy/cedar-go/internal/rust"
)

// TokenKind identifies the type of a lexical token.
type TokenKind int

const (
	TokenEOF TokenKind = iota
	TokenIdent
	TokenString
	TokenDoubleColon
	TokenLBrace
	TokenRBrace
	TokenLBracket
	TokenRBracket
	TokenLAngle
	TokenRAngle
	TokenLParen
	TokenRParen
	TokenComma
	TokenSemicolon
	TokenColon
	TokenQuestion
	TokenEquals
	TokenAt
)

func (k TokenKind) String() string {
	switch k {
	case TokenEOF:
		return "EOF"
	case TokenIdent:
		return "identifier"
	case TokenString:
		return "string"
	case TokenDoubleColon:
		return "::"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenLAngle:
		return "<"
	case TokenRAngle:
		return ">"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenComma:
		return ","
	case TokenSemicolon:
		return ";"
	case TokenColon:
		return ":"
	case TokenQuestion:
		return "?"
	case TokenEquals:
		return "="
	case TokenAt:
		return "@"
	default:
		return "unknown"
	}
}

// Position is a line:column in the source.
type Position struct {
	Line   int
	Column int
}

// Token is a lexical token from the scanner.
type Token struct {
	Kind  TokenKind
	Text  string // raw text
	Value string // for strings: unescaped value; for idents: same as Text
	Pos   Position
}

// Scanner tokenizes Cedar schema text.
type Scanner struct {
	src    []byte
	offset int
	line   int
	col    int
}

// New creates a scanner for the given source bytes.
func New(src []byte) *Scanner {
	return &Scanner{src: src, line: 1, col: 1}
}

// Next returns the next token.
func (s *Scanner) Next() (Token, error) {
	s.skipWhitespaceAndComments()
	if s.offset >= len(s.src) {
		return Token{Kind: TokenEOF, Pos: s.pos()}, nil
	}

	pos := s.pos()
	ch := s.peek()

	switch {
	case ch == '"':
		return s.scanString(pos)
	case ch == '@':
		s.advance()
		return Token{Kind: TokenAt, Text: "@", Pos: pos}, nil
	case ch == '{':
		s.advance()
		return Token{Kind: TokenLBrace, Text: "{", Pos: pos}, nil
	case ch == '}':
		s.advance()
		return Token{Kind: TokenRBrace, Text: "}", Pos: pos}, nil
	case ch == '[':
		s.advance()
		return Token{Kind: TokenLBracket, Text: "[", Pos: pos}, nil
	case ch == ']':
		s.advance()
		return Token{Kind: TokenRBracket, Text: "]", Pos: pos}, nil
	case ch == '<':
		s.advance()
		return Token{Kind: TokenLAngle, Text: "<", Pos: pos}, nil
	case ch == '>':
		s.advance()
		return Token{Kind: TokenRAngle, Text: ">", Pos: pos}, nil
	case ch == '(':
		s.advance()
		return Token{Kind: TokenLParen, Text: "(", Pos: pos}, nil
	case ch == ')':
		s.advance()
		return Token{Kind: TokenRParen, Text: ")", Pos: pos}, nil
	case ch == ',':
		s.advance()
		return Token{Kind: TokenComma, Text: ",", Pos: pos}, nil
	case ch == ';':
		s.advance()
		return Token{Kind: TokenSemicolon, Text: ";", Pos: pos}, nil
	case ch == '?':
		s.advance()
		return Token{Kind: TokenQuestion, Text: "?", Pos: pos}, nil
	case ch == '=':
		s.advance()
		return Token{Kind: TokenEquals, Text: "=", Pos: pos}, nil
	case ch == ':':
		s.advance()
		if s.offset < len(s.src) && s.peek() == ':' {
			s.advance()
			return Token{Kind: TokenDoubleColon, Text: "::", Pos: pos}, nil
		}
		return Token{Kind: TokenColon, Text: ":", Pos: pos}, nil
	case isIdentStart(ch):
		return s.scanIdent(pos), nil
	default:
		s.advance()
		return Token{}, fmt.Errorf("unexpected character %q at line %d, column %d", ch, pos.Line, pos.Column)
	}
}

func (s *Scanner) scanIdent(pos Position) Token {
	start := s.offset
	for s.offset < len(s.src) {
		ch := s.peek()
		if !isIdentContinue(ch) {
			break
		}
		s.advance()
	}
	text := string(s.src[start:s.offset])
	return Token{Kind: TokenIdent, Text: text, Value: text, Pos: pos}
}

func (s *Scanner) scanString(pos Position) (Token, error) {
	start := s.offset
	s.advance() // skip opening "
	var content []byte
	for s.offset < len(s.src) {
		ch := s.peek()
		if ch == '"' {
			content = s.src[start+1 : s.offset]
			s.advance() // skip closing "
			val, _, err := rust.Unquote(content, false)
			if err != nil {
				return Token{}, fmt.Errorf("invalid string at line %d, column %d: %w", pos.Line, pos.Column, err)
			}
			text := string(s.src[start:s.offset])
			return Token{Kind: TokenString, Text: text, Value: val, Pos: pos}, nil
		}
		if ch == '\\' {
			s.advance()
			if s.offset < len(s.src) {
				s.advance()
			}
			continue
		}
		s.advance()
	}
	return Token{}, fmt.Errorf("unterminated string at line %d, column %d", pos.Line, pos.Column)
}

func (s *Scanner) skipWhitespaceAndComments() {
	for s.offset < len(s.src) {
		ch := s.peek()
		if unicode.IsSpace(ch) {
			s.advance()
			continue
		}
		if ch == '/' && s.offset+1 < len(s.src) {
			next := rune(s.src[s.offset+1])
			if next == '/' {
				s.skipLineComment()
				continue
			}
			if next == '*' {
				s.skipBlockComment()
				continue
			}
		}
		break
	}
}

func (s *Scanner) skipLineComment() {
	for s.offset < len(s.src) {
		ch := s.peek()
		s.advance()
		if ch == '\n' {
			break
		}
	}
}

func (s *Scanner) skipBlockComment() {
	s.advance() // /
	s.advance() // *
	for s.offset < len(s.src) {
		ch := s.peek()
		s.advance()
		if ch == '*' && s.offset < len(s.src) && s.peek() == '/' {
			s.advance()
			return
		}
	}
}

func (s *Scanner) peek() rune {
	ch, _ := utf8.DecodeRune(s.src[s.offset:])
	return ch
}

func (s *Scanner) advance() {
	ch, size := utf8.DecodeRune(s.src[s.offset:])
	s.offset += size
	if ch == '\n' {
		s.line++
		s.col = 1
	} else {
		s.col++
	}
}

func (s *Scanner) pos() Position {
	return Position{Line: s.line, Column: s.col}
}

func isIdentStart(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch)
}

func isIdentContinue(ch rune) bool {
	return ch == '_' || unicode.IsLetter(ch) || unicode.IsDigit(ch)
}

// Quote escapes a string value for Cedar output.
func Quote(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, ch := range s {
		switch ch {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\x00':
			b.WriteString(`\0`)
		default:
			if !unicode.IsPrint(ch) {
				fmt.Fprintf(&b, `\u{%x}`, ch)
			} else {
				b.WriteRune(ch)
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
