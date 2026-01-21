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

func TestSchemaIterators(t *testing.T) {
	t.Parallel()

	// Create test declarations for multiple namespaces
	ct1 := ast.CommonType("Type1", ast.String())
	ct2 := ast.CommonType("Type2", ast.Long())
	ct3 := ast.CommonType("Type3", ast.Bool())

	e1 := ast.Entity("User")
	e2 := ast.Entity("Group")
	e3 := ast.Entity("Role")

	enum1 := ast.Enum("Status", "active", "inactive")
	enum2 := ast.Enum("Priority", "high", "low")
	enum3 := ast.Enum("Color", "red", "blue")

	a1 := ast.Action("read")
	a2 := ast.Action("write")
	a3 := ast.Action("delete")

	// Create namespaces with different declarations
	ns1 := ast.Namespace(types.Path("App1"), ct1, e1, enum1, a1)
	ns2 := ast.Namespace(types.Path("App2"), ct2, e2, enum2, a2)
	ns3 := ast.Namespace(types.Path("App3"), ct3, e3, enum3, a3)

	schema := ast.NewSchema(ns1, ns2, ns3)

	t.Run("Namespaces", func(t *testing.T) {
		var namespaces []ast.NamespaceNode
		for ns := range schema.Namespaces() {
			namespaces = append(namespaces, ns)
		}
		testutil.Equals(t, len(namespaces), 3)
		testutil.Equals(t, namespaces[0], ns1)
		testutil.Equals(t, namespaces[1], ns2)
		testutil.Equals(t, namespaces[2], ns3)
	})

	t.Run("CommonTypes", func(t *testing.T) {
		var commonTypes []ast.CommonTypeNode
		var namespaces []*ast.NamespaceNode
		for ns, ct := range schema.CommonTypes() {
			namespaces = append(namespaces, ns)
			commonTypes = append(commonTypes, ct)
		}
		testutil.Equals(t, len(commonTypes), 3)
		testutil.Equals(t, commonTypes[0], ct1)
		testutil.Equals(t, commonTypes[1], ct2)
		testutil.Equals(t, commonTypes[2], ct3)
		testutil.Equals(t, namespaces[0].Name, types.Path("App1"))
		testutil.Equals(t, namespaces[1].Name, types.Path("App2"))
		testutil.Equals(t, namespaces[2].Name, types.Path("App3"))
	})

	t.Run("Entities", func(t *testing.T) {
		var entities []ast.EntityNode
		var namespaces []*ast.NamespaceNode
		for ns, e := range schema.Entities() {
			namespaces = append(namespaces, ns)
			entities = append(entities, e)
		}
		testutil.Equals(t, len(entities), 3)
		testutil.Equals(t, entities[0], e1)
		testutil.Equals(t, entities[1], e2)
		testutil.Equals(t, entities[2], e3)
		testutil.Equals(t, namespaces[0].Name, types.Path("App1"))
		testutil.Equals(t, namespaces[1].Name, types.Path("App2"))
		testutil.Equals(t, namespaces[2].Name, types.Path("App3"))
	})

	t.Run("Enums", func(t *testing.T) {
		var enums []ast.EnumNode
		var namespaces []*ast.NamespaceNode
		for ns, e := range schema.Enums() {
			namespaces = append(namespaces, ns)
			enums = append(enums, e)
		}
		testutil.Equals(t, len(enums), 3)
		testutil.Equals(t, enums[0], enum1)
		testutil.Equals(t, enums[1], enum2)
		testutil.Equals(t, enums[2], enum3)
		testutil.Equals(t, namespaces[0].Name, types.Path("App1"))
		testutil.Equals(t, namespaces[1].Name, types.Path("App2"))
		testutil.Equals(t, namespaces[2].Name, types.Path("App3"))
	})

	t.Run("Actions", func(t *testing.T) {
		var actions []ast.ActionNode
		var namespaces []*ast.NamespaceNode
		for ns, a := range schema.Actions() {
			namespaces = append(namespaces, ns)
			actions = append(actions, a)
		}
		testutil.Equals(t, len(actions), 3)
		testutil.Equals(t, actions[0], a1)
		testutil.Equals(t, actions[1], a2)
		testutil.Equals(t, actions[2], a3)
		testutil.Equals(t, namespaces[0].Name, types.Path("App1"))
		testutil.Equals(t, namespaces[1].Name, types.Path("App2"))
		testutil.Equals(t, namespaces[2].Name, types.Path("App3"))
	})
}

