package schema

import "github.com/cedar-policy/cedar-go/types"

// Schema represents a Cedar schema containing entity types, action definitions,
// and common type declarations organized into namespaces.
type Schema struct {
	// Namespaces contains all namespace definitions in this schema.
	// The empty string key represents the anonymous (root) namespace.
	Namespaces map[types.Path]*Namespace

	// filename is used for error messages in parsing
	filename string
}

// SetFilename sets the filename for error messages in parsing.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// New creates a new empty schema.
func New() *Schema {
	return &Schema{
		Namespaces: make(map[types.Path]*Namespace),
	}
}

// GetNamespace returns the namespace with the given name, or nil if not found.
func (s *Schema) GetNamespace(name types.Path) *Namespace {
	return s.Namespaces[name]
}

// AddNamespace adds a namespace to the schema. If a namespace with the same name
// already exists, it will be replaced.
func (s *Schema) AddNamespace(ns *Namespace) *Schema {
	s.Namespaces[ns.Name] = ns
	return s
}

// AddEntity adds an entity type to the anonymous namespace.
func (s *Schema) AddEntity(e *EntityDecl) *Schema {
	ns := s.ensureNamespace("")
	ns.AddEntity(e)
	return s
}

// AddAction adds an action to the anonymous namespace.
func (s *Schema) AddAction(a *ActionDecl) *Schema {
	ns := s.ensureNamespace("")
	ns.AddAction(a)
	return s
}

// AddCommonType adds a common type to the anonymous namespace.
func (s *Schema) AddCommonType(c *CommonTypeDecl) *Schema {
	ns := s.ensureNamespace("")
	ns.AddCommonType(c)
	return s
}

// ensureNamespace ensures a namespace exists and returns it.
func (s *Schema) ensureNamespace(name types.Path) *Namespace {
	if s.Namespaces == nil {
		s.Namespaces = make(map[types.Path]*Namespace)
	}
	ns, ok := s.Namespaces[name]
	if !ok {
		ns = NewNamespace(name)
		s.Namespaces[name] = ns
	}
	return ns
}

// Namespace represents a Cedar namespace containing entity types, actions,
// and common type definitions.
type Namespace struct {
	Name        types.Path
	Annotations Annotations
	Entities    map[types.Ident]*EntityDecl
	Actions     map[types.String]*ActionDecl
	CommonTypes map[types.Ident]*CommonTypeDecl
}

// NewNamespace creates a new namespace with the given name.
func NewNamespace(name types.Path) *Namespace {
	return &Namespace{
		Name:        name,
		Entities:    make(map[types.Ident]*EntityDecl),
		Actions:     make(map[types.String]*ActionDecl),
		CommonTypes: make(map[types.Ident]*CommonTypeDecl),
	}
}

// GetEntity returns the entity type with the given name, or nil if not found.
func (ns *Namespace) GetEntity(name types.Ident) *EntityDecl {
	return ns.Entities[name]
}

// GetAction returns the action with the given name, or nil if not found.
func (ns *Namespace) GetAction(name types.String) *ActionDecl {
	return ns.Actions[name]
}

// GetCommonType returns the common type with the given name, or nil if not found.
func (ns *Namespace) GetCommonType(name types.Ident) *CommonTypeDecl {
	return ns.CommonTypes[name]
}

// AddEntity adds an entity type to the namespace.
func (ns *Namespace) AddEntity(e *EntityDecl) *Namespace {
	ns.Entities[e.Name] = e
	return ns
}

// AddAction adds an action to the namespace.
func (ns *Namespace) AddAction(a *ActionDecl) *Namespace {
	ns.Actions[a.Name] = a
	return ns
}

// AddCommonType adds a common type to the namespace.
func (ns *Namespace) AddCommonType(c *CommonTypeDecl) *Namespace {
	ns.CommonTypes[c.Name] = c
	return ns
}

// Annotate adds an annotation to the namespace.
func (ns *Namespace) Annotate(key types.Ident, value types.String) *Namespace {
	ns.Annotations = ns.Annotations.Set(key, value)
	return ns
}

// EntityDecl represents an entity type declaration in a schema.
type EntityDecl struct {
	Name          types.Ident
	Annotations   Annotations
	MemberOfTypes []types.Path // Parent entity types this entity can be a member of
	Attributes    []Attribute  // Attributes of the entity (shape)
	Tags          Type         // Optional tags type
	Enum          []types.String // If non-nil, this is an enum entity with these values
}

