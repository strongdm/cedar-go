package ast

import (
	"iter"

	"github.com/cedar-policy/cedar-go/types"
)

// NamespaceNode represents a Cedar namespace declaration.
type NamespaceNode struct {
	Name         types.Path
	Annotations  []Annotation
	Declarations []IsDeclaration
}

func (NamespaceNode) isNode() { _ = 0 }

// Namespace creates a new NamespaceNode with the given path and declarations.
func Namespace(path types.Path, decls ...IsDeclaration) NamespaceNode {
	return NamespaceNode{
		Name:         path,
		Declarations: decls,
	}
}

// Annotate adds an annotation to the namespace and returns the node for chaining.
func (n NamespaceNode) Annotate(key types.Ident, value types.String) NamespaceNode {
	n.Annotations = append(n.Annotations, Annotation{Key: key, Value: value})
	return n
}

// Resolve returns a new NamespaceNode with all type references in its declarations resolved.
func (n NamespaceNode) Resolve(rd *resolveData) (NamespaceNode, error) {
	resolved := NamespaceNode{
		Name:        n.Name,
		Annotations: n.Annotations,
	}

	// Create resolve data with this namespace
	nsRd := rd.withNamespace(&n)

	if len(n.Declarations) > 0 {
		resolved.Declarations = make([]IsDeclaration, len(n.Declarations))
		for i, decl := range n.Declarations {
			switch d := decl.(type) {
			case CommonTypeNode:
				resolvedDecl, err := d.Resolve(nsRd)
				if err != nil {
					return NamespaceNode{}, err
				}
				resolved.Declarations[i] = resolvedDecl
			case EntityNode:
				resolvedDecl, err := d.Resolve(nsRd)
				if err != nil {
					return NamespaceNode{}, err
				}
				resolved.Declarations[i] = resolvedDecl
			case EnumNode:
				resolved.Declarations[i] = d.Resolve(nsRd)
			case ActionNode:
				resolvedDecl, err := d.Resolve(nsRd)
				if err != nil {
					return NamespaceNode{}, err
				}
				resolved.Declarations[i] = resolvedDecl
			}
		}
	}

	return resolved, nil
}

// commonTypes returns an iterator over all CommonTypeNode declarations in the namespace.
func (n NamespaceNode) commonTypes() iter.Seq[CommonTypeNode] {
	return func(yield func(CommonTypeNode) bool) {
		for _, decl := range n.Declarations {
			if ct, ok := decl.(CommonTypeNode); ok {
				if !yield(ct) {
					return
				}
			}
		}
	}
}

