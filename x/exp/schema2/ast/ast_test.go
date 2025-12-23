package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	t.Run("empty schema", func(t *testing.T) {
		t.Parallel()
		s := ast.NewSchema()
		testutil.Equals(t, len(s.Nodes), 0)
	})

	t.Run("schema with namespace", func(t *testing.T) {
		t.Parallel()
		ns := ast.Namespace("MyApp")
		s := ast.NewSchema(ns)
		testutil.Equals(t, len(s.Nodes), 1)
	})
}

func TestNamespace(t *testing.T) {
	t.Parallel()

	t.Run("basic namespace", func(t *testing.T) {
		t.Parallel()
		ns := ast.Namespace("MyApp::Accounts")
		testutil.Equals(t, ns.Name, types.Path("MyApp::Accounts"))
		testutil.Equals(t, len(ns.Declarations), 0)
		testutil.Equals(t, len(ns.Annotations), 0)
	})

	t.Run("namespace with declarations", func(t *testing.T) {
		t.Parallel()
		ns := ast.Namespace("MyApp",
			ast.Entity("User"),
			ast.Entity("Group"),
		)
		testutil.Equals(t, len(ns.Declarations), 2)
	})

	t.Run("namespace with annotations", func(t *testing.T) {
		t.Parallel()
		ns := ast.Namespace("MyApp").
			Annotate("doc", "My application namespace")
		testutil.Equals(t, len(ns.Annotations), 1)
		testutil.Equals(t, ns.Annotations[0].Key, types.Ident("doc"))
		testutil.Equals(t, ns.Annotations[0].Value, types.String("My application namespace"))
	})
}

func TestCommonType(t *testing.T) {
	t.Parallel()

	t.Run("basic common type", func(t *testing.T) {
		t.Parallel()
		ct := ast.CommonType("Name", ast.String())
		testutil.Equals(t, ct.Name, types.Ident("Name"))
		_, ok := ct.Type.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("common type with record", func(t *testing.T) {
		t.Parallel()
		ct := ast.CommonType("Address", ast.Record(
			ast.Attribute("street", ast.String()),
			ast.Optional("apt", ast.String()),
		))
		testutil.Equals(t, ct.Name, types.Ident("Address"))
		rt, ok := ct.Type.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rt.Pairs), 2)
	})

	t.Run("common type with annotation", func(t *testing.T) {
		t.Parallel()
		ct := ast.CommonType("Name", ast.String()).
			Annotate("doc", "A person's name")
		testutil.Equals(t, len(ct.Annotations), 1)
	})
}

func TestEntity(t *testing.T) {
	t.Parallel()

	t.Run("basic entity", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("User")
		testutil.Equals(t, e.Name, types.EntityType("User"))
		testutil.Equals(t, e.MemberOfVal, nil)
		testutil.Equals(t, e.ShapeVal, nil)
		testutil.Equals(t, e.TagsVal, nil)
	})

	t.Run("entity with memberOf", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("User").
			MemberOf(ast.EntityType("Group"), ast.EntityType("Team"))
		testutil.Equals(t, len(e.MemberOfVal), 2)
		testutil.Equals(t, e.MemberOfVal[0].Name, types.EntityType("Group"))
		testutil.Equals(t, e.MemberOfVal[1].Name, types.EntityType("Team"))
	})

	t.Run("entity with shape", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("User").
			Shape(
				ast.Attribute("name", ast.String()),
				ast.Optional("email", ast.String()),
			)
		testutil.Equals(t, e.ShapeVal != nil, true)
		testutil.Equals(t, len(e.ShapeVal.Pairs), 2)
		testutil.Equals(t, e.ShapeVal.Pairs[0].Key, types.String("name"))
		testutil.Equals(t, e.ShapeVal.Pairs[0].Optional, false)
		testutil.Equals(t, e.ShapeVal.Pairs[1].Key, types.String("email"))
		testutil.Equals(t, e.ShapeVal.Pairs[1].Optional, true)
	})

	t.Run("entity with tags", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("Document").
			Tags(ast.String())
		_, ok := e.TagsVal.(ast.StringType)
		testutil.Equals(t, ok, true)
	})

	t.Run("entity with annotation", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("User").
			Annotate("doc", "A user entity")
		testutil.Equals(t, len(e.Annotations), 1)
	})

	t.Run("entity with everything", func(t *testing.T) {
		t.Parallel()
		e := ast.Entity("User").
			MemberOf(ast.EntityType("Group")).
			Shape(ast.Attribute("name", ast.String())).
			Tags(ast.String()).
			Annotate("doc", "A user")
		testutil.Equals(t, len(e.MemberOfVal), 1)
		testutil.Equals(t, e.ShapeVal != nil, true)
		testutil.Equals(t, e.TagsVal != nil, true)
		testutil.Equals(t, len(e.Annotations), 1)
	})
}

