package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

var (
	badType   = types.EntityType("A, action, resource ); permit ( principal")
	badUID    = types.NewEntityUID(badType, "x")
	goodUID   = types.NewEntityUID("A::B", "x")
	badNode   = ast.Value(badUID)
	validNode = ast.Long(1)
)

func TestNodeErrSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		node ast.Node
	}{
		{"Is", ast.Principal().Is(badType)},
		{"IsIn", ast.Principal().IsIn(badType, ast.Value(goodUID))},
		{"EntityUID", ast.EntityUID(types.Ident(badType), "x")},
		{"EntityUIDReserved", ast.EntityUID("if", "x")},
		{"Value", ast.Value(badUID)},
		{"ValueInSet", ast.Value(types.NewSet(goodUID, types.NewSet(badUID)))},
		{"ValueInRecord", ast.Value(types.NewRecord(types.RecordMap{"a": types.NewRecord(types.RecordMap{"b": badUID})}))},
		{"ExtensionCall", ast.ExtensionCall("ip(\"1.2.3.4\") || ip", ast.String("x"))},
		{"NewMethodCall", ast.NewMethodCall(validNode, "isIpv4() || true || x.isIpv4")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Error(t, tt.node.Err())
			testutil.Panic(t, func() { tt.node.Must() })
		})
	}
}

func TestNodeErrValid(t *testing.T) {
	t.Parallel()
	n := ast.Principal().Is("A::B").
		And(ast.Principal().IsIn("A", ast.Value(goodUID))).
		And(ast.EntityUID("A", "x").Equal(ast.Value(types.NewSet(goodUID, types.NewRecord(types.RecordMap{"a": goodUID}))))).
		And(ast.ExtensionCall("ip", ast.String("1.2.3.4")).IsIpv4()).
		And(ast.Context().Access("not an ident").Has("also not"))
	testutil.OK(t, n.Err())
	testutil.Equals(t, n.Must(), n)
}

func TestNodeErrPropagates(t *testing.T) {
	t.Parallel()
	binary := map[string]func(ast.Node, ast.Node) ast.Node{
		"Equal":                     ast.Node.Equal,
		"NotEqual":                  ast.Node.NotEqual,
		"LessThan":                  ast.Node.LessThan,
		"LessThanOrEqual":           ast.Node.LessThanOrEqual,
		"GreaterThan":               ast.Node.GreaterThan,
		"GreaterThanOrEqual":        ast.Node.GreaterThanOrEqual,
		"DecimalLessThan":           ast.Node.DecimalLessThan,
		"DecimalLessThanOrEqual":    ast.Node.DecimalLessThanOrEqual,
		"DecimalGreaterThan":        ast.Node.DecimalGreaterThan,
		"DecimalGreaterThanOrEqual": ast.Node.DecimalGreaterThanOrEqual,
		"And":                       ast.Node.And,
		"Or":                        ast.Node.Or,
		"Add":                       ast.Node.Add,
		"Subtract":                  ast.Node.Subtract,
		"Multiply":                  ast.Node.Multiply,
		"In":                        ast.Node.In,
		"Contains":                  ast.Node.Contains,
		"ContainsAll":               ast.Node.ContainsAll,
		"ContainsAny":               ast.Node.ContainsAny,
		"GetTag":                    ast.Node.GetTag,
		"HasTag":                    ast.Node.HasTag,
		"IsInRange":                 ast.Node.IsInRange,
		"Offset":                    ast.Node.Offset,
		"DurationSince":             ast.Node.DurationSince,
		"IsIn":                      func(l, r ast.Node) ast.Node { return l.IsIn("A", r) },
		"Set":                       func(l, r ast.Node) ast.Node { return ast.Set(l, r) },
		"Record":                    func(l, r ast.Node) ast.Node { return ast.Record(ast.Pairs{{Key: "a", Value: l}, {Key: "a", Value: r}}) },
		"ExtensionCall":             func(l, r ast.Node) ast.Node { return ast.ExtensionCall("decimal", l, r) },
		"NewMethodCall":             func(l, r ast.Node) ast.Node { return ast.NewMethodCall(l, "lessThan", r) },
		"IfThenElseCondThen":        func(l, r ast.Node) ast.Node { return ast.IfThenElse(l, r, validNode) },
		"IfThenElseThenElse":        func(l, r ast.Node) ast.Node { return ast.IfThenElse(validNode, l, r) },
	}
	for name, f := range binary {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testutil.OK(t, f(validNode, validNode).Err())
			testutil.Error(t, f(badNode, validNode).Err())
			testutil.Error(t, f(validNode, badNode).Err())
		})
	}

	unary := map[string]func(ast.Node) ast.Node{
		"Not":            ast.Not,
		"Negate":         ast.Negate,
		"Like":           func(n ast.Node) ast.Node { return n.Like(types.NewPattern("a")) },
		"Is":             func(n ast.Node) ast.Node { return n.Is("A") },
		"Access":         func(n ast.Node) ast.Node { return n.Access("a") },
		"Has":            func(n ast.Node) ast.Node { return n.Has("a") },
		"IsEmpty":        ast.Node.IsEmpty,
		"IsIpv4":         ast.Node.IsIpv4,
		"IsIpv6":         ast.Node.IsIpv6,
		"IsMulticast":    ast.Node.IsMulticast,
		"IsLoopback":     ast.Node.IsLoopback,
		"ToDate":         ast.Node.ToDate,
		"ToTime":         ast.Node.ToTime,
		"ToDays":         ast.Node.ToDays,
		"ToHours":        ast.Node.ToHours,
		"ToMinutes":      ast.Node.ToMinutes,
		"ToSeconds":      ast.Node.ToSeconds,
		"ToMilliseconds": ast.Node.ToMilliseconds,
	}
	for name, f := range unary {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			testutil.OK(t, f(validNode).Err())
			testutil.Error(t, f(badNode).Err())
		})
	}
}

