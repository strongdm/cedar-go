package ast_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestSchema(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		nodes     []ast.IsNode
		wantNodes int
	}{
		{"empty schema", []ast.IsNode{}, 0},
		{"schema with namespace", []ast.IsNode{ast.Namespace("MyApp")}, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := ast.NewSchema(tt.nodes...)
			testutil.Equals(t, len(s.Nodes), tt.wantNodes)
		})
	}
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

	ns1 := ast.Namespace(types.Path("App1"), ct1, e1, enum1, a1)
	ns2 := ast.Namespace(types.Path("App2"), ct2, e2, enum2, a2)
	ns3 := ast.Namespace(types.Path("App3"), ct3, e3, enum3, a3)

	schema := ast.NewSchema(ns1, ns2, ns3)

	tests := []struct {
		name      string
		count     func() int
		wantCount int
	}{
		{
			"Namespaces",
			func() int {
				count := 0
				for range schema.Namespaces() {
					count++
				}
				return count
			},
			3,
		},
		{
			"CommonTypes",
			func() int {
				count := 0
				for range schema.CommonTypes() {
					count++
				}
				return count
			},
			3,
		},
		{
			"Entities",
			func() int {
				count := 0
				for range schema.Entities() {
					count++
				}
				return count
			},
			3,
		},
		{
			"Enums",
			func() int {
				count := 0
				for range schema.Enums() {
					count++
				}
				return count
			},
			3,
		},
		{
			"Actions",
			func() int {
				count := 0
				for range schema.Actions() {
					count++
				}
				return count
			},
			3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.count()
			testutil.Equals(t, got, tt.wantCount)
		})
	}
}