func TestSchemaIteratorsEmpty(t *testing.T) {
	t.Parallel()

	t.Run("empty schema", func(t *testing.T) {
		schema := ast.NewSchema()

		nsCount := 0
		for range schema.Namespaces() {
			nsCount++
		}
		testutil.Equals(t, nsCount, 0)

		ctCount := 0
		for range schema.CommonTypes() {
			ctCount++
		}
		testutil.Equals(t, ctCount, 0)

		eCount := 0
		for range schema.Entities() {
			eCount++
		}
		testutil.Equals(t, eCount, 0)

		enumCount := 0
		for range schema.Enums() {
			enumCount++
		}
		testutil.Equals(t, enumCount, 0)

		aCount := 0
		for range schema.Actions() {
			aCount++
		}
		testutil.Equals(t, aCount, 0)
	})

	t.Run("schema with empty namespaces", func(t *testing.T) {
		ns1 := ast.Namespace(types.Path("Empty1"))
		ns2 := ast.Namespace(types.Path("Empty2"))
		schema := ast.NewSchema(ns1, ns2)

		nsCount := 0
		for range schema.Namespaces() {
			nsCount++
		}
		testutil.Equals(t, nsCount, 2)

		ctCount := 0
		for range schema.CommonTypes() {
			ctCount++
		}
		testutil.Equals(t, ctCount, 0)

		eCount := 0
		for range schema.Entities() {
			eCount++
		}
		testutil.Equals(t, eCount, 0)

		enumCount := 0
		for range schema.Enums() {
			enumCount++
		}
		testutil.Equals(t, enumCount, 0)

		aCount := 0
		for range schema.Actions() {
			aCount++
		}
		testutil.Equals(t, aCount, 0)
	})
}

