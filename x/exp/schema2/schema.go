package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// Schema is a mutable builder for constructing Cedar schemas.
// Use NewSchema() to create a new schema, then chain builder methods
// to define namespaces, entities, and actions.
//
// Call Resolve() to produce an immutable ResolvedSchema with
// fully-qualified names and computed entity hierarchies.
//
// Example:
//
//	schema := NewSchema().
//	    Namespace("MyApp").
//	    Entity("User").In("Group").
//	    Entity("Group").
//	    Action("read").Principals("User").Resources("Document")
//
//	resolved, err := schema.Resolve()
type Schema struct {
	ast *ast.Schema

	// Current builder state for method chaining (used for backward compatibility)
	currentNamespace string
	currentEntity    string
	currentAction    string
}

// NewSchema creates a new empty schema builder.
func NewSchema() *Schema {
	return &Schema{
		ast: ast.NewSchema(),
	}
}

// Namespace switches to (or creates) a namespace for subsequent entity/action definitions.
// Use "" for the anonymous/default namespace.
func (s *Schema) Namespace(name string) *Schema {
	s.ast.GetOrCreateNamespace(name)
	s.currentNamespace = name
	s.currentEntity = ""
	s.currentAction = ""
	return s
}

// Entity defines or switches to an entity type in the current namespace.
// Returns an EntityBuilder for compile-time safe method chaining.
// The EntityBuilder provides In(), Attributes(), Tags(), Enum() methods,
// and terminal methods to continue building (Entity(), Action(), Namespace(), Schema()).
func (s *Schema) Entity(name string) *EntityBuilder {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.GetOrCreateEntity(name)
	s.currentEntity = name
	s.currentAction = ""
	return &EntityBuilder{
		schema:    s,
		namespace: s.currentNamespace,
		name:      name,
	}
}

// Action defines or switches to an action in the current namespace.
// Returns an ActionBuilder for compile-time safe method chaining.
// The ActionBuilder provides In(), Principals(), Resources(), Context() methods,
// and terminal methods to continue building (Entity(), Action(), Namespace(), Schema()).
func (s *Schema) Action(name string) *ActionBuilder {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.GetOrCreateAction(name)
	s.currentAction = name
	s.currentEntity = ""
	return &ActionBuilder{
		schema:    s,
		namespace: s.currentNamespace,
		name:      name,
	}
}

// CommonType defines a common (reusable) type in the current namespace.
func (s *Schema) CommonType(name string, t Type) *Schema {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.CommonTypes[name] = t.toAST()
	return s
}
