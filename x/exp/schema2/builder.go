package schema2

import (
	"github.com/cedar-policy/cedar-go/x/exp/schema2/internal/ast"
)

// EntityBuilder provides methods for defining an entity type.
// Use Schema.Entity() to obtain an EntityBuilder.
type EntityBuilder struct {
	schema    *Schema
	namespace string
	name      string
}

// In specifies parent entity types for this entity (memberOf).
func (eb *EntityBuilder) In(parents ...string) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	et := ns.GetOrCreateEntity(eb.name)
	et.MemberOf = append(et.MemberOf, parents...)
	return eb
}

// Attributes sets the shape (record type) for this entity.
func (eb *EntityBuilder) Attributes(attrs ...*Attribute) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	et := ns.GetOrCreateEntity(eb.name)
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
	return eb
}

// Tags sets the tag type for this entity.
func (eb *EntityBuilder) Tags(t Type) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	et := ns.GetOrCreateEntity(eb.name)
	et.Tags = t.toAST()
	return eb
}

// Enum defines enumerated values for this entity, making it an enum type.
func (eb *EntityBuilder) Enum(values ...string) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	et := ns.GetOrCreateEntity(eb.name)
	et.Enum = values
	return eb
}

// Entity starts defining a new entity type, returning a new EntityBuilder.
func (eb *EntityBuilder) Entity(name string) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	ns.GetOrCreateEntity(name)
	return &EntityBuilder{
		schema:    eb.schema,
		namespace: eb.namespace,
		name:      name,
	}
}

// Action switches to defining an action, returning an ActionBuilder.
func (eb *EntityBuilder) Action(name string) *ActionBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	ns.GetOrCreateAction(name)
	return &ActionBuilder{
		schema:    eb.schema,
		namespace: eb.namespace,
		name:      name,
	}
}

// Namespace switches to a different namespace, returning the Schema.
func (eb *EntityBuilder) Namespace(name string) *Schema {
	eb.schema.ast.GetOrCreateNamespace(name)
	eb.schema.currentNamespace = name
	eb.schema.currentEntity = ""
	eb.schema.currentAction = ""
	return eb.schema
}

// CommonType defines a common (reusable) type in the current namespace.
func (eb *EntityBuilder) CommonType(name string, t Type) *EntityBuilder {
	ns := eb.schema.ast.GetOrCreateNamespace(eb.namespace)
	ns.CommonTypes[name] = t.toAST()
	return eb
}

// Schema returns the underlying Schema, ending the builder chain.
func (eb *EntityBuilder) Schema() *Schema {
	return eb.schema
}

// Resolve converts the schema to a ResolvedSchema.
// This is a convenience method that calls Schema().Resolve().
func (eb *EntityBuilder) Resolve() (*ResolvedSchema, error) {
	return eb.schema.Resolve()
}

// MustResolve converts the schema to a ResolvedSchema, panicking on error.
// This is a convenience method that calls Schema().MustResolve().
func (eb *EntityBuilder) MustResolve() *ResolvedSchema {
	return eb.schema.MustResolve()
}

// MarshalJSON serializes the schema to JSON format.
// This is a convenience method that calls Schema().MarshalJSON().
func (eb *EntityBuilder) MarshalJSON() ([]byte, error) {
	return eb.schema.MarshalJSON()
}

// MarshalJSONIndent serializes the schema to indented JSON format.
// This is a convenience method that calls Schema().MarshalJSONIndent().
func (eb *EntityBuilder) MarshalJSONIndent(prefix, indent string) ([]byte, error) {
	return eb.schema.MarshalJSONIndent(prefix, indent)
}

// MarshalCedar serializes the schema to human-readable Cedar format.
// This is a convenience method that calls Schema().MarshalCedar().
func (eb *EntityBuilder) MarshalCedar() ([]byte, error) {
	return eb.schema.MarshalCedar()
}

