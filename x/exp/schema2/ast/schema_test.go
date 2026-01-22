package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestNewSchema(t *testing.T) {
	t.Parallel()
	entity := ast.Entity("User")
	action := ast.Action("view")
	want := &ast.Schema{
		Nodes: []ast.IsNode{entity, action},
	}
	got := ast.NewSchema(entity, action)
	testutil.Equals(t, got, want)
}

func TestNamespace(t *testing.T) {
	t.Parallel()
	entity := ast.Entity("User")
	action := ast.Action("view")
	want := ast.NamespaceNode{
		Name:         "MyApp",
		Declarations: []ast.IsDeclaration{entity, action},
	}
	got := ast.Namespace("MyApp", entity, action)
	testutil.Equals(t, got, want)
}

func TestNamespaceAnnotate(t *testing.T) {
	t.Parallel()
	want := ast.NamespaceNode{
		Name: "MyApp",
		Annotations: []ast.Annotation{
			{Key: "doc", Value: "MyApp namespace"},
		},
	}
	got := ast.Namespace("MyApp").Annotate("doc", "MyApp namespace")
	testutil.Equals(t, got, want)
}

func TestNamespaceCommonTypes(t *testing.T) {
	t.Parallel()
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	entity := ast.Entity("User")
	ns := ast.Namespace("MyApp", ct1, entity, ct2)

	var got []ast.CommonTypeNode
	for ct := range ns.CommonTypes() {
		got = append(got, ct)
	}
	want := []ast.CommonTypeNode{ct1, ct2}
	testutil.Equals(t, got, want)
}

func TestNamespaceEntities(t *testing.T) {
	t.Parallel()
	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	ct := ast.CommonType("Type1", ast.String())
	ns := ast.Namespace("MyApp", e1, ct, e2)

	var got []ast.EntityNode
	for e := range ns.Entities() {
		got = append(got, e)
	}
	want := []ast.EntityNode{e1, e2}
	testutil.Equals(t, got, want)
}

func TestNamespaceEnums(t *testing.T) {
	t.Parallel()
	en1 := ast.Enum("Status", "active")
	en2 := ast.Enum("Role", "admin")
	entity := ast.Entity("User")
	ns := ast.Namespace("MyApp", en1, entity, en2)

	var got []ast.EnumNode
	for e := range ns.Enums() {
		got = append(got, e)
	}
	want := []ast.EnumNode{en1, en2}
	testutil.Equals(t, got, want)
}

func TestNamespaceActions(t *testing.T) {
	t.Parallel()
	a1 := ast.Action("view")
	a2 := ast.Action("edit")
	entity := ast.Entity("User")
	ns := ast.Namespace("MyApp", a1, entity, a2)

	var got []ast.ActionNode
	for a := range ns.Actions() {
		got = append(got, a)
	}
	want := []ast.ActionNode{a1, a2}
	testutil.Equals(t, got, want)
}

func TestSchemaNamespaces(t *testing.T) {
	t.Parallel()
	ns1 := ast.Namespace("App1")
	ns2 := ast.Namespace("App2")
	entity := ast.Entity("User")
	schema := ast.NewSchema(ns1, entity, ns2)

	var got []ast.NamespaceNode
	for ns := range schema.Namespaces() {
		got = append(got, ns)
	}
	want := []ast.NamespaceNode{ns1, ns2}
	testutil.Equals(t, got, want)
}

func TestSchemaCommonTypes(t *testing.T) {
	t.Parallel()
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	ct3 := ast.CommonType("Type3", ast.Bool())
	ns := ast.Namespace("MyApp", ct2)
	schema := ast.NewSchema(ct1, ns, ct3)

	var got []ast.CommonTypeNode
	for _, ct := range schema.CommonTypes() {
		got = append(got, ct)
	}
	want := []ast.CommonTypeNode{ct1, ct2, ct3}
	testutil.Equals(t, got, want)
}

func TestSchemaEntities(t *testing.T) {
	t.Parallel()
	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	e3 := ast.Entity("Photo")
	ns := ast.Namespace("MyApp", e2)
	schema := ast.NewSchema(e1, ns, e3)

	var got []ast.EntityNode
	for _, e := range schema.Entities() {
		got = append(got, e)
	}
	want := []ast.EntityNode{e1, e2, e3}
	testutil.Equals(t, got, want)
}

func TestSchemaEnums(t *testing.T) {
	t.Parallel()
	en1 := ast.Enum("Status", "active")
	en2 := ast.Enum("Role", "admin")
	en3 := ast.Enum("Level", "high")
	ns := ast.Namespace("MyApp", en2)
	schema := ast.NewSchema(en1, ns, en3)

	var got []ast.EnumNode
	for _, e := range schema.Enums() {
		got = append(got, e)
	}
	want := []ast.EnumNode{en1, en2, en3}
	testutil.Equals(t, got, want)
}

