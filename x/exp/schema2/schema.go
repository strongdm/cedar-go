// Package schema2 provides a new implementation of Cedar schema parsing and serialization.
package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/json"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/parser"
)

// UnmarshalCedar parses Cedar human-readable schema format into an AST.
func UnmarshalCedar(src []byte) (*ast.Schema, error) {
	return parser.ParseSchema(src)
}

// UnmarshalJSON parses Cedar JSON schema format into an AST.
func UnmarshalJSON(src []byte) (*ast.Schema, error) {
	return json.Unmarshal(src)
}

// MarshalJSON serializes a schema to Cedar JSON schema format.
func MarshalJSON(s *ast.Schema) ([]byte, error) {
	return json.Marshal(s)
}
