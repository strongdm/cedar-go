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
type Schema struct {
	ast *ast.Schema

	// Current builder state for method chaining
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
// Subsequent calls to In(), Attributes(), Tags(), Enum() apply to this entity.
func (s *Schema) Entity(name string) *Schema {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.GetOrCreateEntity(name)
	s.currentEntity = name
	s.currentAction = ""
	return s
}

// In specifies parent entity types for the current entity (memberOf).
func (s *Schema) In(parents ...string) *Schema {
	if s.currentEntity == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	et := ns.GetOrCreateEntity(s.currentEntity)
	et.MemberOf = append(et.MemberOf, parents...)
	return s
}

// Attributes sets the shape (record type) for the current entity.
func (s *Schema) Attributes(attrs ...*Attribute) *Schema {
	if s.currentEntity == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	et := ns.GetOrCreateEntity(s.currentEntity)
	if et.Shape == nil {
		et.Shape = &ast.RecordType{Attributes: make(map[string]*ast.Attribute)}
	}
	for _, attr := range attrs {
		et.Shape.Attributes[attr.name] = &ast.Attribute{
			Name:     attr.name,
			Type:     attr.typ.toAST(),
			Required: attr.required,
		}
	}
	return s
}

// Tags sets the tag type for the current entity.
func (s *Schema) Tags(t Type) *Schema {
	if s.currentEntity == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	et := ns.GetOrCreateEntity(s.currentEntity)
	et.Tags = t.toAST()
	return s
}

// Enum defines enumerated values for the current entity, making it an enum type.
func (s *Schema) Enum(values ...string) *Schema {
	if s.currentEntity == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	et := ns.GetOrCreateEntity(s.currentEntity)
	et.Enum = values
	return s
}

// Action defines or switches to an action in the current namespace.
// Subsequent calls to ActionIn(), Principals(), Resources(), Context() apply to this action.
func (s *Schema) Action(name string) *Schema {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.GetOrCreateAction(name)
	s.currentAction = name
	s.currentEntity = ""
	return s
}

// ActionIn specifies parent action groups for the current action (memberOf).
func (s *Schema) ActionIn(parents ...string) *Schema {
	if s.currentAction == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	a := ns.GetOrCreateAction(s.currentAction)
	for _, p := range parents {
		a.MemberOf = append(a.MemberOf, ast.ActionRef{Name: p})
	}
	return s
}

// Principals specifies principal entity types for the current action's appliesTo.
func (s *Schema) Principals(types ...string) *Schema {
	if s.currentAction == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	a := ns.GetOrCreateAction(s.currentAction)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Principals = append(a.AppliesTo.Principals, types...)
	return s
}

// Resources specifies resource entity types for the current action's appliesTo.
func (s *Schema) Resources(types ...string) *Schema {
	if s.currentAction == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	a := ns.GetOrCreateAction(s.currentAction)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Resources = append(a.AppliesTo.Resources, types...)
	return s
}

// Context sets the context type for the current action's appliesTo.
func (s *Schema) Context(t Type) *Schema {
	if s.currentAction == "" {
		return s
	}
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	a := ns.GetOrCreateAction(s.currentAction)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Context = t.toAST()
	return s
}

// CommonType defines a common (reusable) type in the current namespace.
func (s *Schema) CommonType(name string, t Type) *Schema {
	ns := s.ast.GetOrCreateNamespace(s.currentNamespace)
	ns.CommonTypes[name] = t.toAST()
	return s
}
