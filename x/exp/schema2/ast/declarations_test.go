package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestCommonType(t *testing.T) {
	t.Parallel()
	want := ast.CommonTypeNode{
		Name: "PersonType",
		Type: ast.RecordType{
			Pairs: []ast.Pair{
				{Key: "name", Type: ast.StringType{}, Optional: false},
			},
		},
	}
	got := ast.CommonType("PersonType", ast.Record(ast.Attribute("name", ast.String())))
	testutil.Equals(t, got, want)
}

func TestCommonTypeAnnotate(t *testing.T) {
	t.Parallel()
	want := ast.CommonTypeNode{
		Name: "PersonType",
		Type: ast.StringType{},
		Annotations: []ast.Annotation{
			{Key: "doc", Value: "A person type"},
		},
	}
	got := ast.CommonType("PersonType", ast.String()).Annotate("doc", "A person type")
	testutil.Equals(t, got, want)
}

func TestEntity(t *testing.T) {
	t.Parallel()
	want := ast.EntityNode{
		Name: "User",
	}
	got := ast.Entity("User")
	testutil.Equals(t, got, want)
}

func TestEntityMemberOf(t *testing.T) {
	t.Parallel()
	want := ast.EntityNode{
		Name:        "User",
		MemberOfVal: []ast.EntityTypeRef{{Name: "Group"}},
	}
	got := ast.Entity("User").MemberOf(ast.Ref("Group"))
	testutil.Equals(t, got, want)
}

func TestEntityShape(t *testing.T) {
	t.Parallel()
	shape := ast.RecordType{
		Pairs: []ast.Pair{
			{Key: "name", Type: ast.StringType{}, Optional: false},
		},
	}
	want := ast.EntityNode{
		Name:     "User",
		ShapeVal: &shape,
	}
	got := ast.Entity("User").Shape(ast.Attribute("name", ast.String()))
	testutil.Equals(t, got, want)
}

func TestEntityTags(t *testing.T) {
	t.Parallel()
	want := ast.EntityNode{
		Name:    "User",
		TagsVal: ast.StringType{},
	}
	got := ast.Entity("User").Tags(ast.String())
	testutil.Equals(t, got, want)
}

func TestEntityAnnotate(t *testing.T) {
	t.Parallel()
	want := ast.EntityNode{
		Name: "User",
		Annotations: []ast.Annotation{
			{Key: "doc", Value: "User entity"},
		},
	}
	got := ast.Entity("User").Annotate("doc", "User entity")
	testutil.Equals(t, got, want)
}

func TestEnum(t *testing.T) {
	t.Parallel()
	want := ast.EnumNode{
		Name:   "Status",
		Values: []types.String{"active", "inactive"},
	}
	got := ast.Enum("Status", "active", "inactive")
	testutil.Equals(t, got, want)
}

func TestEnumAnnotate(t *testing.T) {
	t.Parallel()
	want := ast.EnumNode{
		Name:   "Status",
		Values: []types.String{"active"},
		Annotations: []ast.Annotation{
			{Key: "doc", Value: "Status enum"},
		},
	}
	got := ast.Enum("Status", "active").Annotate("doc", "Status enum")
	testutil.Equals(t, got, want)
}

func TestEnumEntityUIDs(t *testing.T) {
	t.Parallel()
	enum := ast.Enum("Status", "active", "inactive")
	var got []types.EntityUID
	for uid := range enum.EntityUIDs() {
		got = append(got, uid)
	}
	want := []types.EntityUID{
		types.NewEntityUID("Status", "active"),
		types.NewEntityUID("Status", "inactive"),
	}
	testutil.Equals(t, got, want)
}

func TestUID(t *testing.T) {
	t.Parallel()
	want := ast.EntityRef{
		Type: ast.EntityTypeRef{Name: "Action"},
		ID:   "view",
	}
	got := ast.UID("view")
	testutil.Equals(t, got, want)
}

func TestEntityUID(t *testing.T) {
	t.Parallel()
	want := ast.EntityRef{
		Type: ast.EntityTypeRef{Name: "User"},
		ID:   "alice",
	}
	got := ast.EntityUID("User", "alice")
	testutil.Equals(t, got, want)
}

func TestAction(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
	}
	got := ast.Action("view")
	testutil.Equals(t, got, want)
}

func TestActionMemberOf(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
		MemberOfVal: []ast.EntityRef{
			{Type: ast.EntityTypeRef{Name: "Action"}, ID: "read"},
		},
	}
	got := ast.Action("view").MemberOf(ast.UID("read"))
	testutil.Equals(t, got, want)
}

func TestActionPrincipal(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
		AppliesToVal: &ast.AppliesTo{
			PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
		},
	}
	got := ast.Action("view").Principal(ast.Ref("User"))
	testutil.Equals(t, got, want)
}

func TestActionResource(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
		AppliesToVal: &ast.AppliesTo{
			ResourceTypes: []ast.EntityTypeRef{{Name: "Photo"}},
		},
	}
	got := ast.Action("view").Resource(ast.Ref("Photo"))
	testutil.Equals(t, got, want)
}

func TestActionContext(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
		AppliesToVal: &ast.AppliesTo{
			Context: ast.RecordType{},
		},
	}
	got := ast.Action("view").Context(ast.Record())
	testutil.Equals(t, got, want)
}

func TestActionAnnotate(t *testing.T) {
	t.Parallel()
	want := ast.ActionNode{
		Name: "view",
		Annotations: []ast.Annotation{
			{Key: "doc", Value: "View action"},
		},
	}
	got := ast.Action("view").Annotate("doc", "View action")
	testutil.Equals(t, got, want)
}
