package ast

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestIsNode(t *testing.T) {
	t.Parallel()
	ScopeNode{}.isScope()

	StrOpNode{}.isNode()
	BinaryNode{}.isNode()
	NodeTypeIfThenElse{}.isNode()
	NodeTypeLike{}.isNode()
	NodeTypeIs{}.isNode()
	UnaryNode{}.isNode()
	NodeTypeExtensionCall{}.isNode()
	NodeValue{}.isNode()
	NodeTypeRecord{}.isNode()
	NodeTypeSet{}.isNode()
	NodeTypeVariable{}.isNode()

}

func TestAsNode(t *testing.T) {
	t.Parallel()
	n := NewNode(NodeValue{Value: types.Long(42)})
	v := n.AsIsNode()
	testutil.Equals(t, v, (IsNode)(NodeValue{Value: types.Long(42)}))
}

func TestIsScope(t *testing.T) {
	t.Parallel()
	ScopeNode{}.isScope()
	PrincipalScopeNode{}.isPrincipalScope()
	ActionScopeNode{}.isActionScope()
	ResourceScopeNode{}.isResourceScope()
}

func TestStripNodes(t *testing.T) {
	t.Parallel()
	t.Run("preserveNil", func(t *testing.T) {
		t.Parallel()
		out := stripNodes(nil)
		testutil.Equals(t, out, nil)
	})
	t.Run("preserveNonNil", func(t *testing.T) {
		t.Parallel()
		out := stripNodes([]Node{})
		testutil.Equals(t, out, []IsNode{})
	})
}
