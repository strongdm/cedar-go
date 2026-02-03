package schema2

import (
	"github.com/cedar-policy/cedar-go/types"
)

// ResolvedSchema represents a fully-resolved Cedar schema with qualified names
// and computed entity hierarchies. ResolvedSchema is immutable.
type ResolvedSchema struct {
	entityTypes map[types.EntityType]*ResolvedEntityType
	actions     map[types.EntityUID]*ResolvedAction
}

// EntityTypes returns an iterator over all entity types in the schema.
func (rs *ResolvedSchema) EntityTypes() func(yield func(types.EntityType, *ResolvedEntityType) bool) {
	return func(yield func(types.EntityType, *ResolvedEntityType) bool) {
		for k, v := range rs.entityTypes {
			if !yield(k, v) {
				return
			}
		}
	}
}

// EntityType returns the resolved entity type for the given fully-qualified name.
// Returns nil if the entity type is not found.
func (rs *ResolvedSchema) EntityType(name types.EntityType) *ResolvedEntityType {
	return rs.entityTypes[name]
}

// LookupEntityType returns the resolved entity type for the given name and
// a boolean indicating whether it was found. This is useful when you need
// to distinguish between "not found" and "found with nil value".
func (rs *ResolvedSchema) LookupEntityType(name types.EntityType) (*ResolvedEntityType, bool) {
	et, ok := rs.entityTypes[name]
	return et, ok
}

