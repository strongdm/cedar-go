package ast

import "github.com/cedar-policy/cedar-go/internal/schema/token"

// BadTypeForTesting is a type that implements Type but isn't one of the known types
// (Path, SetType, RecordType). It's used for testing error paths in conversion
// functions that handle unknown types.
//
// This type should only be used in tests.
type BadTypeForTesting struct{}

func (BadTypeForTesting) isNode()             {}
func (BadTypeForTesting) isType()             {}
func (BadTypeForTesting) Pos() token.Position { return token.Position{} }
func (BadTypeForTesting) End() token.Position { return token.Position{} }

// NewBadTypeForTesting creates a BadTypeForTesting for testing purposes.
func NewBadTypeForTesting() Type {
	return BadTypeForTesting{}
}
