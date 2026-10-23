package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/ast"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestNodeErr(t *testing.T) {
	t.Parallel()

	good := ast.Principal().Is("User")
	testutil.OK(t, good.Err())
	testutil.Equals(t, good.Must(), good)

	bad := ast.Principal().Is("Nobody }; permit ( principal, action, resource ) when { true").And(ast.True())
	testutil.Error(t, bad.Err())
	testutil.Panic(t, func() { bad.Must() })
}

func TestPolicyErr(t *testing.T) {
	t.Parallel()

	good := ast.Annotation("id", "a").
		Permit().
		PrincipalIs("User").
		ActionEq(types.NewEntityUID("Action", "view")).
		When(ast.Resource().Is("Document"))
	testutil.OK(t, good.Err())
	testutil.Equals(t, good.Must(), good)

	tests := []struct {
		name   string
		policy *ast.Policy
	}{
		{"scope", ast.Permit().PrincipalIs("Nobody, action, resource ); permit ( principal")},
		{"condition", ast.Permit().When(ast.Principal().Is("Nobody }; permit ( principal, action, resource ) when { true"))},
		{"annotation", ast.Annotation("x(\"y\") permit ( principal, action, resource ); @z", "v").Forbid()},
		{"annotate", ast.Forbid().Annotate("a b", "v")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Error(t, tt.policy.Err())
			testutil.Panic(t, func() { tt.policy.Must() })
		})
	}
}