func TestSchemaIteratorsEarlyBreak(t *testing.T) {
	t.Parallel()

	ns1 := ast.Namespace(types.Path("NS1"), ast.CommonType("T1", ast.String()))
	ns2 := ast.Namespace(types.Path("NS2"), ast.CommonType("T2", ast.String()))
	ns3 := ast.Namespace(types.Path("NS3"), ast.CommonType("T3", ast.String()))

	schema := ast.NewSchema(ns1, ns2, ns3)

	t.Run("Namespaces early break", func(t *testing.T) {
		count := 0
		for ns := range schema.Namespaces() {
			count++
			if ns.Name == ns1.Name {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("CommonTypes early break", func(t *testing.T) {
		count := 0
		for range schema.CommonTypes() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	e1 := ast.Entity("E1")
	e2 := ast.Entity("E2")
	e3 := ast.Entity("E3")
	schemaE := ast.NewSchema(
		ast.Namespace(types.Path("A"), e1),
		ast.Namespace(types.Path("B"), e2),
		ast.Namespace(types.Path("C"), e3),
	)

	t.Run("Entities early break", func(t *testing.T) {
		count := 0
		for range schemaE.Entities() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	enum1 := ast.Enum("E1", "a")
	enum2 := ast.Enum("E2", "b")
	enum3 := ast.Enum("E3", "c")
	schemaEnum := ast.NewSchema(
		ast.Namespace(types.Path("A"), enum1),
		ast.Namespace(types.Path("B"), enum2),
		ast.Namespace(types.Path("C"), enum3),
	)

	t.Run("Enums early break", func(t *testing.T) {
		count := 0
		for range schemaEnum.Enums() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	a1 := ast.Action("A1")
	a2 := ast.Action("A2")
	a3 := ast.Action("A3")
	schemaA := ast.NewSchema(
		ast.Namespace(types.Path("A"), a1),
		ast.Namespace(types.Path("B"), a2),
		ast.Namespace(types.Path("C"), a3),
	)

	t.Run("Actions early break", func(t *testing.T) {
		count := 0
		for range schemaA.Actions() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	// Test early break on top-level declarations
	ct1 := ast.CommonType("CT1", ast.String())
	ct2 := ast.CommonType("CT2", ast.String())
	topE1 := ast.Entity("TopE1")
	topE2 := ast.Entity("TopE2")
	topEnum1 := ast.Enum("TopEnum1", "a")
	topEnum2 := ast.Enum("TopEnum2", "b")
	topA1 := ast.Action("TopA1")
	topA2 := ast.Action("TopA2")

	t.Run("CommonTypes early break on top-level", func(t *testing.T) {
		schema := ast.NewSchema(ct1, ct2, ast.Namespace(types.Path("NS")))
		count := 0
		for range schema.CommonTypes() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Entities early break on top-level", func(t *testing.T) {
		schema := ast.NewSchema(topE1, topE2, ast.Namespace(types.Path("NS")))
		count := 0
		for range schema.Entities() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Enums early break on top-level", func(t *testing.T) {
		schema := ast.NewSchema(topEnum1, topEnum2, ast.Namespace(types.Path("NS")))
		count := 0
		for range schema.Enums() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Actions early break on top-level", func(t *testing.T) {
		schema := ast.NewSchema(topA1, topA2, ast.Namespace(types.Path("NS")))
		count := 0
		for range schema.Actions() {
			count++
			if count == 1 {
				break
			}
		}
		testutil.Equals(t, count, 1)
	})
}

func TestSchemaIteratorsMixed(t *testing.T) {
	t.Parallel()

	// Create a complex schema with multiple namespaces containing different declaration types
	ns1 := ast.Namespace(types.Path("App"),
		ast.CommonType("Name", ast.String()),
		ast.Entity("User"),
		ast.Enum("Status", "active"),
		ast.Action("read"),
		ast.CommonType("Age", ast.Long()),
		ast.Entity("Group"),
	)

	ns2 := ast.Namespace(types.Path("System"),
		ast.Action("write"),
		ast.Enum("Priority", "high"),
		ast.CommonType("ID", ast.String()),
	)

	schema := ast.NewSchema(ns1, ns2)

	t.Run("CommonTypes from mixed declarations", func(t *testing.T) {
		count := 0
		for range schema.CommonTypes() {
			count++
		}
		testutil.Equals(t, count, 3) // Name, Age, ID
	})

	t.Run("Entities from mixed declarations", func(t *testing.T) {
		count := 0
		for range schema.Entities() {
			count++
		}
		testutil.Equals(t, count, 2) // User, Group
	})

	t.Run("Enums from mixed declarations", func(t *testing.T) {
		count := 0
		for range schema.Enums() {
			count++
		}
		testutil.Equals(t, count, 2) // Status, Priority
	})

	t.Run("Actions from mixed declarations", func(t *testing.T) {
		count := 0
		for range schema.Actions() {
			count++
		}
		testutil.Equals(t, count, 2) // read, write
	})
}

func TestSchemaIteratorsTopLevel(t *testing.T) {
	t.Parallel()

	// Create top-level declarations (not in a namespace)
	ct1 := ast.CommonType("TopType", ast.String())
	e1 := ast.Entity("TopEntity")
	enum1 := ast.Enum("TopEnum", "a", "b")
	a1 := ast.Action("topAction")

	// Create a namespace with its own declarations
	ns := ast.Namespace(types.Path("NS"),
		ast.CommonType("NSType", ast.Long()),
		ast.Entity("NSEntity"),
		ast.Enum("NSEnum", "x", "y"),
		ast.Action("nsAction"),
	)

	// Schema with both top-level declarations and namespaced declarations
	schema := ast.NewSchema(ct1, e1, enum1, a1, ns)

	t.Run("CommonTypes includes top-level", func(t *testing.T) {
		var commonTypes []ast.CommonTypeNode
		var namespaces []*ast.NamespaceNode
		for ns, ct := range schema.CommonTypes() {
			namespaces = append(namespaces, ns)
			commonTypes = append(commonTypes, ct)
		}
		testutil.Equals(t, len(commonTypes), 2) // TopType, NSType
		testutil.Equals(t, commonTypes[0], ct1)
		testutil.Equals(t, namespaces[0] == nil, true) // top-level
		testutil.Equals(t, namespaces[1].Name, types.Path("NS"))
	})

	t.Run("Entities includes top-level", func(t *testing.T) {
		var entities []ast.EntityNode
		var namespaces []*ast.NamespaceNode
		for ns, e := range schema.Entities() {
			namespaces = append(namespaces, ns)
			entities = append(entities, e)
		}
		testutil.Equals(t, len(entities), 2) // TopEntity, NSEntity
		testutil.Equals(t, entities[0], e1)
		testutil.Equals(t, namespaces[0] == nil, true) // top-level
		testutil.Equals(t, namespaces[1].Name, types.Path("NS"))
	})

	t.Run("Enums includes top-level", func(t *testing.T) {
		var enums []ast.EnumNode
		var namespaces []*ast.NamespaceNode
		for ns, e := range schema.Enums() {
			namespaces = append(namespaces, ns)
			enums = append(enums, e)
		}
		testutil.Equals(t, len(enums), 2) // TopEnum, NSEnum
		testutil.Equals(t, enums[0], enum1)
		testutil.Equals(t, namespaces[0] == nil, true) // top-level
		testutil.Equals(t, namespaces[1].Name, types.Path("NS"))
	})

	t.Run("Actions includes top-level", func(t *testing.T) {
		var actions []ast.ActionNode
		var namespaces []*ast.NamespaceNode
		for ns, a := range schema.Actions() {
			namespaces = append(namespaces, ns)
			actions = append(actions, a)
		}
		testutil.Equals(t, len(actions), 2) // topAction, nsAction
		testutil.Equals(t, actions[0], a1)
		testutil.Equals(t, namespaces[0] == nil, true) // top-level
		testutil.Equals(t, namespaces[1].Name, types.Path("NS"))
	})
}

func TestSchemaIteratorsOnlyTopLevel(t *testing.T) {
	t.Parallel()

	// Schema with only top-level declarations, no namespaces
	ct := ast.CommonType("Type", ast.String())
	e := ast.Entity("Entity")
	enum := ast.Enum("Enum", "val")
	a := ast.Action("action")

	schema := ast.NewSchema(ct, e, enum, a)

	t.Run("CommonTypes only top-level", func(t *testing.T) {
		count := 0
		for range schema.CommonTypes() {
			count++
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Entities only top-level", func(t *testing.T) {
		count := 0
		for range schema.Entities() {
			count++
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Enums only top-level", func(t *testing.T) {
		count := 0
		for range schema.Enums() {
			count++
		}
		testutil.Equals(t, count, 1)
	})

	t.Run("Actions only top-level", func(t *testing.T) {
		count := 0
		for range schema.Actions() {
			count++
		}
		testutil.Equals(t, count, 1)
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
		var uids []types.EntityUID
		for uid := range e.EntityUIDs(nil) {
			uids = append(uids, uid)
		}
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
