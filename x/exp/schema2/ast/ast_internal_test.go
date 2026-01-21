package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

// TestNodeTypePriorityDefault tests the default case of nodeTypePriority
func TestNodeTypePriorityDefault(t *testing.T) {
	t.Parallel()

	// Test with NamespaceNode which should hit the default case
	ns := NamespaceNode{Name: "Test"}
	priority := nodeTypePriority(ns)
	testutil.Equals(t, 99, priority)
}

// TestNodeNameDefault tests the default case of nodeName
func TestNodeNameDefault(t *testing.T) {
	t.Parallel()

	// Test with NamespaceNode which should hit the default case
	ns := NamespaceNode{Name: "Test"}
	name := nodeName(ns)
	testutil.Equals(t, "", name)
}

// TestSortNodesWithNamespace tests sortNodes with a NamespaceNode
// This is a defensive test to ensure the default cases work correctly
func TestSortNodesWithNamespace(t *testing.T) {
	t.Parallel()

	// Create a mixed list including a namespace (unusual but should handle gracefully)
	nodes := []IsNode{
		EntityNode{Name: "Entity2"},
		NamespaceNode{Name: "Namespace"},
		EntityNode{Name: "Entity1"},
	}

	sortNodes(nodes)

	// Entities should come first (priority 1), then namespace (priority 99)
	e1, ok := nodes[0].(EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.EntityType("Entity1"), e1.Name)

	e2, ok := nodes[1].(EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.EntityType("Entity2"), e2.Name)

	ns, ok := nodes[2].(NamespaceNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.Path("Namespace"), ns.Name)
}

// TestSortDeclarationsComplete tests sortDeclarations with all types
func TestSortDeclarationsComplete(t *testing.T) {
	t.Parallel()

	// Create a mixed list of declarations
	decls := []IsDeclaration{
		ActionNode{Name: "Action2"},
		EnumNode{Name: "Enum1"},
		EntityNode{Name: "Entity2"},
		ActionNode{Name: "Action1"},
		EntityNode{Name: "Entity1"},
	}

	sortDeclarations(decls)

	// Should be sorted by type (entities, enums, actions) then by name
	e1, ok := decls[0].(EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.EntityType("Entity1"), e1.Name)

	e2, ok := decls[1].(EntityNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.EntityType("Entity2"), e2.Name)

	enum, ok := decls[2].(EnumNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.EntityType("Enum1"), enum.Name)

	a1, ok := decls[3].(ActionNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.String("Action1"), a1.Name)

	a2, ok := decls[4].(ActionNode)
	testutil.Equals(t, true, ok)
	testutil.Equals(t, types.String("Action2"), a2.Name)
}
