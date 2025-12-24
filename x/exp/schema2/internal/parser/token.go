package parser

import (
	"fmt"
	"strconv"

	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

// Position represents a position in the source file.
type Position = ast.Position

// TokenType represents the type of a token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdent
	TokenString
	TokenOperator
	TokenUnknown
)

// Token represents a lexical token.
type Token struct {
	Type TokenType
	Pos  Position
	Text string
}

func (t Token) isEOF() bool {
	return t.Type == TokenEOF
}

func (t Token) isIdent() bool {
	return t.Type == TokenIdent
}

func (t Token) isString() bool {
	return t.Type == TokenString
}

// stringValue returns the unquoted string value.
func (t Token) stringValue() (string, error) {
	if t.Type != TokenString {
		return "", fmt.Errorf("expected string, got %v", t.Type)
	}
	return strconv.Unquote(t.Text)
}
