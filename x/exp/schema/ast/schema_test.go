package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestUID(t *testing.T) {
	t.Parallel()
	want := ast.ParentRef{
		ID: "view",
	}
	got := ast.ParentRefFromID("view")
	testutil.Equals(t, got, want)
}

func TestEntityUID(t *testing.T) {
	t.Parallel()
	want := ast.ParentRef{
		Type: ast.EntityTypeRef("User"),
		ID:   "alice",
	}
	got := ast.NewParentRef("User", "alice")
	testutil.Equals(t, got, want)
}