// ActionBuilder provides methods for defining an action.
// Use Schema.Action() or EntityBuilder.Action() to obtain an ActionBuilder.
type ActionBuilder struct {
	schema    *Schema
	namespace string
	name      string
}

// In specifies parent action groups for this action (memberOf).
func (ab *ActionBuilder) In(parents ...string) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	a := ns.GetOrCreateAction(ab.name)
	for _, p := range parents {
		a.MemberOf = append(a.MemberOf, ast.ActionRef{Name: p})
	}
	return ab
}

// Principals specifies principal entity types for this action's appliesTo.
func (ab *ActionBuilder) Principals(types ...string) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	a := ns.GetOrCreateAction(ab.name)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Principals = append(a.AppliesTo.Principals, types...)
	return ab
}

// Resources specifies resource entity types for this action's appliesTo.
func (ab *ActionBuilder) Resources(types ...string) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	a := ns.GetOrCreateAction(ab.name)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Resources = append(a.AppliesTo.Resources, types...)
	return ab
}

// Context sets the context type for this action's appliesTo.
// Context only accepts record types, enforced at compile time.
func (ab *ActionBuilder) Context(t RecordType) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	a := ns.GetOrCreateAction(ab.name)
	if a.AppliesTo == nil {
		a.AppliesTo = &ast.AppliesTo{}
	}
	a.AppliesTo.Context = t.toAST()
	return ab
}

// Action starts defining a new action, returning a new ActionBuilder.
func (ab *ActionBuilder) Action(name string) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	ns.GetOrCreateAction(name)
	return &ActionBuilder{
		schema:    ab.schema,
		namespace: ab.namespace,
		name:      name,
	}
}

// Entity switches to defining an entity, returning an EntityBuilder.
func (ab *ActionBuilder) Entity(name string) *EntityBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	ns.GetOrCreateEntity(name)
	return &EntityBuilder{
		schema:    ab.schema,
		namespace: ab.namespace,
		name:      name,
	}
}

// Namespace switches to a different namespace, returning the Schema.
func (ab *ActionBuilder) Namespace(name string) *Schema {
	ab.schema.ast.GetOrCreateNamespace(name)
	ab.schema.currentNamespace = name
	ab.schema.currentEntity = ""
	ab.schema.currentAction = ""
	return ab.schema
}

// CommonType defines a common (reusable) type in the current namespace.
func (ab *ActionBuilder) CommonType(name string, t Type) *ActionBuilder {
	ns := ab.schema.ast.GetOrCreateNamespace(ab.namespace)
	ns.CommonTypes[name] = t.toAST()
	return ab
}

// Schema returns the underlying Schema, ending the builder chain.
func (ab *ActionBuilder) Schema() *Schema {
	return ab.schema
}

// Resolve converts the schema to a ResolvedSchema.
// This is a convenience method that calls Schema().Resolve().
func (ab *ActionBuilder) Resolve() (*ResolvedSchema, error) {
	return ab.schema.Resolve()
}

// MustResolve converts the schema to a ResolvedSchema, panicking on error.
// This is a convenience method that calls Schema().MustResolve().
func (ab *ActionBuilder) MustResolve() *ResolvedSchema {
	return ab.schema.MustResolve()
}

// MarshalJSON serializes the schema to JSON format.
// This is a convenience method that calls Schema().MarshalJSON().
func (ab *ActionBuilder) MarshalJSON() ([]byte, error) {
	return ab.schema.MarshalJSON()
}

// MarshalJSONIndent serializes the schema to indented JSON format.
// This is a convenience method that calls Schema().MarshalJSONIndent().
func (ab *ActionBuilder) MarshalJSONIndent(prefix, indent string) ([]byte, error) {
	return ab.schema.MarshalJSONIndent(prefix, indent)
}

// MarshalCedar serializes the schema to human-readable Cedar format.
// This is a convenience method that calls Schema().MarshalCedar().
func (ab *ActionBuilder) MarshalCedar() ([]byte, error) {
	return ab.schema.MarshalCedar()
}