// Actions returns an iterator over all actions in the schema.
func (rs *ResolvedSchema) Actions() func(yield func(types.EntityUID, *ResolvedAction) bool) {
	return func(yield func(types.EntityUID, *ResolvedAction) bool) {
		for k, v := range rs.actions {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Action returns the resolved action for the given EntityUID.
// Returns nil if the action is not found.
func (rs *ResolvedSchema) Action(uid types.EntityUID) *ResolvedAction {
	return rs.actions[uid]
}

// LookupAction returns the resolved action for the given EntityUID and
// a boolean indicating whether it was found. This is useful when you need
// to distinguish between "not found" and "found with nil value".
func (rs *ResolvedSchema) LookupAction(uid types.EntityUID) (*ResolvedAction, bool) {
	a, ok := rs.actions[uid]
	return a, ok
}

// ResolvedEntityType represents a fully-resolved entity type.
type ResolvedEntityType struct {
	name        types.EntityType
	descendants []types.EntityType  // entity types that can be members of this type (computed TC)
	attributes  *ResolvedRecordType // nil if no shape
	tags        ResolvedType        // nil if no tags
	kind        ResolvedEntityTypeKind
}

// Name returns the fully-qualified entity type name.
func (ret *ResolvedEntityType) Name() types.EntityType {
	return ret.name
}

// Descendants returns a copy of the entity types that can be members of this type.
// This is the transitive closure computed from memberOf relationships.
func (ret *ResolvedEntityType) Descendants() []types.EntityType {
	if ret.descendants == nil {
		return nil
	}
	result := make([]types.EntityType, len(ret.descendants))
	copy(result, ret.descendants)
	return result
}

// HasDescendant returns true if the given entity type is a descendant.
func (ret *ResolvedEntityType) HasDescendant(et types.EntityType) bool {
	for _, d := range ret.descendants {
		if d == et {
			return true
		}
	}
	return false
}

// Attributes returns the record type for entity attributes, or nil if no shape.
func (ret *ResolvedEntityType) Attributes() *ResolvedRecordType {
	return ret.attributes
}

// Tags returns the tag type for this entity, or nil if no tags.
func (ret *ResolvedEntityType) Tags() ResolvedType {
	return ret.tags
}

// Kind returns whether this is a standard or enum entity type.
func (ret *ResolvedEntityType) Kind() ResolvedEntityTypeKind {
	return ret.kind
}

// IsEnum returns true if this entity type is an enumerated type.
func (ret *ResolvedEntityType) IsEnum() bool {
	_, ok := ret.kind.(EnumEntityType)
	return ok
}

// AsEnum returns the EnumEntityType and true if this is an enum type,
// or a zero value and false if it's not an enum type.
func (ret *ResolvedEntityType) AsEnum() (EnumEntityType, bool) {
	e, ok := ret.kind.(EnumEntityType)
	return e, ok
}

// ResolvedEntityTypeKind indicates whether an entity is standard or enumerated.
type ResolvedEntityTypeKind interface {
	isResolvedEntityTypeKind()
}

// StandardEntityType represents a normal (non-enum) entity type.
type StandardEntityType struct{}

func (StandardEntityType) isResolvedEntityTypeKind() {}

// EnumEntityType represents an enumerated entity type with fixed values.
type EnumEntityType struct {
	values []types.EntityUID
}

func (EnumEntityType) isResolvedEntityTypeKind() {}

// Values returns a copy of the enum values as EntityUIDs.
func (e EnumEntityType) Values() []types.EntityUID {
	result := make([]types.EntityUID, len(e.values))
	copy(result, e.values)
	return result
}

// ResolvedAction represents a fully-resolved action.
type ResolvedAction struct {
	name      types.EntityUID
	memberOf  []types.EntityUID   // parent action groups
	appliesTo *ResolvedAppliesTo  // nil if action doesn't apply
	context   *ResolvedRecordType // nil for empty context
}

// Name returns the action's EntityUID.
func (ra *ResolvedAction) Name() types.EntityUID {
	return ra.name
}

// MemberOf returns a copy of the parent action group UIDs.
func (ra *ResolvedAction) MemberOf() []types.EntityUID {
	if ra.memberOf == nil {
		return nil
	}
	result := make([]types.EntityUID, len(ra.memberOf))
	copy(result, ra.memberOf)
	return result
}

// AppliesTo returns what this action applies to, or nil.
func (ra *ResolvedAction) AppliesTo() *ResolvedAppliesTo {
	return ra.appliesTo
}

// Context returns the context record type, or nil for empty context.
func (ra *ResolvedAction) Context() *ResolvedRecordType {
	return ra.context
}

// ResolvedAppliesTo defines what principal and resource types an action applies to.
type ResolvedAppliesTo struct {
	principals []types.EntityType
	resources  []types.EntityType
}

// Principals returns a copy of the principal entity types.
func (rat *ResolvedAppliesTo) Principals() []types.EntityType {
	result := make([]types.EntityType, len(rat.principals))
	copy(result, rat.principals)
	return result
}

// Resources returns a copy of the resource entity types.
func (rat *ResolvedAppliesTo) Resources() []types.EntityType {
	result := make([]types.EntityType, len(rat.resources))
	copy(result, rat.resources)
	return result
}

// ResolvedType represents a resolved Cedar type.
// All type references have been expanded (common types inlined).
type ResolvedType interface {
	isResolvedType()
}

// ResolvedPrimitiveType represents a built-in Cedar type.
type ResolvedPrimitiveType struct {
	name string // "String", "Long", "Bool"
}

func (ResolvedPrimitiveType) isResolvedType() {}

// Name returns the primitive type name.
func (rpt ResolvedPrimitiveType) Name() string {
	return rpt.name
}

// Primitive type constants
var (
	ResolvedStringType = ResolvedPrimitiveType{name: "String"}
	ResolvedLongType   = ResolvedPrimitiveType{name: "Long"}
	ResolvedBoolType   = ResolvedPrimitiveType{name: "Bool"}
)

// ResolvedEntityRefType references a resolved entity type.
type ResolvedEntityRefType struct {
	name types.EntityType
}

func (ResolvedEntityRefType) isResolvedType() {}

// Name returns the fully-qualified entity type name.
func (rert ResolvedEntityRefType) Name() types.EntityType {
	return rert.name
}

// ResolvedSetType represents Set<T>.
type ResolvedSetType struct {
	element ResolvedType
}

func (ResolvedSetType) isResolvedType() {}

// Element returns the set's element type.
func (rst ResolvedSetType) Element() ResolvedType {
	return rst.element
}

// ResolvedRecordType represents a record with named attributes.
type ResolvedRecordType struct {
	attributes map[string]*ResolvedAttribute
}

func (ResolvedRecordType) isResolvedType() {}

// Attributes returns an iterator over all attributes.
func (rrt *ResolvedRecordType) Attributes() func(yield func(string, *ResolvedAttribute) bool) {
	return func(yield func(string, *ResolvedAttribute) bool) {
		for k, v := range rrt.attributes {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Attribute returns the attribute with the given name, or nil.
func (rrt *ResolvedRecordType) Attribute(name string) *ResolvedAttribute {
	return rrt.attributes[name]
}

// ResolvedExtensionType represents an extension type.
type ResolvedExtensionType struct {
	name string // e.g., "ipaddr", "decimal"
}

func (ResolvedExtensionType) isResolvedType() {}

// Name returns the extension type name.
func (ret ResolvedExtensionType) Name() string {
	return ret.name
}

// ResolvedAttribute represents a field in a resolved record type.
type ResolvedAttribute struct {
	name     string
	typ      ResolvedType
	required bool
}

// Name returns the attribute name.
func (ra *ResolvedAttribute) Name() string {
	return ra.name
}

// Type returns the attribute's resolved type.
func (ra *ResolvedAttribute) Type() ResolvedType {
	return ra.typ
}

// Required returns whether this attribute is required.
func (ra *ResolvedAttribute) Required() bool {
	return ra.required
}

// Type assertion helpers for ResolvedType.
// These provide a convenient alternative to type switches.

// AsPrimitive returns the ResolvedPrimitiveType and true if t is a primitive type,
// or a zero value and false otherwise.
func AsPrimitive(t ResolvedType) (ResolvedPrimitiveType, bool) {
	p, ok := t.(ResolvedPrimitiveType)
	return p, ok
}

// AsEntityRef returns the ResolvedEntityRefType and true if t is an entity reference,
// or a zero value and false otherwise.
func AsEntityRef(t ResolvedType) (ResolvedEntityRefType, bool) {
	e, ok := t.(ResolvedEntityRefType)
	return e, ok
}

// AsSet returns the ResolvedSetType and true if t is a set type,
// or a zero value and false otherwise.
func AsSet(t ResolvedType) (ResolvedSetType, bool) {
	s, ok := t.(ResolvedSetType)
	return s, ok
}

// AsRecord returns the ResolvedRecordType and true if t is a record type,
// or nil and false otherwise.
func AsRecord(t ResolvedType) (*ResolvedRecordType, bool) {
	r, ok := t.(*ResolvedRecordType)
	return r, ok
}

// AsExtension returns the ResolvedExtensionType and true if t is an extension type,
// or a zero value and false otherwise.
func AsExtension(t ResolvedType) (ResolvedExtensionType, bool) {
	e, ok := t.(ResolvedExtensionType)
	return e, ok
}
