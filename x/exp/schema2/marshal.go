package schema2

import (
	"bytes"
	"encoding/json"

	existingast "github.com/cedar-policy/cedar-go/internal/schema/ast"
)

// MarshalJSON serializes the schema to JSON format.
func (s *Schema) MarshalJSON() ([]byte, error) {
	js := s.ast.ToJSON()
	return json.Marshal(js)
}

// MarshalJSONIndent serializes the schema to indented JSON format.
func (s *Schema) MarshalJSONIndent(prefix, indent string) ([]byte, error) {
	js := s.ast.ToJSON()
	return json.MarshalIndent(js, prefix, indent)
}

// MarshalCedar serializes the schema to human-readable Cedar format.
func (s *Schema) MarshalCedar() ([]byte, error) {
	js := s.ast.ToJSON()
	humanAST := existingast.ConvertJSON2Human(js)

	var buf bytes.Buffer
	if err := existingast.Format(humanAST, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// UnmarshalCedar deserializes a schema from human-readable Cedar format.
func (s *Schema) UnmarshalCedar(data []byte) error {
	parsed, err := ParseCedar(data)
	if err != nil {
		return err
	}
	*s = *parsed
	return nil
}
