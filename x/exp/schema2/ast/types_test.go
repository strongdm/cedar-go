package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestString(t *testing.T) {
	t.Parallel()
	want := ast.StringType{}
	got := ast.String()
	testutil.Equals(t, got, want)
}

func TestLong(t *testing.T) {
	t.Parallel()
	want := ast.LongType{}
	got := ast.Long()
	testutil.Equals(t, got, want)
}

func TestBool(t *testing.T) {
	t.Parallel()
	want := ast.BoolType{}
	got := ast.Bool()
	testutil.Equals(t, got, want)
}

func TestIPAddr(t *testing.T) {
	t.Parallel()
	want := ast.ExtensionType{Name: "ipaddr"}
	got := ast.IPAddr()
	testutil.Equals(t, got, want)
}

func TestDecimal(t *testing.T) {
	t.Parallel()
	want := ast.ExtensionType{Name: "decimal"}
	got := ast.Decimal()
	testutil.Equals(t, got, want)
}

func TestDatetime(t *testing.T) {
	t.Parallel()
	want := ast.ExtensionType{Name: "datetime"}
	got := ast.Datetime()
	testutil.Equals(t, got, want)
}

func TestDuration(t *testing.T) {
	t.Parallel()
	want := ast.ExtensionType{Name: "duration"}
	got := ast.Duration()
	testutil.Equals(t, got, want)
}

func TestSet(t *testing.T) {
	t.Parallel()
	want := ast.SetType{Element: ast.StringType{}}
	got := ast.Set(ast.String())
	testutil.Equals(t, got, want)
}

func TestAttribute(t *testing.T) {
	t.Parallel()
	want := ast.Pair{Key: "name", Type: ast.StringType{}, Optional: false}
	got := ast.Attribute("name", ast.String())
	testutil.Equals(t, got, want)
}

func TestOptional(t *testing.T) {
	t.Parallel()
	want := ast.Pair{Key: "email", Type: ast.StringType{}, Optional: true}
	got := ast.Optional("email", ast.String())
	testutil.Equals(t, got, want)
}

func TestRecord(t *testing.T) {
	t.Parallel()
	want := ast.RecordType{
		Pairs: []ast.Pair{
			{Key: "name", Type: ast.StringType{}, Optional: false},
			{Key: "age", Type: ast.LongType{}, Optional: true},
		},
	}
	got := ast.Record(
		ast.Attribute("name", ast.String()),
		ast.Optional("age", ast.Long()),
	)
	testutil.Equals(t, got, want)
}

func TestEntityType(t *testing.T) {
	t.Parallel()
	want := ast.EntityTypeRef{Name: "User"}
	got := ast.EntityType("User")
	testutil.Equals(t, got, want)
}

func TestRef(t *testing.T) {
	t.Parallel()
	want := ast.EntityTypeRef{Name: "Photo"}
	got := ast.Ref("Photo")
	testutil.Equals(t, got, want)
}

func TestType(t *testing.T) {
	t.Parallel()
	want := ast.TypeRef{Name: "Common::Name"}
	got := ast.Type("Common::Name")
	testutil.Equals(t, got, want)
}
