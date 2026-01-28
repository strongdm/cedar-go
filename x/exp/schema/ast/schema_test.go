package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestUID(t *testing.T) {
	t.Parallel()
	want := ast.EntityRef{
		ID: "view",
	}
	got := ast.EntityRefFromID("view")
	testutil.Equals(t, got, want)
}

func TestEntityUID(t *testing.T) {
	t.Parallel()
	want := ast.EntityRef{
		Type: ast.EntityTypeRef{Name: "User"},
		ID:   "alice",
	}
	got := ast.NewEntityRef("User", "alice")
	testutil.Equals(t, got, want)
}
