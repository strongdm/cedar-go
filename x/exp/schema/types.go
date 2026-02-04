// Package schema provides Cedar schema parsing, construction, and resolution.
//
// Cedar schemas define the structure of entity types, actions, and common types
// that applications use for authorization. This package supports:
//   - Parsing schemas from JSON and Cedar text formats
//   - Programmatic schema construction using a builder pattern
//   - Resolution to fully-qualified types (types.EntityType, types.EntityUID)
//
// Example usage:
//
//	// Parse from Cedar text
//	var s schema.Schema
//	err := s.UnmarshalCedar(schemaBytes)
//
//	// Resolve to get fully-qualified types
//	resolved, err := s.Resolve()
//
//	// Access resolved entity types
//	userType := resolved.Namespaces["MyApp"].EntityTypes[types.EntityType("MyApp::User")]
package schema

// Schema represents a parsed but unresolved Cedar schema.
// Type references in a Schema are raw strings that may or may not be qualified.
// Call Resolve() to get a resolved.Schema with fully-qualified type references.
type Schema struct {
	Namespaces map[string]*Namespace // "" key = empty namespace
	filename   string                // for error messages
}

// Namespace contains entity types, actions, common types, and enum types within a namespace.
type Namespace struct {
	EntityTypes map[string]*EntityTypeDef
	EnumTypes   map[string]*EnumTypeDef
	Actions     map[string]*ActionDef
	CommonTypes map[string]*CommonTypeDef
	Annotations Annotations
}

// EntityTypeDef describes an entity type in the schema.
type EntityTypeDef struct {
	MemberOfTypes []string    // parent entity type names (unresolved)
	Shape         *RecordType // attribute definitions
	Tags          Type        // optional tag type
	Annotations   Annotations
}

// EnumTypeDef describes an enumerated entity type in the schema.
// Enum types define a fixed set of entity IDs that can exist for this type.
type EnumTypeDef struct {
	Values      []string // the allowed entity IDs
	Annotations Annotations
}

// ActionDef describes an action in the schema.
type ActionDef struct {
	MemberOf    []*ActionRef // action group membership
	AppliesTo   *AppliesTo   // principal/resource/context
	Annotations Annotations
}

// ActionRef references an action, possibly in another namespace.
type ActionRef struct {
	Type string // action entity type (e.g., "MyNamespace::Action"), empty for same namespace
	ID   string // action name
}

// AppliesTo specifies what principals and resources an action applies to.
type AppliesTo struct {
	PrincipalTypes []string    // entity types that can be principals
	ResourceTypes  []string    // entity types that can be resources
	Context        *RecordType // context record type (inline definition)
	ContextRef     Type        // context type reference (common type name); mutually exclusive with Context
}

// CommonTypeDef is a named type alias.
type CommonTypeDef struct {
	Type        Type
	Annotations Annotations
}

// Annotations are key-value metadata attached to schema elements.
type Annotations map[string]string

// newAnnotations returns an initialized Annotations map.
func newAnnotations() Annotations {
	return make(Annotations)
}

// Type represents a Cedar type in the schema.
// Use the constructor functions (Long, String, Bool, Set, Record, Entity, Extension,
// IPAddr, Decimal, Datetime, Duration, CommonType) to create Type values.
//
// Implementations: PrimitiveType, SetType, RecordType, EntityRef, ExtensionType, CommonTypeRef, EntityOrCommonRef
type Type interface {
	isType()
}

// PrimitiveKind represents Cedar primitive types.
type PrimitiveKind int

const (
	PrimitiveLong PrimitiveKind = iota
	PrimitiveString
	PrimitiveBool
)

func (k PrimitiveKind) String() string {
	switch k {
	case PrimitiveLong:
		return "Long"
	case PrimitiveString:
		return "String"
	case PrimitiveBool:
		return "Bool"
	default:
		return "Unknown"
	}
}

// PrimitiveType represents Long, String, or Bool.
type PrimitiveType struct {
	Kind PrimitiveKind
}

func (PrimitiveType) isType() {}

// SetType represents Set<T>.
type SetType struct {
	Element Type
}

func (SetType) isType() {}

// RecordType represents a record with named attributes.
type RecordType struct {
	Attributes map[string]*Attribute
}

func (*RecordType) isType() {}

// Attribute is a named field in a record type.
type Attribute struct {
	Type        Type
	Required    bool
	Annotations Annotations
}

// EntityRef references an entity type by name.
type EntityRef struct {
	Name string // possibly qualified (e.g., "MyNamespace::User")
}

func (EntityRef) isType() {}

// ExtensionType represents an extension type like ipaddr or decimal.
type ExtensionType struct {
	Name string // e.g., "ipaddr", "decimal", "datetime", "duration"
}

func (ExtensionType) isType() {}

// CommonTypeRef references a common type by name.
type CommonTypeRef struct {
	Name string // possibly qualified
}

func (CommonTypeRef) isType() {}

// EntityOrCommonRef is an ambiguous reference that could be an entity or common type.
// This is resolved during the Resolve() phase.
type EntityOrCommonRef struct {
	Name string
}

func (EntityOrCommonRef) isType() {}

// SetFilename sets the filename for error messages.
func (s *Schema) SetFilename(filename string) {
	s.filename = filename
}

// Filename returns the filename set for error messages.
func (s *Schema) Filename() string {
	return s.filename
}