func TestSchemaIteratorsEmpty(t *testing.T) {
	t.Parallel()

	testIteratorCounts := func(t *testing.T, schema *ast.Schema, wantNS, wantCT, wantE, wantEnum, wantA int) {
		testutil.Equals(t, countIter(schema.Namespaces()), wantNS)
		testutil.Equals(t, countIter2(schema.CommonTypes()), wantCT)
		testutil.Equals(t, countIter2(schema.Entities()), wantE)
		testutil.Equals(t, countIter2(schema.Enums()), wantEnum)
		testutil.Equals(t, countIter2(schema.Actions()), wantA)
	}

	t.Run("empty schema", func(t *testing.T) {
		schema := ast.NewSchema()
		testIteratorCounts(t, schema, 0, 0, 0, 0, 0)
	})

	t.Run("schema with empty namespaces", func(t *testing.T) {
		ns1 := ast.Namespace(types.Path("Empty1"))
		ns2 := ast.Namespace(types.Path("Empty2"))
		schema := ast.NewSchema(ns1, ns2)
		testIteratorCounts(t, schema, 2, 0, 0, 0, 0)
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

	// Test early break for all iterator types
	iteratorTests := []struct {
		name      string
		schema    *ast.Schema
		countFunc func(*ast.Schema) int
	}{
		{
			"Entities",
			ast.NewSchema(
				ast.Namespace(types.Path("A"), ast.Entity("E1")),
				ast.Namespace(types.Path("B"), ast.Entity("E2")),
				ast.Namespace(types.Path("C"), ast.Entity("E3")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Entities() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Enums",
			ast.NewSchema(
				ast.Namespace(types.Path("A"), ast.Enum("E1", "a")),
				ast.Namespace(types.Path("B"), ast.Enum("E2", "b")),
				ast.Namespace(types.Path("C"), ast.Enum("E3", "c")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Enums() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Actions",
			ast.NewSchema(
				ast.Namespace(types.Path("A"), ast.Action("A1")),
				ast.Namespace(types.Path("B"), ast.Action("A2")),
				ast.Namespace(types.Path("C"), ast.Action("A3")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Actions() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
	}

	for _, tt := range iteratorTests {
		t.Run(tt.name+" early break", func(t *testing.T) {
			testutil.Equals(t, tt.countFunc(tt.schema), 1)
		})
	}

	// Test early break on top-level declarations
	topLevelTests := []struct {
		name      string
		schema    *ast.Schema
		countFunc func(*ast.Schema) int
	}{
		{
			"CommonTypes on top-level",
			ast.NewSchema(
				ast.CommonType("CT1", ast.String()),
				ast.CommonType("CT2", ast.String()),
				ast.Namespace(types.Path("NS")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.CommonTypes() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Entities on top-level",
			ast.NewSchema(
				ast.Entity("TopE1"),
				ast.Entity("TopE2"),
				ast.Namespace(types.Path("NS")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Entities() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Enums on top-level",
			ast.NewSchema(
				ast.Enum("TopEnum1", "a"),
				ast.Enum("TopEnum2", "b"),
				ast.Namespace(types.Path("NS")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Enums() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
		{
			"Actions on top-level",
			ast.NewSchema(
				ast.Action("TopA1"),
				ast.Action("TopA2"),
				ast.Namespace(types.Path("NS")),
			),
			func(s *ast.Schema) int {
				count := 0
				for range s.Actions() {
					count++
					if count == 1 {
						break
					}
				}
				return count
			},
		},
	}

	for _, tt := range topLevelTests {
		t.Run(tt.name+" early break", func(t *testing.T) {
			testutil.Equals(t, tt.countFunc(tt.schema), 1)
		})
	}
}

func TestSchemaIteratorsMixed(t *testing.T) {
	t.Parallel()

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

	tests := []struct {
		name      string
		countFunc func(*ast.Schema) int
		want      int
	}{
		{"CommonTypes", func(s *ast.Schema) int { return countIter2(s.CommonTypes()) }, 3},
		{"Entities", func(s *ast.Schema) int { return countIter2(s.Entities()) }, 2},
		{"Enums", func(s *ast.Schema) int { return countIter2(s.Enums()) }, 2},
		{"Actions", func(s *ast.Schema) int { return countIter2(s.Actions()) }, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name+" from mixed declarations", func(t *testing.T) {
			got := tt.countFunc(schema)
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestSchemaIteratorsTopLevel(t *testing.T) {
	t.Parallel()

	ct1 := ast.CommonType("TopType", ast.String())
	e1 := ast.Entity("TopEntity")
	enum1 := ast.Enum("TopEnum", "a", "b")
	a1 := ast.Action("topAction")

	ns := ast.Namespace(types.Path("NS"),
		ast.CommonType("NSType", ast.Long()),
		ast.Entity("NSEntity"),
		ast.Enum("NSEnum", "x", "y"),
		ast.Action("nsAction"),
	)

	schema := ast.NewSchema(ct1, e1, enum1, a1, ns)

	tests := []struct {
		name      string
		countFunc func(*ast.Schema) int
		want      int
	}{
		{"CommonTypes", func(s *ast.Schema) int { return countIter2(s.CommonTypes()) }, 2},
		{"Entities", func(s *ast.Schema) int { return countIter2(s.Entities()) }, 2},
		{"Enums", func(s *ast.Schema) int { return countIter2(s.Enums()) }, 2},
		{"Actions", func(s *ast.Schema) int { return countIter2(s.Actions()) }, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name+" includes top-level", func(t *testing.T) {
			got := tt.countFunc(schema)
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestSchemaIteratorsOnlyTopLevel(t *testing.T) {
	t.Parallel()

	ct := ast.CommonType("Type", ast.String())
	e := ast.Entity("Entity")
	enum := ast.Enum("Enum", "val")
	a := ast.Action("action")

	schema := ast.NewSchema(ct, e, enum, a)

	tests := []struct {
		name      string
		countFunc func(*ast.Schema) int
		want      int
	}{
		{"CommonTypes", func(s *ast.Schema) int { return countIter2(s.CommonTypes()) }, 1},
		{"Entities", func(s *ast.Schema) int { return countIter2(s.Entities()) }, 1},
		{"Enums", func(s *ast.Schema) int { return countIter2(s.Enums()) }, 1},
		{"Actions", func(s *ast.Schema) int { return countIter2(s.Actions()) }, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name+" only top-level", func(t *testing.T) {
			got := tt.countFunc(schema)
			testutil.Equals(t, got, tt.want)
		})
	}
}

func TestNamespace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		ns               ast.NamespaceNode
		wantName         types.Path
		wantDeclarations int
		wantAnnotations  int
	}{
		{
			"basic namespace",
			ast.Namespace("MyApp::Accounts"),
			types.Path("MyApp::Accounts"),
			0,
			0,
		},
		{
			"namespace with declarations",
			ast.Namespace("MyApp", ast.Entity("User"), ast.Entity("Group")),
			types.Path("MyApp"),
			2,
			0,
		},
		{
			"namespace with annotations",
			ast.Namespace("MyApp").Annotate("doc", "My application namespace"),
			types.Path("MyApp"),
			0,
			1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.ns.Name, tt.wantName)
			testutil.Equals(t, len(tt.ns.Declarations), tt.wantDeclarations)
			testutil.Equals(t, len(tt.ns.Annotations), tt.wantAnnotations)
		})
	}
}

func TestCommonType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		ct              ast.CommonTypeNode
		wantName        types.Ident
		wantAnnotations int
		check           func(t *testing.T, ct ast.CommonTypeNode)
	}{
		{
			"basic common type",
			ast.CommonType("Name", ast.String()),
			types.Ident("Name"),
			0,
			func(t *testing.T, ct ast.CommonTypeNode) {
				_, ok := ct.Type.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			"common type with record",
			ast.CommonType("Address", ast.Record(
				ast.Attribute("street", ast.String()),
				ast.Optional("apt", ast.String()),
			)),
			types.Ident("Address"),
			0,
			func(t *testing.T, ct ast.CommonTypeNode) {
				rt, ok := ct.Type.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(rt.Pairs), 2)
			},
		},
		{
			"common type with annotation",
			ast.CommonType("Name", ast.String()).Annotate("doc", "A person's name"),
			types.Ident("Name"),
			1,
			func(t *testing.T, ct ast.CommonTypeNode) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, tt.ct.Name, tt.wantName)
			testutil.Equals(t, len(tt.ct.Annotations), tt.wantAnnotations)
			tt.check(t, tt.ct)
		})
	}
}

func TestEntity(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		e     ast.EntityNode
		check func(t *testing.T, e ast.EntityNode)
	}{
		{
			"basic entity",
			ast.Entity("User"),
			func(t *testing.T, e ast.EntityNode) {
				testutil.Equals(t, e.Name, types.EntityType("User"))
				testutil.Equals(t, e.MemberOfVal, nil)
				testutil.Equals(t, e.ShapeVal, nil)
				testutil.Equals(t, e.TagsVal, nil)
			},
		},
		{
			"entity with memberOf",
			ast.Entity("User").MemberOf(ast.EntityType("Group"), ast.EntityType("Team")),
			func(t *testing.T, e ast.EntityNode) {
				testutil.Equals(t, len(e.MemberOfVal), 2)
				testutil.Equals(t, e.MemberOfVal[0].Name, types.EntityType("Group"))
				testutil.Equals(t, e.MemberOfVal[1].Name, types.EntityType("Team"))
			},
		},
		{
			"entity with shape",
			ast.Entity("User").Shape(
				ast.Attribute("name", ast.String()),
				ast.Optional("email", ast.String()),
			),
			func(t *testing.T, e ast.EntityNode) {
				testutil.Equals(t, e.ShapeVal != nil, true)
				testutil.Equals(t, len(e.ShapeVal.Pairs), 2)
				testutil.Equals(t, e.ShapeVal.Pairs[0].Optional, false)
				testutil.Equals(t, e.ShapeVal.Pairs[1].Optional, true)
			},
		},
		{
			"entity with tags",
			ast.Entity("Document").Tags(ast.String()),
			func(t *testing.T, e ast.EntityNode) {
				_, ok := e.TagsVal.(ast.StringType)
				testutil.Equals(t, ok, true)
			},
		},
		{
			"entity with annotation",
			ast.Entity("User").Annotate("doc", "A user entity"),
			func(t *testing.T, e ast.EntityNode) {
				testutil.Equals(t, len(e.Annotations), 1)
			},
		},
		{
			"entity with everything",
			ast.Entity("User").
				MemberOf(ast.EntityType("Group")).
				Shape(ast.Attribute("name", ast.String())).
				Tags(ast.String()).
				Annotate("doc", "A user"),
			func(t *testing.T, e ast.EntityNode) {
				testutil.Equals(t, len(e.MemberOfVal), 1)
				testutil.Equals(t, e.ShapeVal != nil, true)
				testutil.Equals(t, e.TagsVal != nil, true)
				testutil.Equals(t, len(e.Annotations), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t, tt.e)
		})
	}
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
		for uid := range e.EntityUIDs() {
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

	tests := []struct {
		name  string
		a     ast.ActionNode
		check func(t *testing.T, a ast.ActionNode)
	}{
		{
			"basic action",
			ast.Action("view"),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, a.Name, types.String("view"))
				testutil.Equals(t, a.MemberOfVal, nil)
				testutil.Equals(t, a.AppliesToVal, nil)
			},
		},
		{
			"action with memberOf",
			ast.Action("view").MemberOf(ast.UID("readOnly"), ast.EntityUID("Action", "allActions")),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, len(a.MemberOfVal), 2)
				testutil.Equals(t, a.MemberOfVal[0].Type.Name, types.EntityType("Action"))
				testutil.Equals(t, a.MemberOfVal[0].ID, types.String("readOnly"))
			},
		},
		{
			"action with principal",
			ast.Action("view").Principal(ast.EntityType("User"), ast.EntityType("Admin")),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, a.AppliesToVal != nil, true)
				testutil.Equals(t, len(a.AppliesToVal.PrincipalTypes), 2)
			},
		},
		{
			"action with resource",
			ast.Action("view").Resource(ast.EntityType("Document")),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, a.AppliesToVal != nil, true)
				testutil.Equals(t, len(a.AppliesToVal.ResourceTypes), 1)
			},
		},
		{
			"action with context",
			ast.Action("view").Context(ast.Record(ast.Attribute("ip", ast.IPAddr()))),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, a.AppliesToVal != nil, true)
				rt, ok := a.AppliesToVal.Context.(ast.RecordType)
				testutil.Equals(t, ok, true)
				testutil.Equals(t, len(rt.Pairs), 1)
			},
		},
		{
			"action with annotation",
			ast.Action("view").Annotate("doc", "View action"),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, len(a.Annotations), 1)
			},
		},
		{
			"action with everything",
			ast.Action("view").
				MemberOf(ast.UID("readOnly")).
				Principal(ast.EntityType("User")).
				Resource(ast.EntityType("Document")).
				Context(ast.Record()).
				Annotate("doc", "View action"),
			func(t *testing.T, a ast.ActionNode) {
				testutil.Equals(t, len(a.MemberOfVal), 1)
				testutil.Equals(t, a.AppliesToVal != nil, true)
				testutil.Equals(t, len(a.AppliesToVal.PrincipalTypes), 1)
				testutil.Equals(t, len(a.AppliesToVal.ResourceTypes), 1)
				testutil.Equals(t, a.AppliesToVal.Context != nil, true)
				testutil.Equals(t, len(a.Annotations), 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t, tt.a)
		})
	}
}

// Helper functions for counting iterators
func countIter[T any](iter func(func(T) bool)) int {
	count := 0
	iter(func(T) bool {
		count++
		return true
	})
	return count
}

func countIter2[T, U any](iter func(func(T, U) bool)) int {
	count := 0
	iter(func(T, U) bool {
		count++
		return true
	})
	return count
}
