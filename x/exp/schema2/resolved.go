// Package schema2 provides a programmatic interface for Cedar schemas.
//
// This package provides two distinct schema representations:
//   - Schema: A mutable builder type for constructing schemas programmatically
//   - ResolvedSchema: An immutable type with fully-qualified names and computed hierarchies
//
// Example usage:
//
//	schema := schema2.NewSchema().
//	    Namespace("MyApp").
//	        Entity("User").In("Group").
//	        Entity("Group").
//	        Action("read").Principals("User").Resources("Document")
//
//	resolved, err := schema.Resolve()
//	if err != nil {
//	    // Handle validation errors
//	}
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
