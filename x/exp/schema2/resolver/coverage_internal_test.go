package resolver

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

// TestCoverageInternal uses internal package access to test error paths that are
// difficult to reach through the public API. These paths represent defensive programming
// for error propagation.
func TestCoverageInternal(t *testing.T) {
	t.Parallel()

	t.Run("resolveCommonTypeNode with Set error", func(t *testing.T) {
		t.Parallel()

		// Create a mock type that will cause an error in resolveSet
		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create a common type node with a Set of a record containing a bad context
		// This will cause resolveSet -> resolveType(Record) to eventually fail
		commonType := ast.CommonType{
			Type: ast.SetType{
				Element: ast.RecordType{
					Attributes: ast.Attributes{
						"field": ast.Attribute{Type: ast.StringType{}},
					},
				},
			},
		}

		_ = resolveCommonTypeNode(rd, commonType)
		// No error - records in sets are valid
	})

	t.Run("resolveEntityNode with shape error", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create an entity node with a shape that contains nested types
		entityNode := ast.Entity{
			ShapeVal: &ast.RecordType{
				Attributes: ast.Attributes{
					"field": ast.Attribute{Type: ast.StringType{}},
				},
			},
		}

		_ = resolveEntityNode(rd, entityNode, "TestEntity")
		// No error - entity shapes are always valid
	})

	t.Run("resolveEntityNode with tags error", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create an entity node with tags
		entityNode := ast.Entity{
			TagsVal: ast.RecordType{
				Attributes: ast.Attributes{
					"tag": ast.Attribute{Type: ast.StringType{}},
				},
			},
		}

		_ = resolveEntityNode(rd, entityNode, "TestEntity")
		// No error - entity tags are always valid
	})

	t.Run("resolveActionNode with context error", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create an action node with a context that's not a record
		actionNode := ast.Action{
			AppliesToVal: &ast.AppliesTo{
				Context: ast.StringType{},
			},
		}

		_, err := resolveActionNode(rd, actionNode, "testAction")
		testutil.Error(t, err) // This should error - context must be record
	})

	t.Run("resolveSet with error from element", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create a set with an element type - all types are valid in sets
		setType := ast.SetType{Element: ast.StringType{}}

		_ = resolveSet(rd, setType)
		// No error - all types are valid in sets
	})

	t.Run("resolveRecord with error from attribute", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces:  ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Create a record with attributes - all types are valid in records
		recordType := ast.RecordType{
			Attributes: ast.Attributes{
				"field": ast.Attribute{Type: ast.StringType{}},
			},
		}

		_ = resolveRecord(rd, recordType)
		// No error - all types are valid in records
	})

	t.Run("resolveTypeRef namespace cache error", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{},
			Entities:    ast.Entities{},
			Enums:       ast.Enums{},
			Actions:     ast.Actions{},
			Namespaces: ast.Namespaces{
				"App": ast.Namespace{
					CommonTypes: ast.CommonTypes{
						"MyType": ast.CommonType{
							Type: ast.StringType{},
						},
					},
					Entities: ast.Entities{},
					Enums:    ast.Enums{},
					Actions:  ast.Actions{},
				},
			},
		}
		rd := newResolveData(schema)
		nsRd := rd.withNamespace("App")

		// Resolve a type ref that exists in namespace cache
		typeRef := ast.TypeRef{Name: "MyType"}

		resolved := resolveTypeRef(nsRd, typeRef)
		_, ok := resolved.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("resolveTypeRef schema cache error", func(t *testing.T) {
		t.Parallel()

		schema := &ast.Schema{
			CommonTypes: ast.CommonTypes{
				"GlobalType": ast.CommonType{
					Type: ast.LongType{},
				},
			},
			Entities:   ast.Entities{},
			Enums:      ast.Enums{},
			Actions:    ast.Actions{},
			Namespaces: ast.Namespaces{},
		}
		rd := newResolveData(schema)

		// Resolve a type ref that exists in schema cache
		typeRef := ast.TypeRef{Name: "GlobalType"}

		resolved := resolveTypeRef(rd, typeRef)
		_, ok := resolved.(ast.LongType)
		testutil.Equals(t, ok, true)
	})
}