func TestPolicyErrSources(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		policy *ast.Policy
	}{
		{"PrincipalEq", ast.Permit().PrincipalEq(badUID)},
		{"PrincipalIn", ast.Permit().PrincipalIn(badUID)},
		{"PrincipalIs", ast.Permit().PrincipalIs(badType)},
		{"PrincipalIsInType", ast.Permit().PrincipalIsIn(badType, goodUID)},
		{"PrincipalIsInEntity", ast.Permit().PrincipalIsIn("A", badUID)},
		{"ActionEq", ast.Permit().ActionEq(badUID)},
		{"ActionIn", ast.Permit().ActionIn(badUID)},
		{"ActionInSet", ast.Permit().ActionInSet(goodUID, badUID)},
		{"ResourceEq", ast.Forbid().ResourceEq(badUID)},
		{"ResourceIn", ast.Forbid().ResourceIn(badUID)},
		{"ResourceIs", ast.Forbid().ResourceIs(badType)},
		{"ResourceIsInType", ast.Forbid().ResourceIsIn(badType, goodUID)},
		{"ResourceIsInEntity", ast.Forbid().ResourceIsIn("A", badUID)},
		{"When", ast.Permit().When(badNode)},
		{"Unless", ast.Permit().Unless(badNode)},
		{"Annotate", ast.Permit().Annotate("a(\"b\") permit ( principal, action, resource ); @c", "v")},
		{"AnnotateEmpty", ast.Permit().Annotate("", "v")},
		{"AnnotationPermit", ast.Annotation("a b", "v").Permit()},
		{"AnnotationForbid", ast.Annotation("a", "v").Annotation("a b", "v").Forbid()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Error(t, tt.policy.Err())
			testutil.Panic(t, func() { tt.policy.Must() })
		})
	}
}

func TestPolicyErrValid(t *testing.T) {
	t.Parallel()
	p := ast.Annotation("if", "v").
		Annotation("id", "w").
		Permit().
		Annotate("_x1", "y").
		PrincipalIsIn("A::B", goodUID).
		ActionInSet(goodUID, types.NewEntityUID("Action", "a")).
		ResourceIs("C").
		When(ast.Principal().Is("A::B")).
		Unless(ast.Value(goodUID).In(ast.Resource()))
	testutil.OK(t, p.Err())
	testutil.Equals(t, p.Must(), p)
}

func TestPolicyErrSticky(t *testing.T) {
	t.Parallel()
	p := ast.Permit().PrincipalIs("bad name").ResourceIs("also bad").PrincipalIs("Good")
	testutil.Equals(t, p.Err().Error(), `invalid entity type: "bad name"`)
}