func TestSchemaActions(t *testing.T) {
	t.Parallel()
	a1 := ast.Action("view")
	a2 := ast.Action("edit")
	a3 := ast.Action("delete")
	ns := ast.Namespace("MyApp", a2)
	schema := ast.NewSchema(a1, ns, a3)

	var got []ast.ActionNode
	for _, a := range schema.Actions() {
		got = append(got, a)
	}
	want := []ast.ActionNode{a1, a2, a3}
	testutil.Equals(t, got, want)
}

func TestEnumEntityUIDsEarlyReturn(t *testing.T) {
	t.Parallel()
	enum := ast.Enum("Status", "active", "inactive", "pending")
	count := 0
	for range enum.EntityUIDs() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestNamespaceCommonTypesEarlyReturn(t *testing.T) {
	t.Parallel()
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	ns := ast.Namespace("MyApp", ct1, ct2)

	count := 0
	for range ns.CommonTypes() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestNamespaceEntitiesEarlyReturn(t *testing.T) {
	t.Parallel()
	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	ns := ast.Namespace("MyApp", e1, e2)

	count := 0
	for range ns.Entities() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestNamespaceEnumsEarlyReturn(t *testing.T) {
	t.Parallel()
	en1 := ast.Enum("Status", "active")
	en2 := ast.Enum("Role", "admin")
	ns := ast.Namespace("MyApp", en1, en2)

	count := 0
	for range ns.Enums() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestNamespaceActionsEarlyReturn(t *testing.T) {
	t.Parallel()
	a1 := ast.Action("view")
	a2 := ast.Action("edit")
	ns := ast.Namespace("MyApp", a1, a2)

	count := 0
	for range ns.Actions() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaNamespacesEarlyReturn(t *testing.T) {
	t.Parallel()
	ns1 := ast.Namespace("App1")
	ns2 := ast.Namespace("App2")
	schema := ast.NewSchema(ns1, ns2)

	count := 0
	for range schema.Namespaces() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaCommonTypesEarlyReturn(t *testing.T) {
	t.Parallel()
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	ct3 := ast.CommonType("Type3", ast.Bool())
	ns := ast.Namespace("MyApp", ct2)
	schema := ast.NewSchema(ct1, ns, ct3)

	count := 0
	for range schema.CommonTypes() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaCommonTypesEarlyReturnInNamespace(t *testing.T) {
	t.Parallel()
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	ct3 := ast.CommonType("Type3", ast.Bool())
	ns := ast.Namespace("MyApp", ct1, ct2)
	schema := ast.NewSchema(ns, ct3)

	count := 0
	for range schema.CommonTypes() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaEntitiesEarlyReturn(t *testing.T) {
	t.Parallel()
	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	e3 := ast.Entity("Photo")
	ns := ast.Namespace("MyApp", e2)
	schema := ast.NewSchema(e1, ns, e3)

	count := 0
	for range schema.Entities() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaEntitiesEarlyReturnInNamespace(t *testing.T) {
	t.Parallel()
	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	e3 := ast.Entity("Photo")
	ns := ast.Namespace("MyApp", e1, e2)
	schema := ast.NewSchema(ns, e3)

	count := 0
	for range schema.Entities() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaEnumsEarlyReturn(t *testing.T) {
	t.Parallel()
	en1 := ast.Enum("Status", "active")
	en2 := ast.Enum("Role", "admin")
	en3 := ast.Enum("Level", "high")
	ns := ast.Namespace("MyApp", en2)
	schema := ast.NewSchema(en1, ns, en3)

	count := 0
	for range schema.Enums() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaEnumsEarlyReturnInNamespace(t *testing.T) {
	t.Parallel()
	en1 := ast.Enum("Status", "active")
	en2 := ast.Enum("Role", "admin")
	en3 := ast.Enum("Level", "high")
	ns := ast.Namespace("MyApp", en1, en2)
	schema := ast.NewSchema(ns, en3)

	count := 0
	for range schema.Enums() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaActionsEarlyReturn(t *testing.T) {
	t.Parallel()
	a1 := ast.Action("view")
	a2 := ast.Action("edit")
	a3 := ast.Action("delete")
	ns := ast.Namespace("MyApp", a2)
	schema := ast.NewSchema(a1, ns, a3)

	count := 0
	for range schema.Actions() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}

func TestSchemaActionsEarlyReturnInNamespace(t *testing.T) {
	t.Parallel()
	a1 := ast.Action("view")
	a2 := ast.Action("edit")
	a3 := ast.Action("delete")
	ns := ast.Namespace("MyApp", a1, a2)
	schema := ast.NewSchema(ns, a3)

	count := 0
	for range schema.Actions() {
		count++
		if count == 1 {
			break
		}
	}
	testutil.Equals(t, count, 1)
}
