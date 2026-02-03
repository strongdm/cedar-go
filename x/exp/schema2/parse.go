package schema2

import (
	"encoding/json"

	existingast "github.com/cedar-policy/cedar-go/internal/schema/ast"
	"github.com/cedar-policy/cedar-go/internal/schema/parser"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// ParseJSON parses a Cedar schema from JSON format.
func ParseJSON(data []byte) (*Schema, error) {
	var js existingast.JSONSchema
	if err := json.Unmarshal(data, &js); err != nil {
		return nil, err
	}

	return &Schema{
		ast: ast.FromJSON(js),
	}, nil
}

// ParseCedar parses a Cedar schema from human-readable Cedar format.
func ParseCedar(data []byte) (*Schema, error) {
	humanAST, err := parser.ParseFile("", data)
	if err != nil {
		return nil, err
	}

	// Convert human AST to JSON, then to our internal AST
	js := existingast.ConvertHuman2JSON(humanAST)

	return &Schema{
		ast: ast.FromJSON(js),
	}, nil
}

// ParseCedarWithFilename parses a Cedar schema from human-readable format
// with a filename for error messages.
func ParseCedarWithFilename(filename string, data []byte) (*Schema, error) {
	humanAST, err := parser.ParseFile(filename, data)
	if err != nil {
		return nil, err
	}

	// Convert human AST to JSON, then to our internal AST
	js := existingast.ConvertHuman2JSON(humanAST)

	return &Schema{
		ast: ast.FromJSON(js),
	}, nil
}
