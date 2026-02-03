package schema2

import (
	"encoding/json"

	existingast "github.com/cedar-policy/cedar-go/internal/schema/ast"
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