func TestEnum(t *testing.T) {
	t.Parallel()

	t.Run("basic enum", func(t *testing.T) {
		t.Parallel()
		e := ast.Enum("Status", "active", "inactive", "pending")
		testutil.Equals(t, e.Name, types.EntityType("Status"))
		testutil.Equals(t, len(e.Values), 3)
		testutil.Equals(t, e.Values[0], types.String("active"))
		testutil.Equals(t, e.Values[1], types.String("inactive"))
		testutil.Equals(t, e.Values[2], types.String("pending"))
	})

	t.Run("enum with annotation", func(t *testing.T) {
		t.Parallel()
		e := ast.Enum("Status", "active", "inactive").
			Annotate("doc", "Status values")
		testutil.Equals(t, len(e.Annotations), 1)
	})

	t.Run("enum EntityUIDs", func(t *testing.T) {
		t.Parallel()
		e := ast.Enum("Status", "active", "inactive")
		uids := e.EntityUIDs()
		testutil.Equals(t, len(uids), 2)
		testutil.Equals(t, uids[0].Type, types.EntityType("Status"))
		testutil.Equals(t, uids[0].ID, types.String("active"))
		testutil.Equals(t, uids[1].Type, types.EntityType("Status"))
		testutil.Equals(t, uids[1].ID, types.String("inactive"))
	})
}

func TestAction(t *testing.T) {
	t.Parallel()

	t.Run("basic action", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view")
		testutil.Equals(t, a.Name, types.String("view"))
		testutil.Equals(t, a.MemberOfVal, nil)
		testutil.Equals(t, a.AppliesToVal, nil)
	})

	t.Run("action with memberOf", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			MemberOf(ast.UID("readOnly"), ast.EntityUID("Action", "allActions"))
		testutil.Equals(t, len(a.MemberOfVal), 2)
		testutil.Equals(t, a.MemberOfVal[0].Type.Name, types.EntityType("Action"))
		testutil.Equals(t, a.MemberOfVal[0].ID, types.String("readOnly"))
		testutil.Equals(t, a.MemberOfVal[1].ID, types.String("allActions"))
	})

	t.Run("action with principal", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			Principal(ast.EntityType("User"), ast.EntityType("Admin"))
		testutil.Equals(t, a.AppliesToVal != nil, true)
		testutil.Equals(t, len(a.AppliesToVal.PrincipalTypes), 2)
	})

	t.Run("action with resource", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			Resource(ast.EntityType("Document"))
		testutil.Equals(t, a.AppliesToVal != nil, true)
		testutil.Equals(t, len(a.AppliesToVal.ResourceTypes), 1)
	})

	t.Run("action with context", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			Context(ast.Record(
				ast.Attribute("ip", ast.IPAddr()),
			))
		testutil.Equals(t, a.AppliesToVal != nil, true)
		rt, ok := a.AppliesToVal.Context.(ast.RecordType)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, len(rt.Pairs), 1)
	})

	t.Run("action with annotation", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			Annotate("doc", "View action")
		testutil.Equals(t, len(a.Annotations), 1)
	})

	t.Run("action with everything", func(t *testing.T) {
		t.Parallel()
		a := ast.Action("view").
			MemberOf(ast.UID("readOnly")).
			Principal(ast.EntityType("User")).
			Resource(ast.EntityType("Document")).
			Context(ast.Record()).
			Annotate("doc", "View action")
		testutil.Equals(t, len(a.MemberOfVal), 1)
		testutil.Equals(t, a.AppliesToVal != nil, true)
		testutil.Equals(t, len(a.AppliesToVal.PrincipalTypes), 1)
		testutil.Equals(t, len(a.AppliesToVal.ResourceTypes), 1)
		testutil.Equals(t, a.AppliesToVal.Context != nil, true)
		testutil.Equals(t, len(a.Annotations), 1)
	})
}
