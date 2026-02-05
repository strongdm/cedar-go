package schema

// Annotations are key-value metadata attached to schema elements.
type Annotations map[string]string

func newAnnotations() Annotations {
	return make(Annotations)
}

// Namespace contains entity types, actions, common types, and enum types within a namespace.
type Namespace struct {
	EntityTypes map[string]*EntityTypeDef
	EnumTypes   map[string]*EnumTypeDef
	Actions     map[string]*ActionDef
	CommonTypes map[string]*CommonTypeDef
	Annotations Annotations
}

func newNamespace() *Namespace {
	return &Namespace{
		EntityTypes: make(map[string]*EntityTypeDef),
		EnumTypes:   make(map[string]*EnumTypeDef),
		Actions:     make(map[string]*ActionDef),
		CommonTypes: make(map[string]*CommonTypeDef),
		Annotations: newAnnotations(),
	}
}

// EntityTypeDef describes an entity type in the schema.
type EntityTypeDef struct {
	MemberOfTypes []string
	Shape         *RecordTypeExpr
	Tags          TypeExpr
	Annotations   Annotations
}

// EnumTypeDef describes an enumerated entity type with a fixed set of entity IDs.
type EnumTypeDef struct {
	Values      []string
	Annotations Annotations
}

// ActionDef describes an action in the schema.
type ActionDef struct {
	MemberOf    []*ActionRef
	AppliesTo   *AppliesTo
	Annotations Annotations
}

// ActionRef references an action, possibly in another namespace.
type ActionRef struct {
	Type string // namespace-qualified action entity type (e.g., "MyNS::Action"), empty for same namespace
	ID   string
}

// AppliesTo specifies what principals and resources an action applies to.
type AppliesTo struct {
	PrincipalTypes []string
	ResourceTypes  []string
	Context        TypeExpr // RecordTypeExpr or TypeNameExpr
}

// CommonTypeDef is a named type alias.
type CommonTypeDef struct {
	Type        TypeExpr
	Annotations Annotations
}