// entities returns an iterator over all EntityNode declarations in the namespace.
func (n NamespaceNode) entities() iter.Seq[EntityNode] {
	return func(yield func(EntityNode) bool) {
		for _, decl := range n.Declarations {
			if e, ok := decl.(EntityNode); ok {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// enums returns an iterator over all EnumNode declarations in the namespace.
func (n NamespaceNode) enums() iter.Seq[EnumNode] {
	return func(yield func(EnumNode) bool) {
		for _, decl := range n.Declarations {
			if e, ok := decl.(EnumNode); ok {
				if !yield(e) {
					return
				}
			}
		}
	}
}

// actions returns an iterator over all ActionNode declarations in the namespace.
func (n NamespaceNode) actions() iter.Seq[ActionNode] {
	return func(yield func(ActionNode) bool) {
		for _, decl := range n.Declarations {
			if a, ok := decl.(ActionNode); ok {
				if !yield(a) {
					return
				}
			}
		}
	}
}

// CommonTypeNode represents a Cedar common type declaration (type alias).
type CommonTypeNode struct {
	Name        types.Ident
	Annotations []Annotation
	Type        IsType
}

func (CommonTypeNode) isNode()        { _ = 0 }
func (CommonTypeNode) isDeclaration() { _ = 0 }

// CommonType creates a new CommonTypeNode with the given name and type.
func CommonType(name types.Ident, t IsType) CommonTypeNode {
	return CommonTypeNode{
		Name: name,
		Type: t,
	}
}

// Annotate adds an annotation to the common type and returns the node for chaining.
func (c CommonTypeNode) Annotate(key types.Ident, value types.String) CommonTypeNode {
	c.Annotations = append(c.Annotations, Annotation{Key: key, Value: value})
	return c
}

// FullName returns the fully qualified name for this common type.
// If namespace is nil, the type name is used as-is (top-level).
// If namespace is provided (e.g., "MyApp"), the name is "MyApp::TypeName".
func (c CommonTypeNode) FullName(namespace *NamespaceNode) types.Path {
	if namespace == nil {
		return types.Path(c.Name)
	}
	return types.Path(string(namespace.Name) + "::" + string(c.Name))
}

// Resolve returns a new CommonTypeNode with all type references resolved.
func (c CommonTypeNode) Resolve(rd *resolveData) (CommonTypeNode, error) {
	resolvedType, err := c.Type.resolve(rd)
	if err != nil {
		return CommonTypeNode{}, err
	}
	return CommonTypeNode{
		Name:        c.Name,
		Annotations: c.Annotations,
		Type:        resolvedType,
	}, nil
}

// EntityNode represents a Cedar entity type declaration.
type EntityNode struct {
	Name        types.EntityType
	Annotations []Annotation
	MemberOfVal []EntityTypeRef
	ShapeVal    *RecordType
	TagsVal     IsType
}

func (EntityNode) isNode()        { _ = 0 }
func (EntityNode) isDeclaration() { _ = 0 }

// Entity creates a new EntityNode with the given name.
func Entity(name types.EntityType) EntityNode {
	return EntityNode{Name: name}
}

// MemberOf sets the entity types this entity can be a member of.
func (e EntityNode) MemberOf(parents ...EntityTypeRef) EntityNode {
	e.MemberOfVal = parents
	return e
}

// Shape sets the shape (attributes) of the entity.
func (e EntityNode) Shape(pairs ...Pair) EntityNode {
	r := Record(pairs...)
	e.ShapeVal = &r
	return e
}

// Tags sets the tags type for the entity.
func (e EntityNode) Tags(t IsType) EntityNode {
	e.TagsVal = t
	return e
}

// Annotate adds an annotation to the entity and returns the node for chaining.
func (e EntityNode) Annotate(key types.Ident, value types.String) EntityNode {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// EntityType returns the fully qualified entity type name for this entity.
// If namespace is nil, the entity name is used as-is (top-level).
// If namespace is provided (e.g., "MyApp"), the type is "MyApp::EntityName".
func (e EntityNode) EntityType(namespace *NamespaceNode) types.EntityType {
	if namespace == nil {
		return e.Name
	}
	return types.EntityType(string(namespace.Name) + "::" + string(e.Name))
}

// Resolve returns a new EntityNode with all type references resolved.
func (e EntityNode) Resolve(rd *resolveData) (EntityNode, error) {
	resolved := EntityNode{
		Name:        e.Name,
		Annotations: e.Annotations,
	}

	// Resolve MemberOf references
	if len(e.MemberOfVal) > 0 {
		resolved.MemberOfVal = make([]EntityTypeRef, len(e.MemberOfVal))
		for i, ref := range e.MemberOfVal {
			resolvedRef, err := ref.resolve(rd)
			if err != nil {
				return EntityNode{}, err
			}
			resolved.MemberOfVal[i] = resolvedRef.(EntityTypeRef)
		}
	}

	// Resolve Shape
	if e.ShapeVal != nil {
		resolvedShape, err := e.ShapeVal.resolve(rd)
		if err != nil {
			return EntityNode{}, err
		}
		recordType := resolvedShape.(RecordType)
		resolved.ShapeVal = &recordType
	}

	// Resolve Tags
	if e.TagsVal != nil {
		resolvedTags, err := e.TagsVal.resolve(rd)
		if err != nil {
			return EntityNode{}, err
		}
		resolved.TagsVal = resolvedTags
	}

	return resolved, nil
}

// EnumNode represents a Cedar enum entity type declaration.
type EnumNode struct {
	Name        types.EntityType
	Annotations []Annotation
	Values      []types.String
}

func (EnumNode) isNode()        { _ = 0 }
func (EnumNode) isDeclaration() { _ = 0 }

// Enum creates a new EnumNode with the given name and values.
func Enum(name types.EntityType, values ...types.String) EnumNode {
	return EnumNode{
		Name:   name,
		Values: values,
	}
}

// Annotate adds an annotation to the enum and returns the node for chaining.
func (e EnumNode) Annotate(key types.Ident, value types.String) EnumNode {
	e.Annotations = append(e.Annotations, Annotation{Key: key, Value: value})
	return e
}

// EntityUIDs returns an iterator over EntityUID values for each enum value.
// If namespace is nil, the entity type is used as-is (top-level).
// If namespace is provided (e.g., "MyApp"), the type is "MyApp::EnumName".
func (e EnumNode) EntityUIDs(namespace *NamespaceNode) iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		var entityType types.EntityType
		if namespace == nil {
			entityType = e.Name
		} else {
			entityType = types.EntityType(string(namespace.Name) + "::" + string(e.Name))
		}
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(entityType, v)) {
				return
			}
		}
	}
}

// Resolve returns the EnumNode unchanged (enums have no type references to resolve).
func (e EnumNode) Resolve(rd *resolveData) EnumNode {
	return e
}

// ActionNode represents a Cedar action declaration.
type ActionNode struct {
	Name         types.String
	Annotations  []Annotation
	MemberOfVal  []EntityRef
	AppliesToVal *AppliesTo
}

func (ActionNode) isNode()        { _ = 0 }
func (ActionNode) isDeclaration() { _ = 0 }

// AppliesTo represents the principal, resource, and context types for an action.
type AppliesTo struct {
	PrincipalTypes []EntityTypeRef
	ResourceTypes  []EntityTypeRef
	Context        IsType
}

// EntityRef represents a reference to a specific entity (type + id).
type EntityRef struct {
	Type EntityTypeRef
	ID   types.String
}

// UID creates an EntityRef with just an ID (type is inferred as Action).
func UID(id types.String) EntityRef {
	return EntityRef{
		Type: EntityTypeRef{Name: "Action"},
		ID:   id,
	}
}

// EntityUID creates an EntityRef with an explicit type and ID.
func EntityUID(typ types.EntityType, id types.String) EntityRef {
	return EntityRef{
		Type: EntityTypeRef{Name: typ},
		ID:   id,
	}
}

// Action creates a new ActionNode with the given name.
func Action(name types.String) ActionNode {
	return ActionNode{Name: name}
}

// MemberOf sets the action groups this action is a member of.
func (a ActionNode) MemberOf(refs ...EntityRef) ActionNode {
	a.MemberOfVal = refs
	return a
}

// Principal sets the principal types for the action.
func (a ActionNode) Principal(principals ...EntityTypeRef) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.PrincipalTypes = principals
	return a
}

