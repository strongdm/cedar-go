package schema2

import (
	"encoding/json"

	existingast "github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// UnmarshalJSON implements json.Unmarshaler for Schema.
// This parses a Cedar schema from JSON format into the receiver.
func (s *Schema) UnmarshalJSON(data []byte) error {
	var js existingast.JSONSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return err
	}
	s.ast = ast.FromJSON(js)
	return nil
}

// parseConfig holds configuration options for parsing Cedar schemas.
type parseConfig struct {
	filename string
}

// ParseOption is a functional option for configuring Cedar schema parsing.
type ParseOption func(*parseConfig)

// WithFilename sets the filename for error messages during parsing.
func WithFilename(name string) ParseOption {
	return func(cfg *parseConfig) {
		cfg.filename = name
	}
}

// ParseCedar parses a Cedar schema from human-readable Cedar format.
// Optional ParseOption arguments can be used to configure parsing behavior.
//
// Example:
//
//	schema, err := ParseCedar(data)
//	schema, err := ParseCedar(data, WithFilename("schema.cedarschema"))
func ParseCedar(data []byte, opts ...ParseOption) (*Schema, error) {
	cfg := &parseConfig{
		filename: "",
	}
	for _, opt := range opts {
		opt(cfg)
	}

	humanAST, err := parser.ParseFile(cfg.filename, data)
	if err != nil {
		return nil, err
	}

	// Convert human AST to JSON, then to our internal AST
	js := existingast.ConvertHuman2JSON(humanAST)

	return &Schema{
		ast: ast.FromJSON(js),
	}, nil
}
