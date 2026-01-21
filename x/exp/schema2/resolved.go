package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// ResolvedSchema wraps ast.ResolvedSchema to provide access to resolved schema data.
type ResolvedSchema struct {
	ast.ResolvedSchema
}