// Resource sets the resource types for the action.
func (a ActionNode) Resource(resources ...EntityTypeRef) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.ResourceTypes = resources
	return a
}

// Context sets the context type for the action.
func (a ActionNode) Context(t IsType) ActionNode {
	if a.AppliesToVal == nil {
		a.AppliesToVal = &AppliesTo{}
	}
	a.AppliesToVal.Context = t
	return a
}

// Annotate adds an annotation to the action and returns the node for chaining.
func (a ActionNode) Annotate(key types.Ident, value types.String) ActionNode {
	a.Annotations = append(a.Annotations, Annotation{Key: key, Value: value})
	return a
}

// EntityUID returns the fully qualified EntityUID for this action.
// If namespace is nil, the type is "Action" (top-level).
// If namespace is provided (e.g., "Bananas"), the type is "Bananas::Action".
func (a ActionNode) EntityUID(namespace *NamespaceNode) types.EntityUID {
	var entityType types.EntityType
	if namespace == nil {
		entityType = "Action"
	} else {
		entityType = types.EntityType(string(namespace.Name) + "::Action")
	}
	return types.NewEntityUID(entityType, a.Name)
}

// Resolve returns a new ActionNode with all type references resolved.
func (a ActionNode) Resolve(rd *resolveData) (ActionNode, error) {
	resolved := ActionNode{
		Name:        a.Name,
		Annotations: a.Annotations,
	}

	// Resolve MemberOf references
	if len(a.MemberOfVal) > 0 {
		resolved.MemberOfVal = make([]EntityRef, len(a.MemberOfVal))
		for i, ref := range a.MemberOfVal {
			resolvedType, err := ref.Type.resolve(rd)
			if err != nil {
				return ActionNode{}, err
			}
			resolved.MemberOfVal[i] = EntityRef{
				Type: resolvedType.(EntityTypeRef),
				ID:   ref.ID,
			}
		}
	}

	// Resolve AppliesTo
	if a.AppliesToVal != nil {
		resolved.AppliesToVal = &AppliesTo{}

		if len(a.AppliesToVal.PrincipalTypes) > 0 {
			resolved.AppliesToVal.PrincipalTypes = make([]EntityTypeRef, len(a.AppliesToVal.PrincipalTypes))
			for i, ref := range a.AppliesToVal.PrincipalTypes {
				resolvedRef, err := ref.resolve(rd)
				if err != nil {
					return ActionNode{}, err
				}
				resolved.AppliesToVal.PrincipalTypes[i] = resolvedRef.(EntityTypeRef)
			}
		}

		if len(a.AppliesToVal.ResourceTypes) > 0 {
			resolved.AppliesToVal.ResourceTypes = make([]EntityTypeRef, len(a.AppliesToVal.ResourceTypes))
			for i, ref := range a.AppliesToVal.ResourceTypes {
				resolvedRef, err := ref.resolve(rd)
				if err != nil {
					return ActionNode{}, err
				}
				resolved.AppliesToVal.ResourceTypes[i] = resolvedRef.(EntityTypeRef)
			}
		}

		if a.AppliesToVal.Context != nil {
			resolvedContext, err := a.AppliesToVal.Context.resolve(rd)
			if err != nil {
				return ActionNode{}, err
			}
			resolved.AppliesToVal.Context = resolvedContext
		}
	}

	return resolved, nil
}
