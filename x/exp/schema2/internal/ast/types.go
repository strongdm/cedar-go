// Package ast defines internal AST types for unresolved Cedar schemas.
// These types are used by the builder and converted to resolved types via resolution.
package ast

// Schema represents an unresolved Cedar schema with namespace definitions.
type Schema struct {
	Namespaces map[string]*Namespace // key is namespace name, "" for anonymous
}

// Namespace contains entity types, actions, and common types within a namespace.
type Namespace struct {
	Name        string
	Entities    map[string]*EntityType
	Actions     map[string]*Action
	CommonTypes map[string]Type
}

// EntityType represents an unresolved entity type definition.
type EntityType struct {
	Name     string
	MemberOf []string    // parent entity type names (unqualified or qualified)
	Shape    *RecordType // nil if no attributes
	Tags     Type        // nil if no tags
	Enum     []string    // non-nil for enum entities, contains enum values
}

// Action represents an unresolved action definition.
type Action struct {
	Name      string
	MemberOf  []ActionRef // parent action groups
	AppliesTo *AppliesTo  // nil if action doesn't apply to anything
}

// ActionRef references an action, possibly in another namespace.
type ActionRef struct {
	Namespace string // empty for same namespace
	Name      string
}

// AppliesTo defines what principal and resource types an action applies to.
type AppliesTo struct {
	Principals []string // entity type names (unqualified or qualified)
	Resources  []string // entity type names (unqualified or qualified)
	Context    Type     // nil for empty context, usually RecordType
}

// Type represents a Cedar type in the schema.
// Implementations: PrimitiveType, EntityRefType, SetType, RecordType, CommonTypeRef
type Type interface {
	isType()
}

// PrimitiveType represents built-in Cedar types.
type PrimitiveType struct {
	Name string // "String", "Long", "Bool"
}

func (PrimitiveType) isType() {}

// EntityRefType references an entity type.
type EntityRefType struct {
	Name string // entity type name (unqualified or qualified)
}

func (EntityRefType) isType() {}

// SetType represents Set<T>.
type SetType struct {
	Element Type
}

func (SetType) isType() {}

// RecordType represents a record with named attributes.
type RecordType struct {
	Attributes map[string]*Attribute
}

func (RecordType) isType() {}

// CommonTypeRef references a common type definition.
type CommonTypeRef struct {
	Name string // common type name (unqualified or qualified)
}

func (CommonTypeRef) isType() {}

// ExtensionType represents an extension type like ipaddr or decimal.
type ExtensionType struct {
	Name string // e.g., "ipaddr", "decimal"
}

func (ExtensionType) isType() {}

// Attribute represents a field in a record type.
type Attribute struct {
	Name     string
	Type     Type
	Required bool
}

// NewSchema creates a new empty schema.
func NewSchema() *Schema {
	return &Schema{
		Namespaces: make(map[string]*Namespace),
	}
}

// GetOrCreateNamespace returns the namespace with the given name, creating it if needed.
func (s *Schema) GetOrCreateNamespace(name string) *Namespace {
	if ns, ok := s.Namespaces[name]; ok {
		return ns
	}
	ns := &Namespace{
		Name:        name,
		Entities:    make(map[string]*EntityType),
		Actions:     make(map[string]*Action),
		CommonTypes: make(map[string]Type),
	}
	s.Namespaces[name] = ns
	return ns
}

// GetOrCreateEntity returns the entity type with the given name, creating it if needed.
func (ns *Namespace) GetOrCreateEntity(name string) *EntityType {
	if et, ok := ns.Entities[name]; ok {
		return et
	}
	et := &EntityType{
		Name: name,
	}
	ns.Entities[name] = et
	return et
}

// GetOrCreateAction returns the action with the given name, creating it if needed.
func (ns *Namespace) GetOrCreateAction(name string) *Action {
	if a, ok := ns.Actions[name]; ok {
		return a
	}
	a := &Action{
		Name: name,
	}
	ns.Actions[name] = a
	return a
}