// NewEntity creates a new entity type declaration.
func NewEntity(name types.Ident) *EntityDecl {
	return &EntityDecl{
		Name: name,
	}
}

// MemberOf sets the parent entity types this entity can be a member of.
func (e *EntityDecl) MemberOf(parents ...types.Path) *EntityDecl {
	e.MemberOfTypes = append(e.MemberOfTypes, parents...)
	return e
}

// SetAttributes sets the attributes (shape) of the entity.
func (e *EntityDecl) SetAttributes(attrs ...Attribute) *EntityDecl {
	e.Attributes = attrs
	return e
}

// SetTags sets the tags type for the entity.
func (e *EntityDecl) SetTags(t Type) *EntityDecl {
	e.Tags = t
	return e
}

// SetEnum sets the enum values for the entity, making it an enum entity.
func (e *EntityDecl) SetEnum(values ...types.String) *EntityDecl {
	e.Enum = values
	return e
}

// Annotate adds an annotation to the entity.
func (e *EntityDecl) Annotate(key types.Ident, value types.String) *EntityDecl {
	e.Annotations = e.Annotations.Set(key, value)
	return e
}

// ActionDecl represents an action declaration in a schema.
type ActionDecl struct {
	Name           types.String
	Annotations    Annotations
	MemberOf       []ActionRef    // Parent action groups
	PrincipalTypes []types.Path   // Entity types that can be principals for this action
	ResourceTypes  []types.Path   // Entity types that can be resources for this action
	Context        Type           // Context type (nil if no context)
}

// ActionRef represents a reference to an action (possibly in another namespace).
type ActionRef struct {
	Namespace types.Path   // Optional namespace, empty if in same namespace
	Name      types.String // Action name
}

// NewAction creates a new action declaration.
func NewAction(name types.String) *ActionDecl {
	return &ActionDecl{
		Name: name,
	}
}

// InActions sets the parent action groups.
func (a *ActionDecl) InActions(refs ...ActionRef) *ActionDecl {
	a.MemberOf = append(a.MemberOf, refs...)
	return a
}

// SetPrincipalTypes sets the entity types that can be principals for this action.
func (a *ActionDecl) SetPrincipalTypes(types ...types.Path) *ActionDecl {
	a.PrincipalTypes = append(a.PrincipalTypes, types...)
	return a
}

// SetResourceTypes sets the entity types that can be resources for this action.
func (a *ActionDecl) SetResourceTypes(types ...types.Path) *ActionDecl {
	a.ResourceTypes = append(a.ResourceTypes, types...)
	return a
}

// SetContext sets the context type for this action.
func (a *ActionDecl) SetContext(t Type) *ActionDecl {
	a.Context = t
	return a
}

// Annotate adds an annotation to the action.
func (a *ActionDecl) Annotate(key types.Ident, value types.String) *ActionDecl {
	a.Annotations = a.Annotations.Set(key, value)
	return a
}

// CommonTypeDecl represents a common (reusable) type declaration.
type CommonTypeDecl struct {
	Name        types.Ident
	Annotations Annotations
	Type        Type
}

// NewCommonType creates a new common type declaration.
func NewCommonType(name types.Ident, t Type) *CommonTypeDecl {
	return &CommonTypeDecl{
		Name: name,
		Type: t,
	}
}

// Annotate adds an annotation to the common type.
func (c *CommonTypeDecl) Annotate(key types.Ident, value types.String) *CommonTypeDecl {
	c.Annotations = c.Annotations.Set(key, value)
	return c
}

// Annotations is a collection of key-value annotation pairs.
type Annotations []Annotation

// Annotation represents a single annotation.
type Annotation struct {
	Key   types.Ident
	Value types.String
}

// Get returns the value for the given key, or empty string if not found.
func (a Annotations) Get(key types.Ident) (types.String, bool) {
	for _, ann := range a {
		if ann.Key == key {
			return ann.Value, true
		}
	}
	return "", false
}

// Set sets or updates an annotation value.
func (a Annotations) Set(key types.Ident, value types.String) Annotations {
	for i, ann := range a {
		if ann.Key == key {
			a[i].Value = value
			return a
		}
	}
	return append(a, Annotation{Key: key, Value: value})
}
