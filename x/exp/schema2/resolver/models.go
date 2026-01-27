package resolver

import (
	"iter"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// ResolvedSchema represents a schema with all type references resolved and indexed for efficient lookup.
type ResolvedSchema struct {
	Namespaces map[types.Path]ResolvedNamespace    // Namespace path -> ResolvedNamespace
	Entities   map[types.EntityType]ResolvedEntity // Fully qualified entity type -> ResolvedEntity
	Enums      map[types.EntityType]ResolvedEnum   // Fully qualified entity type -> ResolvedEnum
	Actions    map[types.EntityUID]ResolvedAction  // Fully qualified action UID -> ResolvedAction
}

// ResolvedNamespace represents a namespace without the declarations included.
// All declarations have been moved into the other maps.
type ResolvedNamespace struct {
	Name        types.Path
	Annotations ast.Annotations
}

// ResolvedEntity represents an entity type with all type references fully resolved.
// All EntityTypeRef references have been converted to types.EntityType.
type ResolvedEntity struct {
	Name        types.EntityType   // Fully qualified entity type
	Annotations ast.Annotations    // Entity annotations
	MemberOf    []types.EntityType // Fully qualified parent entity types
	Shape       *ast.RecordType    // Entity shape (with all type references resolved)
	Tags        ast.IsType         // Tags type (with all type references resolved)
}

// ResolvedEnum represents an enum type with all references fully resolved.
type ResolvedEnum struct {
	Name        types.EntityType // Fully qualified enum type
	Annotations ast.Annotations  // Enum annotations
	Values      []types.String   // Enum values
}

// EntityUIDs returns an iterator over EntityUID values for each enum value.
// The Name field should already be fully qualified.
func (e ResolvedEnum) EntityUIDs() iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(e.Name, v)) {
				return
			}
		}
	}
}

// ResolvedAppliesTo represents the appliesTo clause with all type references fully resolved.
// All EntityTypeRef references have been converted to types.EntityType.
type ResolvedAppliesTo struct {
	PrincipalTypes []types.EntityType // Fully qualified principal entity types
	ResourceTypes  []types.EntityType // Fully qualified resource entity types
	Context        ast.RecordType     // Context type (with all type references resolved)
}

// ResolvedAction represents an action with all type references fully resolved.
// All EntityTypeRef and EntityRef references have been converted to types.EntityType and types.EntityUID.
type ResolvedAction struct {
	Name        types.String       // Action name (local, not qualified)
	Annotations ast.Annotations    // Action annotations
	MemberOf    []types.EntityUID  // Fully qualified parent action UIDs
	AppliesTo   *ResolvedAppliesTo // AppliesTo clause with all type references resolved
}
