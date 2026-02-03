package resolver

import (
	"iter"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

type Annotations ast.Annotations

// Schema is a Cedar schema with resolved types and indexed declarations.
// Common types are inlined. All names are fully qualified.
type Schema struct {
	Namespaces map[types.Path]Namespace
	Entities   map[types.EntityType]Entity
	Enums      map[types.EntityType]Enum
	Actions    map[types.EntityUID]Action
}

type Namespace struct {
	Name        types.Path
	Annotations Annotations
}

type Entity struct {
	Name        types.EntityType
	Annotations Annotations
	MemberOf    []types.EntityType
	Shape       *RecordType
	Tags        IsType
}

type Enum struct {
	Name        types.EntityType
	Annotations Annotations
	Values      []types.String
}

// EntityUIDs iterates over valid EntityUIDs for this enum.
func (e Enum) EntityUIDs() iter.Seq[types.EntityUID] {
	return func(yield func(types.EntityUID) bool) {
		for _, v := range e.Values {
			if !yield(types.NewEntityUID(e.Name, v)) {
				return
			}
		}
	}
}

type AppliesTo struct {
	Principals []types.EntityType
	Resources  []types.EntityType
	Context    RecordType
}

// Action defines what principals can do to resources.
// If AppliesTo is nil, the action never applies.
type Action struct {
	Name        types.String
	Annotations Annotations
	MemberOf    []types.EntityUID
	AppliesTo   *AppliesTo
}
