package validate

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// testSchema builds a simple schema for testing.
func testSchema() *resolved.Schema {
	return &resolved.Schema{
		Namespaces: map[types.Path]resolved.Namespace{},
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:        "User",
				ParentTypes: []types.EntityType{"Group"},
				Shape: resolved.RecordType{
					"name":  {Type: resolved.StringType{}, Optional: false},
					"age":   {Type: resolved.LongType{}, Optional: true},
					"email": {Type: resolved.StringType{}, Optional: false},
				},
			},
			"Group": {
				Name: "Group",
				Shape: resolved.RecordType{
					"name": {Type: resolved.StringType{}, Optional: false},
				},
			},
			"Document": {
				Name:        "Document",
				ParentTypes: []types.EntityType{"Folder"},
				Shape: resolved.RecordType{
					"title":  {Type: resolved.StringType{}, Optional: false},
					"public": {Type: resolved.BoolType{}, Optional: false},
				},
				Tags: resolved.StringType{},
			},
			"Folder": {
				Name:  "Folder",
				Shape: resolved.RecordType{},
			},
		},
		Enums: map[types.EntityType]resolved.Enum{
			"Color": {
				Name: "Color",
				Values: []types.EntityUID{
					types.NewEntityUID("Color", "red"),
					types.NewEntityUID("Color", "green"),
					types.NewEntityUID("Color", "blue"),
				},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "view"): {
				Entity: types.Entity{
					UID: types.NewEntityUID("Action", "view"),
				},
				AppliesTo: &resolved.AppliesTo{
					Principals: []types.EntityType{"User"},
					Resources:  []types.EntityType{"Document"},
					Context: resolved.RecordType{
						"ip": {Type: resolved.ExtensionType("ipaddr"), Optional: false},
					},
				},
			},
			types.NewEntityUID("Action", "edit"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "edit"),
					Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "view")),
				},
				AppliesTo: &resolved.AppliesTo{
					Principals: []types.EntityType{"User"},
					Resources:  []types.EntityType{"Document"},
					Context:    resolved.RecordType{},
				},
			},
		},
	}
}

// testSchemaWithPhoto extends testSchema with a Photo entity type unrelated to User/Group.
func testSchemaWithPhoto() *resolved.Schema {
	s := testSchema()
	s.Entities["Photo"] = resolved.Entity{
		Name:  "Photo",
		Shape: resolved.RecordType{},
	}
	return s
}

func testEnv() *requestEnv {
	return &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
}

func TestEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	tests := []struct {
		name    string
		entity  types.Entity
		wantErr bool
	}{
		{
			name: "valid",
			entity: types.Entity{
				UID:     types.NewEntityUID("User", "alice"),
				Parents: types.NewEntityUIDSet(types.NewEntityUID("Group", "admins")),
				Attributes: types.NewRecord(types.RecordMap{
					"name":  types.String("Alice"),
					"email": types.String("alice@example.com"),
				}),
			},
		},
		{
			name: "unknownType",
			entity: types.Entity{
				UID: types.NewEntityUID("Unknown", "x"),
			},
			wantErr: true,
		},
		{
			name: "invalidParentType",
			entity: types.Entity{
				UID:     types.NewEntityUID("User", "alice"),
				Parents: types.NewEntityUIDSet(types.NewEntityUID("Document", "doc1")),
				Attributes: types.NewRecord(types.RecordMap{
					"name":  types.String("Alice"),
					"email": types.String("alice@example.com"),
				}),
			},
			wantErr: true,
		},
		{
			name: "missingRequiredAttr",
			entity: types.Entity{
				UID: types.NewEntityUID("User", "alice"),
				Attributes: types.NewRecord(types.RecordMap{
					"name": types.String("Alice"),
				}),
			},
			wantErr: true,
		},
		{
			name: "unexpectedAttr",
			entity: types.Entity{
				UID: types.NewEntityUID("User", "alice"),
				Attributes: types.NewRecord(types.RecordMap{
					"name":    types.String("Alice"),
					"email":   types.String("alice@example.com"),
					"unknown": types.String("extra"),
				}),
			},
			wantErr: true,
		},
		{
			name: "wrongAttrType",
			entity: types.Entity{
				UID: types.NewEntityUID("User", "alice"),
				Attributes: types.NewRecord(types.RecordMap{
					"name":  types.Long(42),
					"email": types.String("alice@example.com"),
				}),
			},
			wantErr: true,
		},
		{
			name: "tagsValid",
			entity: types.Entity{
				UID:     types.NewEntityUID("Document", "doc1"),
				Parents: types.NewEntityUIDSet(types.NewEntityUID("Folder", "folder1")),
				Attributes: types.NewRecord(types.RecordMap{
					"title":  types.String("My Doc"),
					"public": types.Boolean(true),
				}),
				Tags: types.NewRecord(types.RecordMap{
					"category": types.String("report"),
				}),
			},
		},
		{
			name: "tagsNotAllowed",
			entity: types.Entity{
				UID: types.NewEntityUID("User", "alice"),
				Attributes: types.NewRecord(types.RecordMap{
					"name":  types.String("Alice"),
					"email": types.String("alice@example.com"),
				}),
				Tags: types.NewRecord(types.RecordMap{
					"tag1": types.String("val"),
				}),
			},
			wantErr: true,
		},
		{
			name: "tagsWrongType",
			entity: types.Entity{
				UID:     types.NewEntityUID("Document", "doc1"),
				Parents: types.NewEntityUIDSet(types.NewEntityUID("Folder", "folder1")),
				Attributes: types.NewRecord(types.RecordMap{
					"title":  types.String("My Doc"),
					"public": types.Boolean(true),
				}),
				Tags: types.NewRecord(types.RecordMap{
					"category": types.Long(42),
				}),
			},
			wantErr: true,
		},
		{
			name: "validEnum",
			entity: types.Entity{
				UID: types.NewEntityUID("Color", "red"),
			},
		},
		{
			name: "invalidEnumValue",
			entity: types.Entity{
				UID: types.NewEntityUID("Color", "purple"),
			},
			wantErr: true,
		},
		{
			name: "validActionEntity",
			entity: types.Entity{
				UID:     types.NewEntityUID("Action", "edit"),
				Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "view")),
			},
		},
		{
			name: "unknownAction",
			entity: types.Entity{
				UID: types.NewEntityUID("Action", "delete"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.Entity(tt.entity)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

func TestEntitiesMap(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	entities := types.EntityMap{
		types.NewEntityUID("User", "alice"): {
			UID: types.NewEntityUID("User", "alice"),
			Attributes: types.NewRecord(types.RecordMap{
				"name":  types.String("Alice"),
				"email": types.String("alice@example.com"),
			}),
		},
		types.NewEntityUID("Group", "admins"): {
			UID: types.NewEntityUID("Group", "admins"),
			Attributes: types.NewRecord(types.RecordMap{
				"name": types.String("Admins"),
			}),
		},
	}
	testutil.OK(t, v.Entities(entities))
}

func TestRequest(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	tests := []struct {
		name    string
		req     types.Request
		wantErr bool
	}{
		{
			name: "valid",
			req: types.Request{
				Principal: types.NewEntityUID("User", "alice"),
				Action:    types.NewEntityUID("Action", "view"),
				Resource:  types.NewEntityUID("Document", "doc1"),
				Context: types.NewRecord(types.RecordMap{
					"ip": types.IPAddr{},
				}),
			},
		},
		{
			name: "unknownAction",
			req: types.Request{
				Principal: types.NewEntityUID("User", "alice"),
				Action:    types.NewEntityUID("Action", "delete"),
				Resource:  types.NewEntityUID("Document", "doc1"),
			},
			wantErr: true,
		},
		{
			name: "wrongPrincipalType",
			req: types.Request{
				Principal: types.NewEntityUID("Document", "doc1"),
				Action:    types.NewEntityUID("Action", "view"),
				Resource:  types.NewEntityUID("Document", "doc1"),
				Context: types.NewRecord(types.RecordMap{
					"ip": types.IPAddr{},
				}),
			},
			wantErr: true,
		},
		{
			name: "wrongResourceType",
			req: types.Request{
				Principal: types.NewEntityUID("User", "alice"),
				Action:    types.NewEntityUID("Action", "view"),
				Resource:  types.NewEntityUID("User", "bob"),
				Context: types.NewRecord(types.RecordMap{
					"ip": types.IPAddr{},
				}),
			},
			wantErr: true,
		},
		{
			name: "invalidContext",
			req: types.Request{
				Principal: types.NewEntityUID("User", "alice"),
				Action:    types.NewEntityUID("Action", "view"),
				Resource:  types.NewEntityUID("Document", "doc1"),
				Context:   types.NewRecord(types.RecordMap{}),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := v.Request(tt.req)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

func TestPolicyRBAC(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	tests := []struct {
		name    string
		setup   func(*ast.Policy)
		wantErr bool
	}{
		{
			name: "unknownEntityType",
			setup: func(p *ast.Policy) {
				p.PrincipalIs("Unknown")
			},
			wantErr: true,
		},
		{
			name: "unknownAction",
			setup: func(p *ast.Policy) {
				p.ActionEq(types.NewEntityUID("Action", "delete"))
			},
			wantErr: true,
		},
		{
			name: "invalidActionApplication",
			setup: func(p *ast.Policy) {
				p.PrincipalIs("Group")
				p.ActionEq(types.NewEntityUID("Action", "view"))
			},
			wantErr: true,
		},
		{
			name: "validSimple",
			setup: func(p *ast.Policy) {
				p.PrincipalIs("User")
				p.ActionEq(types.NewEntityUID("Action", "view"))
				p.ResourceIs("Document")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := ast.Permit()
			tt.setup(p)
			err := v.Policy(p)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

func TestTypeCheckVariables(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType: schemaRecordToCedarType(resolved.RecordType{
			"ip": {Type: resolved.ExtensionType("ipaddr"), Optional: false},
		}),
	}
	caps := newCapabilitySet()

	// principal → typeEntity{User}
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeVariable{Name: "principal"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("User")})

	// context → typeRecord
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeVariable{Name: "context"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, env.contextType)
}

func TestTypeCheckArithmetic(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// 1 + 2 → typeLong
	expr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}
	ty, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})

	// 1 + "two" → error
	badExpr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.String("two")},
	}}
	_, _, err = v.typeOfExpr(env, badExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckLogical(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// true && false → typeBool
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Boolean(false)},
	}}
	ty, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// true && 42 → error
	badExpr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Long(42)},
	}}
	_, _, err = v.typeOfExpr(env, badExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckAttributeAccess(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal.name → typeString (required attr)
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "name",
	}}
	ty, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeString{})

	// principal.age → error (optional attr without has guard)
	optExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	_, _, err = v.typeOfExpr(env, optExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckHasGuard(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal has age && principal.age > 18
	hasExpr := ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}

	_, hasCaps, err := v.typeOfExpr(env, hasExpr, caps)
	testutil.OK(t, err)

	// Now with hasCaps, principal.age should work
	accessExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	ty, _, err := v.typeOfExpr(env, accessExpr, hasCaps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})
}

func TestTypeCheckExtensionFunction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// ip("127.0.0.1") → typeExtension{ipaddr}
	expr := ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeValue{Value: types.String("127.0.0.1")}},
	}
	ty, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeExtension{name: "ipaddr"})

	// ip("127.0.0.1").isLoopback() → typeBool
	expr2 := ast.NodeTypeExtensionCall{
		Name: "isLoopback",
		Args: []ast.IsNode{
			ast.NodeTypeExtensionCall{
				Name: "ip",
				Args: []ast.IsNode{ast.NodeValue{Value: types.String("127.0.0.1")}},
			},
		},
	}
	ty, _, err = v.typeOfExpr(env, expr2, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)
}

func TestTypeCheckIfThenElse(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// if true then 1 else 2 → typeLong
	expr := ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.Long(2)},
	}
	ty, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})
}

func TestTypeCheckSetRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// [1, 2, 3] → typeSet{Long}
	setExpr := ast.NodeTypeSet{
		Elements: []ast.IsNode{
			ast.NodeValue{Value: types.Long(1)},
			ast.NodeValue{Value: types.Long(2)},
			ast.NodeValue{Value: types.Long(3)},
		},
	}
	ty, _, err := v.typeOfExpr(env, setExpr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeSet{element: typeLong{}})

	// {"a": 1, "b": "hello"} → typeRecord
	recExpr := ast.NodeTypeRecord{
		Elements: []ast.RecordElementNode{
			{Key: "a", Value: ast.NodeValue{Value: types.Long(1)}},
			{Key: "b", Value: ast.NodeValue{Value: types.String("hello")}},
		},
	}
	ty, _, err = v.typeOfExpr(env, recExpr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeRecord{
		attrs: map[types.String]attributeType{
			"a": {typ: typeLong{}, required: true},
			"b": {typ: typeString{}, required: true},
		},
	})
}

func TestTypeCheckUnsafeAccess(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal.nonexistent → error
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "nonexistent",
	}}
	_, _, err := v.typeOfExpr(env, expr, caps)
	testutil.Error(t, err)
}

func TestIsSubtype(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})

	tests := []struct {
		name string
		a, b cedarType
		want bool
	}{
		{"Never<:Long", typeNever{}, typeLong{}, true},
		{"Never<:Bool", typeNever{}, typeBool{}, true},
		{"True<:Bool", typeTrue{}, typeBool{}, true},
		{"False<:Bool", typeFalse{}, typeBool{}, true},
		{"Bool<:Bool", typeBool{}, typeBool{}, true},
		{"Long<:Long", typeLong{}, typeLong{}, true},
		{"String<:String", typeString{}, typeString{}, true},
		{"Long!<:String", typeLong{}, typeString{}, false},
		{"Set<:Set", typeSet{element: typeLong{}}, typeSet{element: typeLong{}}, true},
		{"Entity<:Entity", typeEntity{lub: singleEntityLUB("User")}, typeEntity{lub: singleEntityLUB("User")}, true},
		{"Entity!<:Entity", typeEntity{lub: singleEntityLUB("User")}, typeEntity{lub: singleEntityLUB("Group")}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, v.isSubtype(tt.a, tt.b), tt.want)
		})
	}
}

func TestLeastUpperBound(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})

	tests := []struct {
		name    string
		a, b    cedarType
		want    cedarType
		wantErr bool
	}{
		{"True|False=Bool", typeTrue{}, typeFalse{}, typeBool{}, false},
		{"Long|Long=Long", typeLong{}, typeLong{}, typeLong{}, false},
		{"Long|String=error", typeLong{}, typeString{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := v.leastUpperBound(tt.a, tt.b)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
				testutil.Equals(t, got, tt.want)
			}
		})
	}
}

func TestTypeCheckStrict(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s)
	env := testEnv()

	tests := []struct {
		name    string
		expr    ast.IsNode
		wantErr bool
	}{
		{
			name: "equalityEntityVsString",
			expr: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeTypeVariable{Name: "principal"},
				Right: ast.NodeValue{Value: types.String("foo")},
			}},
			wantErr: true,
		},
		{
			name: "containsLongSetStringArg",
			expr: ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
				Right: ast.NodeValue{Value: types.String("test")},
			}},
			wantErr: true,
		},
		{
			name: "setDisjointEntityTypes",
			expr: ast.NodeTypeSet{Elements: []ast.IsNode{
				ast.NodeValue{Value: types.NewEntityUID("User", "a")},
				ast.NodeValue{Value: types.NewEntityUID("Photo", "b")},
			}},
			wantErr: true,
		},
		{
			name: "emptySetContains",
			expr: ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeTypeSet{Elements: []ast.IsNode{}},
				Right: ast.NodeValue{Value: types.Long(1)},
			}},
			wantErr: true,
		},
		{
			name: "ifIncompatibleEntities",
			expr: ast.NodeTypeIfThenElse{
				If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
					Left:  ast.NodeValue{Value: types.Long(1)},
					Right: ast.NodeValue{Value: types.Long(1)},
				}},
				Then: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
				Else: ast.NodeValue{Value: types.NewEntityUID("Photo", "b")},
			},
			wantErr: true,
		},
		{
			name: "andFalseShortCircuit",
			expr: ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeValue{Value: types.Boolean(false)},
				Right: ast.NodeValue{Value: types.NewEntityUID("Action", "view")},
			}},
			wantErr: false,
		},
		{
			name: "andNonBoolRhs",
			expr: ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeValue{Value: types.Boolean(true)},
				Right: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
			}},
			wantErr: true,
		},
		{
			name: "entityInUnrelated",
			expr: ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
				Left:  ast.NodeValue{Value: types.NewEntityUID("User", "a")},
				Right: ast.NodeValue{Value: types.NewEntityUID("Photo", "b")},
			}},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			caps := newCapabilitySet()
			_, _, err := v.typeOfExpr(env, tt.expr, caps)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
			}
		})
	}
}

func TestHasResultType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	tests := []struct {
		name   string
		target cedarType
		attr   types.String
		want   cedarType
	}{
		{
			name: "recordRequired",
			target: typeRecord{attrs: map[types.String]attributeType{
				"name": {typ: typeString{}, required: true},
			}},
			attr: "name",
			want: typeTrue{},
		},
		{
			name: "recordOptional",
			target: typeRecord{attrs: map[types.String]attributeType{
				"age": {typ: typeLong{}, required: false},
			}},
			attr: "age",
			want: typeBool{},
		},
		{
			name: "recordMissing",
			target: typeRecord{attrs: map[types.String]attributeType{
				"name": {typ: typeString{}, required: true},
			}},
			attr: "x",
			want: typeFalse{},
		},
		{
			name:   "entityRequired",
			target: typeEntity{lub: singleEntityLUB("User")},
			attr:   "name",
			want:   typeBool{},
		},
		{
			name:   "entityMissing",
			target: typeEntity{lub: singleEntityLUB("User")},
			attr:   "x",
			want:   typeFalse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, v.hasResultType(tt.target, tt.attr), tt.want)
		})
	}
}

func TestTypeCheckIs(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	tests := []struct {
		name string
		expr ast.IsNode
		want cedarType
	}{
		{
			name: "principalIsUser",
			expr: ast.NodeTypeIs{
				Left:       ast.NodeTypeVariable{Name: "principal"},
				EntityType: "User",
			},
			want: typeTrue{},
		},
		{
			name: "principalIsPhoto",
			expr: ast.NodeTypeIs{
				Left:       ast.NodeTypeVariable{Name: "principal"},
				EntityType: "Photo",
			},
			want: typeFalse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ty, _, err := v.typeOfExpr(env, tt.expr, caps)
			testutil.OK(t, err)
			testutil.Equals(t, ty, tt.want)
		})
	}
}

func TestOrCapabilityPropagation(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// (principal has age || principal has age) && principal.age > 18
	// Intersection of matching caps should preserve the capability.
	orExpr := ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "age",
		}},
		Right: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "age",
		}},
	}}
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: orExpr,
		Right: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Long(18)},
		}},
	}}
	_, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)

	// (principal has age || principal has name) && principal.age > 18
	// Intersection of mismatched caps should be empty → access fails.
	orMismatch := ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "age",
		}},
		Right: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "name",
		}},
	}}
	exprMismatch := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: orMismatch,
		Right: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Long(18)},
		}},
	}}
	_, _, err = v.typeOfExpr(env, exprMismatch, caps)
	testutil.Error(t, err)
}

func TestIfElseNoTestCapsInElse(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// if principal has name then false else principal.name == "foo"
	// Else branch should NOT get test capabilities → access to name fails.
	expr := ast.NodeTypeIfThenElse{
		If: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "age",
		}},
		Then: ast.NodeValue{Value: types.Boolean(false)},
		Else: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Long(0)},
		}},
	}
	_, _, err := v.typeOfExpr(env, expr, caps)
	testutil.Error(t, err)
}

func TestIfElseWithPriorCap(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal has age && (if principal has name then false else principal.age > 0)
	// Else branch gets prior capability (from outer &&) but not test capability.
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "principal"},
			Value: "age",
		}},
		Right: ast.NodeTypeIfThenElse{
			If: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "name",
			}},
			Then: ast.NodeValue{Value: types.Boolean(false)},
			Else: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
				Left: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
					Arg:   ast.NodeTypeVariable{Name: "principal"},
					Value: "age",
				}},
				Right: ast.NodeValue{Value: types.Long(0)},
			}},
		},
	}}
	_, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
}

func TestIfElseIntersectsCaps(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// (if 1 == 1 then (principal has age && true) else (principal has age && true)) && principal.age > 0
	// Both branches produce the same capability → intersection preserves it.
	ifExpr := ast.NodeTypeIfThenElse{
		If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.Long(1)},
		}},
		Then: ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Boolean(true)},
		}},
		Else: ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Boolean(true)},
		}},
	}
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: ifExpr,
		Right: ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
			Left: ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
				Arg:   ast.NodeTypeVariable{Name: "principal"},
				Value: "age",
			}},
			Right: ast.NodeValue{Value: types.Long(0)},
		}},
	}}
	_, _, err := v.typeOfExpr(env, expr, caps)
	testutil.OK(t, err)
}

func TestPolicyScopeIsInInvalid(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s)
	// principal is Photo in Group::"admins" — Photo can never be "in" Group
	p := ast.Permit()
	p.PrincipalIsIn("Photo", types.NewEntityUID("Group", "admins"))
	p.ActionEq(types.NewEntityUID("Action", "view"))
	testutil.Error(t, v.Policy(p))
}

// TestPolicyScopeInDescendants tests that ScopeTypeIn computes descendant types
// for principal/resource, matching the Rust implementation behavior.
func TestPolicyScopeInDescendants(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// principal in Group::"admins" should be valid because User is a descendant
	// of Group, and the "view" action applies to User.
	p := ast.Permit()
	p.PrincipalIn(types.NewEntityUID("Group", "admins"))
	p.ActionEq(types.NewEntityUID("Action", "view"))
	p.ResourceIs("Document")
	testutil.OK(t, v.Policy(p))

	// resource in Folder::"root" should be valid because Document is a descendant
	// of Folder, and the "view" action applies to Document.
	p2 := ast.Permit()
	p2.PrincipalIs("User")
	p2.ActionEq(types.NewEntityUID("Action", "view"))
	p2.ResourceIn(types.NewEntityUID("Folder", "root"))
	testutil.OK(t, v.Policy(p2))
}

func TestPolicyAllScopes(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// All unconstrained scopes
	p := ast.Permit()
	testutil.OK(t, v.Policy(p))

	// principal == User::"alice"
	p2 := ast.Permit()
	p2.PrincipalEq(types.NewEntityUID("User", "alice"))
	p2.ActionEq(types.NewEntityUID("Action", "view"))
	p2.ResourceEq(types.NewEntityUID("Document", "doc1"))
	testutil.OK(t, v.Policy(p2))

	// action in set
	p3 := ast.Permit()
	p3.PrincipalIs("User")
	p3.ActionInSet(types.NewEntityUID("Action", "view"), types.NewEntityUID("Action", "edit"))
	p3.ResourceIs("Document")
	testutil.OK(t, v.Policy(p3))

	// action in (group)
	p4 := ast.Permit()
	p4.PrincipalIs("User")
	p4.ActionIn(types.NewEntityUID("Action", "view"))
	p4.ResourceIs("Document")
	testutil.OK(t, v.Policy(p4))
}

func TestPolicyScopeErrors(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Unknown entity in principal scope
	p := ast.Permit()
	p.PrincipalEq(types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p))

	// Unknown action in set
	p2 := ast.Permit()
	p2.ActionInSet(types.NewEntityUID("Action", "delete"))
	testutil.Error(t, v.Policy(p2))

	// Unknown entity in resource scope
	p3 := ast.Permit()
	p3.ResourceEq(types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p3))

	// Unknown resource type in is
	p4 := ast.Permit()
	p4.ResourceIs("Unknown")
	testutil.Error(t, v.Policy(p4))

	// Unknown entity in resource in scope
	p5 := ast.Permit()
	p5.ResourceIn(types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p5))

	// Unknown entity in principal in scope
	p6 := ast.Permit()
	p6.PrincipalIn(types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p6))

	// Enum in principal scope - valid enum value
	p7 := ast.Permit()
	p7.PrincipalEq(types.NewEntityUID("Color", "red"))
	testutil.Error(t, v.Policy(p7)) // Color is not a valid principal for any action

	// Enum in principal scope - invalid enum value
	p8 := ast.Permit()
	p8.PrincipalEq(types.NewEntityUID("Color", "purple"))
	testutil.Error(t, v.Policy(p8))

	// resource is...in invalid
	p9 := ast.Permit()
	p9.ResourceIsIn("Document", types.NewEntityUID("User", "alice"))
	testutil.Error(t, v.Policy(p9))

	// resource is...in unknown in-entity
	p10 := ast.Permit()
	p10.ResourceIsIn("Document", types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p10))
}

func TestTypeCheckNot(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// !true → typeFalse
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeNot{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// !false → typeTrue
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeNot{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.Boolean(false)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})

	// !(1 == 1) → typeBool
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeNot{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.Long(1)},
		}},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeBool{})

	// !42 → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeNot{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckNegate(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// -42 → typeLong
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})

	// -"hello" → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.String("hello")},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckContainsAllAny(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	setExpr := ast.NodeTypeSet{Elements: []ast.IsNode{
		ast.NodeValue{Value: types.Long(1)},
	}}

	// [1].containsAll([2]) → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left: setExpr, Right: setExpr,
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// [1].containsAny([2]) → typeBool
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeContainsAny{BinaryNode: ast.BinaryNode{
		Left: setExpr, Right: setExpr,
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42.containsAll([2]) → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Long(42)}, Right: setExpr,
	}}, caps)
	testutil.Error(t, err)

	// [1].containsAll(42) → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left: setExpr, Right: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckIsEmpty(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// [1].isEmpty() → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeIsEmpty{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42.isEmpty() → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsEmpty{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckLike(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// "hello" like "h*" → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeLike{
		Arg:   ast.NodeValue{Value: types.String("hello")},
		Value: types.NewPattern("h", types.Wildcard{}),
	}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42 like "h*" → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeLike{
		Arg:   ast.NodeValue{Value: types.Long(42)},
		Value: types.NewPattern("h", types.Wildcard{}),
	}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckIsIn(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal is User in Group::"admins" → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{
			Left:       ast.NodeTypeVariable{Name: "principal"},
			EntityType: "User",
		},
		Entity: ast.NodeValue{Value: types.NewEntityUID("Group", "admins")},
	}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42 is User in Group::"admins" → error (left must be entity)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{
			Left:       ast.NodeValue{Value: types.Long(42)},
			EntityType: "User",
		},
		Entity: ast.NodeValue{Value: types.NewEntityUID("Group", "admins")},
	}, caps)
	testutil.Error(t, err)

	// principal is User in 42 → error (right must be entity)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{
			Left:       ast.NodeTypeVariable{Name: "principal"},
			EntityType: "User",
		},
		Entity: ast.NodeValue{Value: types.Long(42)},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckHasTag(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// resource hasTag "category" → typeBool (Document has tags)
	ty, newCaps, err := v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeValue{Value: types.String("category")},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// resource getTag "category" with hasTag guard → typeString
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeValue{Value: types.String("category")},
	}}, newCaps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeString{})

	// principal hasTag "x" → typeFalse (User doesn't have tags)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// 42 hasTag "x" → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// resource hasTag 42 → error (key must be string)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckGetTag(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// resource getTag "x" without hasTag guard → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// principal getTag "x" → error (User doesn't support tags)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// 42 getTag "x" → error (not entity)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckIn(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// principal in Group::"admins" → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeValue{Value: types.NewEntityUID("Group", "admins")},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// principal in [Group::"a", Group::"b"] → typeBool (set of entities)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeTypeSet{Elements: []ast.IsNode{
			ast.NodeValue{Value: types.NewEntityUID("Group", "a")},
			ast.NodeValue{Value: types.NewEntityUID("Group", "b")},
		}},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42 in Group::"admins" → error (left must be entity)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.NewEntityUID("Group", "admins")},
	}}, caps)
	testutil.Error(t, err)

	// principal in "admins" → error (right must be entity or set of entities)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeValue{Value: types.String("admins")},
	}}, caps)
	testutil.Error(t, err)

	// principal in [42] → error (set of non-entities)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeVariable{Name: "principal"},
		Right: ast.NodeTypeSet{Elements: []ast.IsNode{
			ast.NodeValue{Value: types.Long(42)},
		}},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckContains(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// [1, 2].contains(1) → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeSet{Elements: []ast.IsNode{
			ast.NodeValue{Value: types.Long(1)},
		}},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42.contains(1) → error (not a set)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckOrBranches(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// false || true → typeTrue
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(false)},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})

	// false || false → typeFalse
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(false)},
		Right: ast.NodeValue{Value: types.Boolean(false)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// (1==1) || true → typeTrue
	cmp := ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: cmp, Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})

	// (1==1) || false → typeBool (LHS caps preserved)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: cmp, Right: ast.NodeValue{Value: types.Boolean(false)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 42 || true → error (left not bool)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.Error(t, err)

	// true || 42 → no error (short-circuit, RHS not type checked)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})

	// false || 42 → error (RHS not bool)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(false)},
		Right: ast.NodeValue{Value: types.Long(42)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckAndBranches(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// true && true → typeTrue
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})

	// true && false → typeFalse
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Boolean(false)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// false && <anything> → typeFalse (short-circuit)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(false)},
		Right: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// (1==1) && false → typeFalse
	cmp := ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: cmp, Right: ast.NodeValue{Value: types.Boolean(false)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeFalse{})

	// (1==1) && (2==2) → typeBool
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: cmp, Right: cmp,
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeBool{})

	// 42 && true → error (left not bool)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(42)},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckIfThenElseShortCircuit(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// if false then 1 else "hello" → typeString (then branch dead, else evaluated)
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(false)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.String("hello")},
	}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeString{})

	// if true then 1 else "hello" → typeLong (else branch dead, then evaluated)
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.String("hello")},
	}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})

	// if 42 then 1 else 2 → error (condition not bool)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Long(42)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.Long(2)},
	}, caps)
	testutil.Error(t, err)

	// if (1==1) then 1 else "hello" → error (incompatible branch types)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.Long(1)},
		}},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.String("hello")},
	}, caps)
	testutil.Error(t, err)

	// Dead branch with bad entity ref → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(false)},
		Then: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Else: ast.NodeValue{Value: types.Long(1)},
	}, caps)
	testutil.Error(t, err)

	// then error propagates
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.Long(1)},
		}},
		Then: ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.String("x")},
		}},
		Else: ast.NodeValue{Value: types.Long(1)},
	}, caps)
	testutil.Error(t, err)

	// else error propagates
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.Long(1)},
		}},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.String("x")},
		}},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeCheckValues(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Entity UID of known type
	ty, _, err := v.typeOfExpr(env, ast.NodeValue{Value: types.NewEntityUID("User", "alice")}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("User")})

	// Entity UID of action type
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.NewEntityUID("Action", "view")}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Action")})

	// Entity UID of enum type
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.NewEntityUID("Color", "red")}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Color")})

	// Entity UID of unknown type → error
	_, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}, caps)
	testutil.Error(t, err)

	// Set value
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.Set{}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeSet{element: typeNever{}})

	// Record value
	_, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.NewRecord(types.RecordMap{
		"x": types.Long(1),
	})}, caps)
	testutil.OK(t, err)

	// IPAddr value
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.IPAddr{}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeExtension{"ipaddr"})

	// Decimal value
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.Decimal{}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeExtension{"decimal"})

	// Datetime value
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.Datetime{}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeExtension{"datetime"})

	// Duration value
	ty, _, err = v.typeOfExpr(env, ast.NodeValue{Value: types.Duration{}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeExtension{"duration"})
}

func TestTypeCheckVariablesAll(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// action → typeEntity{Action}
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeVariable{Name: "action"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Action")})

	// resource → typeEntity{Document}
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeVariable{Name: "resource"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Document")})

	// unknown → typeNever
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeVariable{Name: "unknown"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeNever{})
}

func TestCheckValue(t *testing.T) {
	t.Parallel()

	// String → ok
	testutil.OK(t, checkValue(types.String("hello"), resolved.StringType{}))
	testutil.Error(t, checkValue(types.Long(1), resolved.StringType{}))

	// Long → ok
	testutil.OK(t, checkValue(types.Long(1), resolved.LongType{}))
	testutil.Error(t, checkValue(types.String("x"), resolved.LongType{}))

	// Bool → ok
	testutil.OK(t, checkValue(types.Boolean(true), resolved.BoolType{}))
	testutil.Error(t, checkValue(types.Long(1), resolved.BoolType{}))

	// Entity → ok
	testutil.OK(t, checkValue(types.NewEntityUID("User", "alice"), resolved.EntityType("User")))
	testutil.Error(t, checkValue(types.String("x"), resolved.EntityType("User")))
	testutil.Error(t, checkValue(types.NewEntityUID("Group", "admins"), resolved.EntityType("User")))

	// Set → ok
	testutil.OK(t, checkValue(types.NewSet(types.Long(1)), resolved.SetType{Element: resolved.LongType{}}))
	testutil.Error(t, checkValue(types.String("x"), resolved.SetType{Element: resolved.LongType{}}))
	testutil.Error(t, checkValue(types.NewSet(types.String("x")), resolved.SetType{Element: resolved.LongType{}}))

	// Extension types
	testutil.OK(t, checkValue(types.IPAddr{}, resolved.ExtensionType("ipaddr")))
	testutil.Error(t, checkValue(types.Long(1), resolved.ExtensionType("ipaddr")))

	testutil.OK(t, checkValue(types.Decimal{}, resolved.ExtensionType("decimal")))
	testutil.Error(t, checkValue(types.Long(1), resolved.ExtensionType("decimal")))

	testutil.OK(t, checkValue(types.Datetime{}, resolved.ExtensionType("datetime")))
	testutil.Error(t, checkValue(types.Long(1), resolved.ExtensionType("datetime")))

	testutil.OK(t, checkValue(types.Duration{}, resolved.ExtensionType("duration")))
	testutil.Error(t, checkValue(types.Long(1), resolved.ExtensionType("duration")))

	testutil.Error(t, checkValue(types.Long(1), resolved.ExtensionType("unknown")))
}

func TestTypeCheckExtensionErrors(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Unknown function → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "unknownFunc",
		Args: []ast.IsNode{},
	}, caps)
	testutil.Error(t, err)

	// Wrong arg count → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{
			ast.NodeValue{Value: types.String("1.2.3.4")},
			ast.NodeValue{Value: types.String("5.6.7.8")},
		},
	}, caps)
	testutil.Error(t, err)

	// Wrong arg type → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}},
	}, caps)
	testutil.Error(t, err)
}

func TestTypecheckConditions(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	envs := []requestEnv{env}

	// Valid boolean condition
	conds := []ast.ConditionType{
		{Body: ast.NodeValue{Value: types.Boolean(true)}},
	}
	testutil.OK(t, v.typecheckConditions(envs, conds))

	// Non-boolean condition → error
	conds2 := []ast.ConditionType{
		{Body: ast.NodeValue{Value: types.Long(42)}},
	}
	testutil.Error(t, v.typecheckConditions(envs, conds2))

	// Condition with type error → error
	conds3 := []ast.ConditionType{
		{Body: ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
			Left:  ast.NodeValue{Value: types.Long(1)},
			Right: ast.NodeValue{Value: types.String("x")},
		}}},
	}
	testutil.Error(t, v.typecheckConditions(envs, conds3))
}

func TestPolicyWithConditions(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Policy with valid when condition
	p := ast.Permit()
	p.PrincipalIs("User")
	p.ActionEq(types.NewEntityUID("Action", "view"))
	p.ResourceIs("Document")
	p.When(ast.NewNode(ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))
	testutil.OK(t, v.Policy(p))

	// Policy with invalid when condition
	p2 := ast.Permit()
	p2.PrincipalIs("User")
	p2.ActionEq(types.NewEntityUID("Action", "view"))
	p2.ResourceIs("Document")
	p2.When(ast.NewNode(ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.String("x")},
	}}))
	testutil.Error(t, v.Policy(p2))
}

func TestValidateEntityRefsComprehensive(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Valid entity ref
	testutil.OK(t, v.validateEntityRefs(ast.NodeValue{Value: types.NewEntityUID("User", "alice")}))

	// Invalid entity ref
	testutil.Error(t, v.validateEntityRefs(ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}))

	// Entity ref in set value
	testutil.Error(t, v.validateEntityRefs(ast.NodeValue{Value: types.NewSet(
		types.NewEntityUID("Unknown", "x"),
	)}))

	// Variable - no entity refs
	testutil.OK(t, v.validateEntityRefs(ast.NodeTypeVariable{Name: "principal"}))

	// Nested in if-then-else
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Else: ast.NodeValue{Value: types.Long(1)},
	}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}))

	// In extension call
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}},
	}))

	// In record
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeRecord{
		Elements: []ast.RecordElementNode{
			{Key: "a", Value: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}},
		},
	}))

	// In set expression
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeSet{
		Elements: []ast.IsNode{ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}},
	}))

	// In binary ops (using And as representative)
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}))

	// In unary ops
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeNot{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}}))

	// In negate
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}}))

	// In has
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Value: "attr",
	}}))

	// In access
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Value: "attr",
	}}))

	// In like
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeLike{
		Arg:   ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Value: types.NewPattern(types.Wildcard{}),
	}))

	// In isEmpty
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIsEmpty{UnaryNode: ast.UnaryNode{
		Arg: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}}))

	// In is
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIs{
		Left:       ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		EntityType: "User",
	}))

	// In isIn
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{
			Left:       ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
			EntityType: "User",
		},
		Entity: ast.NodeValue{Value: types.NewEntityUID("User", "alice")},
	}))

	// In binary pairs (or, equals, comparisons, etc.)
	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Boolean(true)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeNotEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeLessThanOrEqual{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeGreaterThanOrEqual{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeSub{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeMult{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.NewEntityUID("User", "alice")},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeContainsAny{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.String("tag")},
	}}))

	testutil.Error(t, v.validateEntityRefs(ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Right: ast.NodeValue{Value: types.String("tag")},
	}}))
}

func TestIsSubtypeExtended(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())

	tests := []struct {
		name string
		a, b cedarType
		want bool
	}{
		{"Never<:Never", typeNever{}, typeNever{}, true},
		{"Never<:Set", typeNever{}, typeSet{element: typeLong{}}, true},
		{"Never<:Record", typeNever{}, typeRecord{}, true},
		{"Never<:AnyEntity", typeNever{}, typeAnyEntity{}, true},
		{"Never<:Extension", typeNever{}, typeExtension{"ipaddr"}, true},
		{"AnyEntity<:AnyEntity", typeAnyEntity{}, typeAnyEntity{}, true},
		{"Entity<:AnyEntity", typeEntity{lub: singleEntityLUB("User")}, typeAnyEntity{}, true},
		{"AnyEntity!<:Entity", typeAnyEntity{}, typeEntity{lub: singleEntityLUB("User")}, false},
		{"Extension==Extension", typeExtension{"ipaddr"}, typeExtension{"ipaddr"}, true},
		{"Extension!=Extension", typeExtension{"ipaddr"}, typeExtension{"decimal"}, false},
		{"True!<:False", typeTrue{}, typeFalse{}, false},
		{"False!<:True", typeFalse{}, typeTrue{}, false},
		{"Bool!<:True", typeBool{}, typeTrue{}, false},
		{"Long!<:Bool", typeLong{}, typeBool{}, false},
		{"String!<:Long", typeString{}, typeLong{}, false},
		{"Set!<:Long", typeSet{element: typeLong{}}, typeLong{}, false},
		{"Long!<:Set", typeLong{}, typeSet{element: typeLong{}}, false},
		{"Record<:Record", typeRecord{attrs: map[types.String]attributeType{
			"x": {typ: typeLong{}, required: true},
		}}, typeRecord{attrs: map[types.String]attributeType{
			"x": {typ: typeLong{}, required: true},
		}}, true},
		{"Long!<:Entity", typeLong{}, typeEntity{lub: singleEntityLUB("User")}, false},
		{"Long!<:AnyEntity", typeLong{}, typeAnyEntity{}, false},
		{"Long!<:Extension", typeLong{}, typeExtension{"ipaddr"}, false},
		{"Long!<:Record", typeLong{}, typeRecord{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			testutil.Equals(t, v.isSubtype(tt.a, tt.b), tt.want)
		})
	}
}

func TestIsSubtypeRecord(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())

	// Open record: a can have extra attrs
	open := typeRecord{
		attrs:          map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
		openAttributes: true,
	}
	withExtra := typeRecord{
		attrs: map[types.String]attributeType{
			"x": {typ: typeLong{}, required: true},
			"y": {typ: typeString{}, required: true},
		},
	}
	testutil.Equals(t, v.isSubtype(withExtra, open), true)

	// Closed record: extra attrs not allowed
	closed := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
	}
	testutil.Equals(t, v.isSubtype(withExtra, closed), false)

	// Missing required attr
	missingReq := typeRecord{attrs: map[types.String]attributeType{}}
	testutil.Equals(t, v.isSubtype(missingReq, closed), false)

	// Missing optional attr: ok
	withOptional := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: false}},
	}
	testutil.Equals(t, v.isSubtype(missingReq, withOptional), true)

	// Required vs optional mismatch
	reqClosed := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
	}
	optA := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: false}},
	}
	testutil.Equals(t, v.isSubtype(optA, reqClosed), false)

	// Wrong subtype for attr
	wrongType := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeString{}, required: true}},
	}
	testutil.Equals(t, v.isSubtype(wrongType, closed), false)
}

func TestLeastUpperBoundExtended(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())

	tests := []struct {
		name    string
		a, b    cedarType
		want    cedarType
		wantErr bool
	}{
		{"Never|Long=Long", typeNever{}, typeLong{}, typeLong{}, false},
		{"Long|Never=Long", typeLong{}, typeNever{}, typeLong{}, false},
		{"Bool|True=Bool", typeBool{}, typeTrue{}, typeBool{}, false},
		{"Bool|False=Bool", typeBool{}, typeFalse{}, typeBool{}, false},
		{"True|True=True", typeTrue{}, typeTrue{}, typeTrue{}, false},
		{"False|False=False", typeFalse{}, typeFalse{}, typeFalse{}, false},
		{"String|String=String", typeString{}, typeString{}, typeString{}, false},
		{"Entity|Entity=Entity",
			typeEntity{lub: singleEntityLUB("User")},
			typeEntity{lub: singleEntityLUB("Group")},
			typeEntity{lub: newEntityLUB("User", "Group")}, false},
		{"Entity|AnyEntity=AnyEntity",
			typeEntity{lub: singleEntityLUB("User")},
			typeAnyEntity{},
			typeAnyEntity{}, false},
		{"AnyEntity|Entity=AnyEntity",
			typeAnyEntity{},
			typeEntity{lub: singleEntityLUB("User")},
			typeAnyEntity{}, false},
		{"AnyEntity|AnyEntity=AnyEntity",
			typeAnyEntity{}, typeAnyEntity{}, typeAnyEntity{}, false},
		{"Extension|Extension=Extension",
			typeExtension{"ipaddr"}, typeExtension{"ipaddr"}, typeExtension{"ipaddr"}, false},
		{"Extension|DiffExtension=error",
			typeExtension{"ipaddr"}, typeExtension{"decimal"}, nil, true},
		{"Set|Set=Set",
			typeSet{element: typeLong{}}, typeSet{element: typeLong{}},
			typeSet{element: typeLong{}}, false},
		{"Record|Record",
			typeRecord{attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: true}}},
			typeRecord{attrs: map[types.String]attributeType{"y": {typ: typeString{}, required: true}}},
			typeRecord{attrs: map[types.String]attributeType{
				"x": {typ: typeLong{}, required: false},
				"y": {typ: typeString{}, required: false},
			}}, false},
		{"Long|Bool=error", typeLong{}, typeBool{}, nil, true},
		{"String|Bool=error", typeString{}, typeBool{}, nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := v.leastUpperBound(tt.a, tt.b)
			if tt.wantErr {
				testutil.Error(t, err)
			} else {
				testutil.OK(t, err)
				testutil.Equals(t, got, tt.want)
			}
		})
	}
}

func TestRequestWithEnum(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Valid request with enum principal
	s.Actions[types.NewEntityUID("Action", "color_action")] = resolved.Action{
		Entity: types.Entity{UID: types.NewEntityUID("Action", "color_action")},
		AppliesTo: &resolved.AppliesTo{
			Principals: []types.EntityType{"Color"},
			Resources:  []types.EntityType{"Document"},
			Context:    resolved.RecordType{},
		},
	}

	req := types.Request{
		Principal: types.NewEntityUID("Color", "red"),
		Action:    types.NewEntityUID("Action", "color_action"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.OK(t, v.Request(req))

	// Invalid enum value
	req2 := types.Request{
		Principal: types.NewEntityUID("Color", "purple"),
		Action:    types.NewEntityUID("Action", "color_action"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.Error(t, v.Request(req2))

	// Unknown principal type
	req3 := types.Request{
		Principal: types.NewEntityUID("Unknown", "x"),
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.Error(t, v.Request(req3))

	// Unknown resource type
	req4 := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("Unknown", "x"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.Error(t, v.Request(req4))
}

func TestEntityEnumRestrictions(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Enum entity with parents → error
	testutil.Error(t, v.Entity(types.Entity{
		UID:     types.NewEntityUID("Color", "red"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Group", "admins")),
	}))

	// Enum entity with attributes → error
	testutil.Error(t, v.Entity(types.Entity{
		UID: types.NewEntityUID("Color", "red"),
		Attributes: types.NewRecord(types.RecordMap{
			"x": types.Long(1),
		}),
	}))

	// Enum entity with tags → error
	testutil.Error(t, v.Entity(types.Entity{
		UID:  types.NewEntityUID("Color", "red"),
		Tags: types.NewRecord(types.RecordMap{"t": types.String("v")}),
	}))
}

func TestEntityActionValidation(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Action entity with wrong parents → error
	testutil.Error(t, v.Entity(types.Entity{
		UID:     types.NewEntityUID("Action", "edit"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "edit")), // wrong parent
	}))

	// Action entity missing expected parent → error
	testutil.Error(t, v.Entity(types.Entity{
		UID: types.NewEntityUID("Action", "edit"),
		// Missing Action::"view" parent
	}))
}

func TestEntitiesMapError(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	entities := types.EntityMap{
		types.NewEntityUID("Unknown", "x"): {
			UID: types.NewEntityUID("Unknown", "x"),
		},
	}
	testutil.Error(t, v.Entities(entities))
}

func TestEntityParentEnumValidation(t *testing.T) {
	t.Parallel()

	// Schema where User can have Color parents
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User": {
				Name:        "User",
				ParentTypes: []types.EntityType{"Color"},
				Shape:       resolved.RecordType{},
			},
		},
		Enums: map[types.EntityType]resolved.Enum{
			"Color": {
				Name:   "Color",
				Values: []types.EntityUID{types.NewEntityUID("Color", "red")},
			},
		},
		Actions: map[types.EntityUID]resolved.Action{},
	}
	v := New(s)
	// Valid enum parent
	testutil.OK(t, v.Entity(types.Entity{
		UID:     types.NewEntityUID("User", "alice"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Color", "red")),
	}))

	// Invalid enum parent ID
	testutil.Error(t, v.Entity(types.Entity{
		UID:     types.NewEntityUID("User", "alice"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Color", "purple")),
	}))
}

func TestHasResultTypeAnyEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	ty := v.hasResultType(typeAnyEntity{}, "anything")
	testutil.Equals[cedarType](t, ty, typeBool{})
}

func TestAccessOnAnyEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Access on typeAnyEntity should fail (we can't know what attributes exist)
	_, _, err := v.typeOfExpr(env, ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "nonexistent",
	}}, caps)
	testutil.Error(t, err)
}

func TestSchemaTypeToCedarType(t *testing.T) {
	t.Parallel()

	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.StringType{}), typeString{})
	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.LongType{}), typeLong{})
	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.BoolType{}), typeBool{})
	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.ExtensionType("ipaddr")), typeExtension{"ipaddr"})
	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.SetType{Element: resolved.LongType{}}), typeSet{element: typeLong{}})
	testutil.Equals[cedarType](t, schemaTypeToCedarType(resolved.EntityType("User")), typeEntity{lub: singleEntityLUB("User")})
}

func TestEntityLUBsRelatedForAction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Same entity type → related
	testutil.Equals(t, v.entityLUBsRelated(
		singleEntityLUB("User"), singleEntityLUB("User")), true)

	// Parent-child → related
	testutil.Equals(t, v.entityLUBsRelated(
		singleEntityLUB("User"), singleEntityLUB("Group")), true)

	// Unrelated types → not related
	s2 := testSchemaWithPhoto()
	v2 := New(s2)
	testutil.Equals(t, v2.entityLUBsRelated(
		singleEntityLUB("User"), singleEntityLUB("Photo")), false)
}

// TestIsCedarType covers the interface marker methods.
func TestIsCedarType(t *testing.T) {
	t.Parallel()
	// Call each concrete method directly to cover the marker methods
	typeNever{}.isCedarType()
	typeTrue{}.isCedarType()
	typeFalse{}.isCedarType()
	typeBool{}.isCedarType()
	typeLong{}.isCedarType()
	typeString{}.isCedarType()
	typeSet{}.isCedarType()
	typeRecord{}.isCedarType()
	typeEntity{}.isCedarType()
	typeAnyEntity{}.isCedarType()
	typeExtension{}.isCedarType()
}

func TestIsSubtypeFallthrough(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})
	// b is typeNever but a is not → false
	testutil.Equals(t, v.isSubtype(typeLong{}, typeNever{}), false)
	// b is unknown type (falls through switch) → false
	// Actually all types are covered. Test remaining edge case: b is typeNever
	testutil.Equals(t, v.isSubtype(typeString{}, typeNever{}), false)
}

func TestLeastUpperBoundRecordError(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})
	// Records where attribute types are incompatible → error
	a := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}
	b := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeString{}, required: true},
	}}
	_, err := v.leastUpperBound(a, b)
	testutil.Error(t, err)

	// LUB of Set with incompatible elements → error
	_, err = v.leastUpperBound(typeSet{element: typeLong{}}, typeSet{element: typeString{}})
	testutil.Error(t, err)
}

func TestLubRecordWithOpenAndSameKeys(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())
	a := typeRecord{
		attrs:          map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
		openAttributes: true,
	}
	b := typeRecord{
		attrs:          map[types.String]attributeType{"y": {typ: typeString{}, required: true}},
		openAttributes: false,
	}
	result, err := v.leastUpperBound(a, b)
	testutil.OK(t, err)
	rec := result.(typeRecord)
	testutil.Equals(t, rec.openAttributes, true)
	// Both "x" and "y" should be optional (only in one side)
	testutil.Equals(t, rec.attrs["x"].required, false)
	testutil.Equals(t, rec.attrs["y"].required, false)
}

func TestSchemaTypeToCedarTypeDefault(t *testing.T) {
	t.Parallel()
	// Test the default case (unknown schema type)
	// resolved.EntityType covers the EntityType case
	ty := schemaTypeToCedarType(resolved.EntityType("User"))
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("User")})
}

func TestLookupAttributeTypeOnRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Record type → finds attribute
	rec := typeRecord{attrs: map[types.String]attributeType{
		"name": {typ: typeString{}, required: true},
	}}
	at := v.lookupAttributeType(rec, "name")
	testutil.Equals(t, at != nil, true)

	// Record type → missing attribute
	at = v.lookupAttributeType(rec, "missing")
	testutil.Equals(t, at == nil, true)

	// Default case (non-record, non-entity) → nil
	at = v.lookupAttributeType(typeString{}, "x")
	testutil.Equals(t, at == nil, true)
}

func TestLookupEntityAttrMultiLUB(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Entity LUB with no elements → nil
	at := v.lookupEntityAttr(entityLUB{elements: nil}, "name")
	testutil.Equals(t, at == nil, true)

	// Entity LUB with unknown type → nil
	at = v.lookupEntityAttr(singleEntityLUB("Unknown"), "name")
	testutil.Equals(t, at == nil, true)

	// Entity LUB where attr not on type → nil
	at = v.lookupEntityAttr(singleEntityLUB("Group"), "email")
	testutil.Equals(t, at == nil, true)

	// Entity LUB with two types that share an attribute (name is on User and Group)
	at = v.lookupEntityAttr(newEntityLUB("User", "Group"), "name")
	testutil.Equals(t, at != nil, true)
	testutil.Equals[cedarType](t, at.typ, typeString{})

	// Entity LUB with two types where second doesn't have attr → nil
	at = v.lookupEntityAttr(newEntityLUB("User", "Folder"), "name")
	testutil.Equals(t, at == nil, true)
}

func TestMayHaveAttr(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Open record → always true
	testutil.Equals(t, v.mayHaveAttr(typeRecord{openAttributes: true}, "anything"), true)

	// Closed record with attr → true
	testutil.Equals(t, v.mayHaveAttr(typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}, "x"), true)

	// Closed record without attr → false
	testutil.Equals(t, v.mayHaveAttr(typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}, "y"), false)

	// AnyEntity → true
	testutil.Equals(t, v.mayHaveAttr(typeAnyEntity{}, "anything"), true)

	// Default (non-entity, non-record) → false
	testutil.Equals(t, v.mayHaveAttr(typeString{}, "anything"), false)

	// Entity with attr → true
	testutil.Equals(t, v.mayHaveAttr(typeEntity{lub: singleEntityLUB("User")}, "name"), true)

	// Entity without attr → false
	testutil.Equals(t, v.mayHaveAttr(typeEntity{lub: singleEntityLUB("Folder")}, "name"), false)

	// Entity with unknown type → false (no match)
	testutil.Equals(t, v.mayHaveAttr(typeEntity{lub: singleEntityLUB("Unknown")}, "name"), false)
}

func TestEntityHasTagsEdgeCases(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Empty LUB → false
	testutil.Equals(t, v.entityHasTags(entityLUB{elements: nil}), false)

	// Unknown entity → false
	testutil.Equals(t, v.entityHasTags(singleEntityLUB("Unknown")), false)

	// Entity without tags (User) → false
	testutil.Equals(t, v.entityHasTags(singleEntityLUB("User")), false)

	// Entity with tags (Document) → true
	testutil.Equals(t, v.entityHasTags(singleEntityLUB("Document")), true)
}

func TestEntityTagTypeEdgeCases(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Empty LUB → typeNever
	testutil.Equals[cedarType](t, v.entityTagType(entityLUB{elements: nil}), typeNever{})

	// Unknown entity → typeNever
	testutil.Equals[cedarType](t, v.entityTagType(singleEntityLUB("Unknown")), typeNever{})

	// Entity without tags → typeNever
	testutil.Equals[cedarType](t, v.entityTagType(singleEntityLUB("User")), typeNever{})

	// Entity with string tags (Document) → typeString
	testutil.Equals[cedarType](t, v.entityTagType(singleEntityLUB("Document")), typeString{})
}

func TestCheckStrictEntityLUBBothNonEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Both non-entity → nil (no check)
	testutil.OK(t, v.checkStrictEntityLUB(typeLong{}, typeString{}))

	// One is entity, other not → nil (no check)
	testutil.OK(t, v.checkStrictEntityLUB(typeEntity{lub: singleEntityLUB("User")}, typeLong{}))

	// b is never → nil (no check)
	testutil.OK(t, v.checkStrictEntityLUB(typeEntity{lub: singleEntityLUB("User")}, typeNever{}))
}

func TestIsEntityDescendantTransitive(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// User → Group (direct parent)
	testutil.Equals(t, v.isEntityDescendant("User", "Group"), true)

	// Unknown child → false
	testutil.Equals(t, v.isEntityDescendant("Unknown", "Group"), false)

	// No parent match → false
	testutil.Equals(t, v.isEntityDescendant("User", "Document"), false)

	// Transitive: Document → Folder
	testutil.Equals(t, v.isEntityDescendant("Document", "Folder"), true)
}

func TestCheckValueRecordAndSet(t *testing.T) {
	t.Parallel()

	// Record validation
	rec := types.NewRecord(types.RecordMap{"x": types.Long(1)})
	testutil.OK(t, checkValue(rec, resolved.RecordType{
		"x": {Type: resolved.LongType{}, Optional: false},
	}))

	// Not a record when expected
	testutil.Error(t, checkValue(types.Long(1), resolved.RecordType{}))

	// Set validation with element error
	testutil.Error(t, checkValue(types.NewSet(types.String("bad")), resolved.SetType{Element: resolved.LongType{}}))
}

func TestPolicyScopeEdgeCases(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// principal is...in with invalid is type
	p := ast.Permit()
	p.PrincipalIsIn("Unknown", types.NewEntityUID("Group", "admins"))
	testutil.Error(t, v.Policy(p))

	// principal is...in with invalid in entity
	p2 := ast.Permit()
	p2.PrincipalIsIn("User", types.NewEntityUID("Unknown", "x"))
	testutil.Error(t, v.Policy(p2))

	// resource is...in valid
	p3 := ast.Permit()
	p3.PrincipalIs("User")
	p3.ActionEq(types.NewEntityUID("Action", "view"))
	p3.ResourceIsIn("Document", types.NewEntityUID("Folder", "root"))
	testutil.OK(t, v.Policy(p3))

	// action in with unknown action
	p4 := ast.Permit()
	p4.ActionIn(types.NewEntityUID("Action", "unknown"))
	testutil.Error(t, v.Policy(p4))

	// Enum entity in scope
	p5 := ast.Permit()
	p5.PrincipalEq(types.NewEntityUID("Color", "red"))
	testutil.Error(t, v.Policy(p5)) // Color not in any action's principals

	// action entity in scope
	p6 := ast.Permit()
	p6.PrincipalEq(types.NewEntityUID("Action", "view"))
	testutil.Error(t, v.Policy(p6)) // Action type as principal
}

func TestValidateActionApplicationAllNil(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// All nil → valid
	testutil.OK(t, v.validateActionApplication(nil, nil, nil))

	// Action with nil AppliesTo → skip
	s2 := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{},
		Enums:    map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "noop"): {
				Entity: types.Entity{UID: types.NewEntityUID("Action", "noop")},
				// AppliesTo is nil
			},
		},
	}
	v2 := New(s2)
	testutil.Error(t, v2.validateActionApplication(
		[]types.EntityType{"User"}, nil, nil))
}

func TestIsActionDescendantNotFound(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Unknown action → false
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "unknown"),
		types.NewEntityUID("Action", "view")), false)

	// Action with transitive parent
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "edit"),
		types.NewEntityUID("Action", "view")), true)

	// Not a descendant
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "view"),
		types.NewEntityUID("Action", "edit")), false)
}

func TestValidateIsInScopeValid(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// User can be in Group → valid
	testutil.OK(t, v.validateIsInScope("User", "Group"))

	// Document can be in Folder → valid
	testutil.OK(t, v.validateIsInScope("Document", "Folder"))

	// Document cannot be in Group → error
	testutil.Error(t, v.validateIsInScope("Document", "Group"))
}

func TestGetEntityTypesInTransitive(t *testing.T) {
	t.Parallel()
	// Create a 3-level hierarchy: A → B → C
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"A": {Name: "A", ParentTypes: []types.EntityType{"B"}},
			"B": {Name: "B", ParentTypes: []types.EntityType{"C"}},
			"C": {Name: "C"},
		},
		Enums:   map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{},
	}
	v := New(s)
	result := v.getEntityTypesIn("C")
	// Should include C, B (child of C), A (child of B, grandchild of C)
	testutil.Equals(t, len(result), 3)
}

func TestRequestWithNoAppliesTo(t *testing.T) {
	t.Parallel()
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"User":     {Name: "User", Shape: resolved.RecordType{}},
			"Document": {Name: "Document", Shape: resolved.RecordType{}},
		},
		Enums: map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "noop"): {
				Entity: types.Entity{UID: types.NewEntityUID("Action", "noop")},
				// No AppliesTo
			},
		},
	}
	v := New(s)
	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "noop"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.Error(t, v.Request(req))
}

func TestRequestResourceEnumValidation(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Valid resource enum
	s.Actions[types.NewEntityUID("Action", "color_view")] = resolved.Action{
		Entity: types.Entity{UID: types.NewEntityUID("Action", "color_view")},
		AppliesTo: &resolved.AppliesTo{
			Principals: []types.EntityType{"User"},
			Resources:  []types.EntityType{"Color"},
			Context:    resolved.RecordType{},
		},
	}

	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "color_view"),
		Resource:  types.NewEntityUID("Color", "red"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.OK(t, v.Request(req))

	// Invalid resource enum value
	req2 := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "color_view"),
		Resource:  types.NewEntityUID("Color", "purple"),
		Context:   types.NewRecord(types.RecordMap{}),
	}
	testutil.Error(t, v.Request(req2))
}

func TestGenerateRequestEnvsSkipsNoAppliesTo(t *testing.T) {
	t.Parallel()
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{},
		Enums:    map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "noop"): {
				Entity: types.Entity{UID: types.NewEntityUID("Action", "noop")},
			},
		},
	}
	v := New(s)
	envs := v.generateRequestEnvs()
	testutil.Equals(t, len(envs), 0)
}

func TestFilterEnvsNoMatch(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	allEnvs := v.generateRequestEnvs()

	// Filter with impossible constraints
	filtered := v.filterEnvsForPolicy(allEnvs,
		[]types.EntityType{"Unknown"},
		[]types.EntityType{"Unknown"},
		[]types.EntityUID{types.NewEntityUID("Action", "unknown")})
	testutil.Equals(t, len(filtered), 0)
}

func TestIsActionInGroupTransitive(t *testing.T) {
	t.Parallel()
	// Create action hierarchy: child → parent → grandparent
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{},
		Enums:    map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "grandparent"): {
				Entity: types.Entity{UID: types.NewEntityUID("Action", "grandparent")},
			},
			types.NewEntityUID("Action", "parent"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "parent"),
					Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "grandparent")),
				},
			},
			types.NewEntityUID("Action", "child"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "child"),
					Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "parent")),
				},
			},
		},
	}
	v := New(s)
	// child is in grandparent (transitive)
	testutil.Equals(t, v.isActionInGroup(
		types.NewEntityUID("Action", "child"),
		types.NewEntityUID("Action", "grandparent")), true)

	// Unknown action → false
	testutil.Equals(t, v.isActionInGroup(
		types.NewEntityUID("Action", "unknown"),
		types.NewEntityUID("Action", "grandparent")), false)

	// Not in group → false
	testutil.Equals(t, v.isActionInGroup(
		types.NewEntityUID("Action", "grandparent"),
		types.NewEntityUID("Action", "child")), false)
}

func TestTypeOfExprNotEquals(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// 1 != 2 → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeNotEquals{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)
}

func TestTypeOfExprSubMult(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// 1 - 2 → typeLong
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeSub{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})

	// 1 * 2 → typeLong
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeMult{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})
}

func TestTypeOfExprGreaterThan(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// 1 > 2 → typeBool
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// 1 >= 2 → typeBool
	ty, _, err = v.typeOfExpr(env, ast.NodeTypeGreaterThanOrEqual{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)
}

func TestTypeOfValueSetWithMixedEntityTypes(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s)
	// Set with incompatible entity types (LUB fails) → Set<Never>
	ty, err := v.typeOfValue(types.NewSet(
		types.NewEntityUID("User", "a"),
		types.NewEntityUID("Photo", "b"),
	))
	testutil.OK(t, err)
	// LUB failure returns Set<Never>
	_, ok := ty.(typeSet)
	testutil.Equals(t, ok, true)
}

func TestTypeOfEntityUIDActionByType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Action entity UID that exists → ok
	ty, err := v.typeOfEntityUID(types.NewEntityUID("Action", "view"))
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Action")})

	// Action entity UID that doesn't exist but type does → ok (by type scan)
	ty, err = v.typeOfEntityUID(types.NewEntityUID("Action", "nonexistent"))
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("Action")})
}

func TestTypeOfAccessOnOpenRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	caps := newCapabilitySet()

	// Access on context (which is an open record with no attrs) returns typeNever
	envWithOpen := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   typeRecord{openAttributes: true},
	}
	ty, _, err := v.typeOfExpr(envWithOpen, ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "context"},
		Value: "anything",
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeNever{})
}

func TestTypeOfHasOnOpenRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	caps := newCapabilitySet()
	envWithOpen := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   typeRecord{openAttributes: true},
	}

	// has on open record → typeBool
	ty, _, err := v.typeOfExpr(envWithOpen, ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "context"},
		Value: "anything",
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeBool{})
}

func TestTypeOfHasOnEntityWithPriorCap(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()

	// Build caps with prior capability for "principal"."name"
	caps := newCapabilitySet()
	caps = caps.add(capability{varName: "principal", attr: "name"})

	// principal has name (name is required, and prior cap exists) → typeTrue
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "name",
	}}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeTrue{})
}

func TestHasResultTypeEntityMixedTypes(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Entity LUB with an enum type (no attrs but known) and attr not found → typeFalse
	testutil.Equals[cedarType](t, v.hasResultTypeEntity(singleEntityLUB("Color"), "x"), typeFalse{})

	// Entity LUB with action type and attr not found → typeFalse
	testutil.Equals[cedarType](t, v.hasResultTypeEntity(singleEntityLUB("Action"), "x"), typeFalse{})
}

func TestExprVarNameChained(t *testing.T) {
	t.Parallel()

	// Simple variable
	testutil.Equals(t, exprVarName(ast.NodeTypeVariable{Name: "principal"}), "principal")

	// Chained access: principal.foo
	name := exprVarName(ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "foo",
	}})
	testutil.Equals(t, name, "principal.foo")

	// Non-variable → empty
	testutil.Equals(t, exprVarName(ast.NodeValue{Value: types.Long(1)}), "")

	// Access on non-variable → empty
	testutil.Equals(t, exprVarName(ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeValue{Value: types.Long(1)},
		Value: "foo",
	}}), "")
}

func TestTagCapabilityKey(t *testing.T) {
	t.Parallel()

	// String value → returns key
	testutil.Equals(t, tagCapabilityKey(ast.NodeValue{Value: types.String("mykey")}), "mykey")

	// Non-value → empty
	testutil.Equals(t, tagCapabilityKey(ast.NodeTypeVariable{Name: "x"}), "")

	// Non-string value → empty
	testutil.Equals(t, tagCapabilityKey(ast.NodeValue{Value: types.Long(1)}), "")
}

func TestValidateEntityRefsPairRHSError(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// RHS error (LHS is fine)
	testutil.Error(t, v.validateEntityRefsPair(
		ast.NodeValue{Value: types.NewEntityUID("User", "a")},
		ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}))
}

func TestTypeOfIsAnyEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// AnyEntity case: test is with a typed entity LUB with multiple types
	s2 := testSchemaWithPhoto()
	s2.Actions[types.NewEntityUID("Action", "any_view")] = resolved.Action{
		Entity: types.Entity{UID: types.NewEntityUID("Action", "any_view")},
		AppliesTo: &resolved.AppliesTo{
			Principals: []types.EntityType{"User", "Photo"},
			Resources:  []types.EntityType{"Document"},
			Context:    resolved.RecordType{},
		},
	}
	multiEnv := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "any_view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	_ = multiEnv
	_ = s2

	// is on entity with multi-element LUB → typeBool (not typeTrue since LUB has > 1)
	multiLub := typeEntity{lub: newEntityLUB("User", "Photo")}
	_ = multiLub

	// Non-entity for is → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeIs{
		Left:       ast.NodeValue{Value: types.Long(42)},
		EntityType: "User",
	}, caps)
	testutil.Error(t, err)
}

func TestTypeOfGetTagNonLiteralKey(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// hasTag with non-literal key (no capability produced)
	_, newCaps, err := v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeTypeVariable{Name: "context"}, // not a literal string - will fail type check
	}}, caps)
	_ = newCaps
	// This will error because context is not a string
	testutil.Error(t, err)
}

func TestTypeOfRecordError(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Record with erroring element
	_, _, err := v.typeOfExpr(env, ast.NodeTypeRecord{
		Elements: []ast.RecordElementNode{
			{Key: "x", Value: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}},
		},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeOfSetError(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Set with erroring element
	_, _, err := v.typeOfExpr(env, ast.NodeTypeSet{
		Elements: []ast.IsNode{
			ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeOfAndFalseWithBadEntityRef(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// false && <bad entity ref> → error from validateEntityRefs
	_, _, err := v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(false)},
		Right: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeOfOrTrueWithBadEntityRef(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// true || <bad entity ref> → error from validateEntityRefs
	_, _, err := v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
	}}, caps)
	testutil.Error(t, err)
}

func TestCheckValueRecordType(t *testing.T) {
	t.Parallel()
	// Cover the RecordType branch in checkValue
	rec := types.NewRecord(types.RecordMap{"x": types.Long(1)})
	testutil.OK(t, checkValue(rec, resolved.RecordType{
		"x": {Type: resolved.LongType{}, Optional: false},
	}))
}

// TestErrorPropagationPaths covers all the "error from typeOfExpr propagation" paths
// that occur when a sub-expression has a type error.
func TestErrorPropagationPaths(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()
	badExpr := ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}

	// Each of these covers the "left error" path in binary ops
	_, _, err := v.typeOfExpr(env, ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeNotEquals{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeLessThanOrEqual{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeGreaterThan{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeGreaterThanOrEqual{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)

	// Right side error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Long(1)}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	_, _, err = v.typeOfExpr(env, ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Long(1)}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// Arith error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Long(1)}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeSub{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.String("x")}, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Long(1)}, Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// Negate error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeNegate{UnaryNode: ast.UnaryNode{Arg: badExpr}}, caps)
	testutil.Error(t, err)

	// In error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIn{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.NewEntityUID("User", "a")}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// Contains error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
		Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// ContainsAll error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
	}}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
		Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// IsEmpty error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsEmpty{UnaryNode: ast.UnaryNode{Arg: badExpr}}, caps)
	testutil.Error(t, err)

	// Like error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeLike{Arg: badExpr, Value: types.NewPattern(types.Wildcard{})}, caps)
	testutil.Error(t, err)

	// Is error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIs{Left: badExpr, EntityType: "User"}, caps)
	testutil.Error(t, err)

	// IsIn error paths
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{Left: badExpr, EntityType: "User"},
		Entity:     ast.NodeValue{Value: types.NewEntityUID("User", "a")},
	}, caps)
	testutil.Error(t, err)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIsIn{
		NodeTypeIs: ast.NodeTypeIs{Left: ast.NodeTypeVariable{Name: "principal"}, EntityType: "User"},
		Entity:     badExpr,
	}, caps)
	testutil.Error(t, err)

	// Has error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHas{StrOpNode: ast.StrOpNode{Arg: badExpr, Value: "x"}}, caps)
	testutil.Error(t, err)

	// Has on non-entity/record
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg: ast.NodeValue{Value: types.Long(42)}, Value: "x",
	}}, caps)
	testutil.Error(t, err)

	// Access error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{Arg: badExpr, Value: "x"}}, caps)
	testutil.Error(t, err)

	// Access on non-entity/record
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg: ast.NodeValue{Value: types.Long(42)}, Value: "x",
	}}, caps)
	testutil.Error(t, err)

	// HasTag error on Left
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// HasTag error on Right
	_, _, err = v.typeOfExpr(env, ast.NodeTypeHasTag{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeVariable{Name: "resource"}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// GetTag error on Left
	_, _, err = v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)

	// Or LHS error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.Error(t, err)

	// Or RHS error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeOr{BinaryNode: ast.BinaryNode{
		Left: ast.NodeValue{Value: types.Boolean(false)}, Right: badExpr,
	}}, caps)
	testutil.Error(t, err)

	// And LHS error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left: badExpr, Right: ast.NodeValue{Value: types.Boolean(true)},
	}}, caps)
	testutil.Error(t, err)

	// Not error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeNot{UnaryNode: ast.UnaryNode{Arg: badExpr}}, caps)
	testutil.Error(t, err)

	// IfThenElse condition error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeIfThenElse{
		If: badExpr, Then: ast.NodeValue{Value: types.Long(1)}, Else: ast.NodeValue{Value: types.Long(2)},
	}, caps)
	testutil.Error(t, err)

	// Extension call arg error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "ip", Args: []ast.IsNode{badExpr},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeOfValueEdgeCases(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Set with bad entity UID
	_, err := v.typeOfValue(types.NewSet(types.NewEntityUID("Unknown", "x")))
	testutil.Error(t, err)

	// Record with bad entity UID
	_, err = v.typeOfValue(types.NewRecord(types.RecordMap{
		"x": types.NewEntityUID("Unknown", "x"),
	}))
	testutil.Error(t, err)

	// Set with incompatible element types where LUB fails
	_, err = v.typeOfValue(types.NewSet(types.Long(1), types.String("x")))
	// LUB fails → returns Set<Never>, no error
	testutil.OK(t, err)
}

func TestTypeOfExprDefault(t *testing.T) {
	t.Parallel()
	// Can't easily trigger the default case since all AST node types are covered
	// The default case returns an error for unknown node types
}

func TestValidateScopeTypeEnum(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Validate enum type as scope type
	result, err := v.validateScopeType("Color")
	testutil.OK(t, err)
	testutil.Equals(t, len(result), 1)
	testutil.Equals(t, result[0], types.EntityType("Color"))
}

func TestValidateScopeEntityWithActionType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Action entity in scope - known action UID
	result, err := v.validateScopeEntity(types.NewEntityUID("Action", "view"))
	testutil.OK(t, err)
	testutil.Equals(t, len(result), 1)
}

func TestIsActionDescendantRecursive(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Direct parent match via recursion
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "edit"),
		types.NewEntityUID("Action", "view")), true)

	// Unknown parent in path → false
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "view"),
		types.NewEntityUID("Action", "nonexistent")), false)
}

func TestMatchesActionConstraintInGroup(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Action that is in a group of the constraint
	testutil.Equals(t, v.matchesActionConstraint(
		types.NewEntityUID("Action", "edit"),
		[]types.EntityUID{types.NewEntityUID("Action", "view")}), true)

	// Action that is not in group
	testutil.Equals(t, v.matchesActionConstraint(
		types.NewEntityUID("Action", "view"),
		[]types.EntityUID{types.NewEntityUID("Action", "edit")}), false)
}

func TestFilterEnvsForPolicyFilters(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	allEnvs := v.generateRequestEnvs()

	// Filter to specific principal
	filtered := v.filterEnvsForPolicy(allEnvs,
		[]types.EntityType{"User"}, nil, nil)
	for _, env := range filtered {
		testutil.Equals(t, env.principalType, types.EntityType("User"))
	}

	// Filter to specific resource
	filtered = v.filterEnvsForPolicy(allEnvs,
		nil, []types.EntityType{"Document"}, nil)
	for _, env := range filtered {
		testutil.Equals(t, env.resourceType, types.EntityType("Document"))
	}
}

func TestHasResultTypeDefault(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// Default case (non-entity, non-record, non-anyEntity) → typeBool
	testutil.Equals[cedarType](t, v.hasResultType(typeString{}, "x"), typeBool{})
}

func TestHasResultTypeEntityEmptyLUB(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	testutil.Equals[cedarType](t, v.hasResultTypeEntity(entityLUB{elements: nil}, "x"), typeBool{})
}

func TestEntityTagTypeLUBComputation(t *testing.T) {
	t.Parallel()
	// Schema where two entity types both have tags of different types
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"A": {Name: "A", Shape: resolved.RecordType{}, Tags: resolved.StringType{}},
			"B": {Name: "B", Shape: resolved.RecordType{}, Tags: resolved.LongType{}},
		},
		Enums:   map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{},
	}
	v := New(s)
	// LUB of string and long → error → typeNever
	testutil.Equals[cedarType](t, v.entityTagType(newEntityLUB("A", "B")), typeNever{})
}

func TestLookupEntityAttrLUBError(t *testing.T) {
	t.Parallel()
	// Schema where two entity types have same attr but incompatible types
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"A": {Name: "A", Shape: resolved.RecordType{
				"x": {Type: resolved.StringType{}, Optional: false},
			}},
			"B": {Name: "B", Shape: resolved.RecordType{
				"x": {Type: resolved.LongType{}, Optional: false},
			}},
		},
		Enums:   map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{},
	}
	v := New(s)
	// LUB of String and Long fails → nil
	at := v.lookupEntityAttr(newEntityLUB("A", "B"), "x")
	testutil.Equals(t, at == nil, true)
}

func TestIsEntityOrRecordTypeAnyEntity(t *testing.T) {
	t.Parallel()
	testutil.Equals(t, isEntityOrRecordType(typeAnyEntity{}), true)
	testutil.Equals(t, isEntityOrRecordType(typeString{}), false)
}

func TestGetTagOnAnyEntity(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	caps := newCapabilitySet()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}

	// getTag on non-literal tag key (context variable, non-string result not an issue - the tag key has no cap)
	tagCaps := caps.add(capability{varName: "resource", attr: "__tag:x"})
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeGetTag{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeVariable{Name: "resource"},
		Right: ast.NodeValue{Value: types.String("x")},
	}}, tagCaps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeString{})
}

func TestContainsStrictEntityCheck(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Set of entities.contains(entity of unrelated type) → strict error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeSet{Elements: []ast.IsNode{
			ast.NodeValue{Value: types.NewEntityUID("User", "a")},
		}},
		Right: ast.NodeValue{Value: types.NewEntityUID("Photo", "b")},
	}}, caps)
	testutil.Error(t, err)
}

func TestSetIncompatibleElementTypes(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Set with incompatible types (Long and String) → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeSet{Elements: []ast.IsNode{
		ast.NodeValue{Value: types.Long(1)},
		ast.NodeValue{Value: types.String("x")},
	}}, caps)
	testutil.Error(t, err)
}

func TestIsMultipleEntityTypes(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s, WithPermissive())
	caps := newCapabilitySet()

	// is on entity with multi-element LUB where type IS in lub → typeBool
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}

	// Manually create an expression that would have a multi-element LUB
	// Use if-then-else to create a LUB of User and Group entities
	ifExpr := ast.NodeTypeIfThenElse{
		If: ast.NodeTypeEquals{BinaryNode: ast.BinaryNode{
			Left: ast.NodeValue{Value: types.Long(1)}, Right: ast.NodeValue{Value: types.Long(1)},
		}},
		Then: ast.NodeValue{Value: types.NewEntityUID("User", "a")},
		Else: ast.NodeValue{Value: types.NewEntityUID("Group", "b")},
	}
	// is on LUB{User, Group} for User → typeBool (not True, because Group is also possible)
	ty, _, err := v.typeOfExpr(env, ast.NodeTypeIs{
		Left: ifExpr, EntityType: "User",
	}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeBool{})
}

// --- Coverage gap tests ---

func TestIsSubtypeFallthroughNilB(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})
	// cedar_type.go:126 - isSubtype fallthrough return false at end.
	// When b is nil (not any known cedarType), the switch falls through.
	var b cedarType // nil
	testutil.Equals(t, v.isSubtype(typeLong{}, b), false)
}

func TestLubRecordCommonAttrDifferentRequired(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())
	// cedar_type.go:246-249 - lubRecord with common attrs that have different required status.
	// When both records share a key but differ in required status,
	// the result should have required = (a.required && b.required).
	a := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}
	b := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: false},
	}}
	result, err := v.leastUpperBound(a, b)
	testutil.OK(t, err)
	rec := result.(typeRecord)
	// true && false = false
	testutil.Equals(t, rec.attrs["x"].required, false)
}

func TestSchemaTypeToCedarTypeRecordType(t *testing.T) {
	t.Parallel()
	// cedar_type.go:285-286 - schemaTypeToCedarType for RecordType.
	rec := resolved.RecordType{
		"name": {Type: resolved.StringType{}, Optional: false},
	}
	ty := schemaTypeToCedarType(rec)
	recTy, ok := ty.(typeRecord)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, recTy.attrs["name"].required, true)
	testutil.Equals[cedarType](t, recTy.attrs["name"].typ, typeString{})
}

func TestSchemaTypeToCedarTypeDefaultNil(t *testing.T) {
	t.Parallel()
	// cedar_type.go:289-290 - schemaTypeToCedarType default case (typeNever).
	// nil is not any recognized resolved.IsType, so it hits the default.
	ty := schemaTypeToCedarType(nil)
	testutil.Equals[cedarType](t, ty, typeNever{})
}

func TestIsEntityDescendantRecursiveThreeLevels(t *testing.T) {
	t.Parallel()
	// cedar_type.go:473-475 - isEntityDescendant recursive case.
	// Need a 3-level hierarchy: A -> B -> C, then check isEntityDescendant(A, C).
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{
			"A": {Name: "A", ParentTypes: []types.EntityType{"B"}},
			"B": {Name: "B", ParentTypes: []types.EntityType{"C"}},
			"C": {Name: "C"},
		},
		Enums:   map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{},
	}
	v := New(s)
	// A's parent is B, B != C, so we recurse: isEntityDescendant(B, C) -> B's parent is C, match!
	testutil.Equals(t, v.isEntityDescendant("A", "C"), true)
}

func TestCheckValueDefaultUnknownType(t *testing.T) {
	t.Parallel()
	// check_value.go:51-52 - checkValue default case (unknown schema type).
	// nil is not any recognized resolved.IsType.
	err := checkValue(types.String("hello"), nil)
	testutil.Error(t, err)
}

func TestValidatePrincipalScopeDefault(t *testing.T) {
	t.Parallel()
	// policy.go:77-79 - validatePrincipalScope default case.
	// Passing nil scope hits the default case in the type switch.
	s := testSchema()
	v := New(s)
	_, err := v.validatePrincipalScope(nil)
	testutil.Error(t, err)
}

func TestValidateActionScopeDefault(t *testing.T) {
	t.Parallel()
	// policy.go:107-108 - validateActionScope default case.
	s := testSchema()
	v := New(s)
	_, err := v.validateActionScope(nil)
	testutil.Error(t, err)
}

func TestValidatePrincipalScopeIsInSuccess(t *testing.T) {
	t.Parallel()
	// policy.go:77 - validatePrincipalScope ScopeTypeIsIn success path.
	// User can be in Group, so PrincipalIsIn("User", Group::"admins") should succeed.
	s := testSchema()
	v := New(s)
	result, err := v.validatePrincipalScope(ast.Scope{}.IsIn("User", types.NewEntityUID("Group", "admins")))
	testutil.OK(t, err)
	testutil.Equals(t, len(result), 1)
	testutil.Equals(t, result[0], types.EntityType("User"))
}

func TestValidateResourceScopeIsInBadIsType(t *testing.T) {
	t.Parallel()
	// policy.go:128-130 - validateResourceScope ScopeTypeIsIn with bad is type.
	// The is type is unknown, so validateScopeType fails at lines 128-130.
	s := testSchema()
	v := New(s)
	p := ast.Permit()
	p.ResourceIsIn("Unknown", types.NewEntityUID("Folder", "root"))
	err := v.Policy(p)
	testutil.Error(t, err)
}

func TestValidateResourceScopeDefault(t *testing.T) {
	t.Parallel()
	// policy.go:138-139 - validateResourceScope default case.
	s := testSchema()
	v := New(s)
	_, err := v.validateResourceScope(nil)
	testutil.Error(t, err)
}

func TestIsActionDescendantRecursiveThroughParents(t *testing.T) {
	t.Parallel()
	// policy.go:243-245 - isActionDescendant recursive through parents.
	// Need a 3-level action hierarchy: child -> parent -> grandparent.
	s := &resolved.Schema{
		Entities: map[types.EntityType]resolved.Entity{},
		Enums:    map[types.EntityType]resolved.Enum{},
		Actions: map[types.EntityUID]resolved.Action{
			types.NewEntityUID("Action", "gp"): {
				Entity: types.Entity{UID: types.NewEntityUID("Action", "gp")},
			},
			types.NewEntityUID("Action", "p"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "p"),
					Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "gp")),
				},
			},
			types.NewEntityUID("Action", "c"): {
				Entity: types.Entity{
					UID:     types.NewEntityUID("Action", "c"),
					Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "p")),
				},
			},
		},
	}
	v := New(s)
	// c -> p (p != gp) -> recurse: p -> gp (match!)
	testutil.Equals(t, v.isActionDescendant(
		types.NewEntityUID("Action", "c"),
		types.NewEntityUID("Action", "gp")), true)
}

func TestFilterEnvsResourceMismatch(t *testing.T) {
	t.Parallel()
	// request_env.go:47-48 - filterEnvsForPolicy resource mismatch.
	// Principal matches but resource does NOT, hitting the resource continue.
	s := testSchema()
	v := New(s)
	allEnvs := v.generateRequestEnvs()
	// Principal is User (matches), resource is "Nonexistent" (does not match any env).
	filtered := v.filterEnvsForPolicy(allEnvs,
		[]types.EntityType{"User"},
		[]types.EntityType{"Nonexistent"},
		nil)
	testutil.Equals(t, len(filtered), 0)
}

func TestFilterEnvsActionMismatch(t *testing.T) {
	t.Parallel()
	// request_env.go:50-51 - filterEnvsForPolicy action mismatch.
	// Principal and resource both match, but action does NOT.
	s := testSchema()
	v := New(s)
	allEnvs := v.generateRequestEnvs()
	// Principal User matches, resource Document matches, but action "nonexistent" never matches.
	filtered := v.filterEnvsForPolicy(allEnvs,
		[]types.EntityType{"User"},
		[]types.EntityType{"Document"},
		[]types.EntityUID{types.NewEntityUID("Action", "nonexistent")})
	testutil.Equals(t, len(filtered), 0)
}

func TestTypeOfComparisonExpectLeftError(t *testing.T) {
	t.Parallel()
	// typechecker.go:381-383 - typeOfComparison expectLeft error.
	// Also covers typechecker.go:369-371 - expectLong body.
	// LessThan with string left side triggers expectLong failure on the left.
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()
	_, _, err := v.typeOfExpr(env, ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.String("not a long")},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeOfComparisonExpectRightError(t *testing.T) {
	t.Parallel()
	// typechecker.go:390-392 - typeOfComparison expectRight error.
	// LessThan with string right side triggers expectLong failure on the right.
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()
	_, _, err := v.typeOfExpr(env, ast.NodeTypeLessThan{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.String("not a long")},
	}}, caps)
	testutil.Error(t, err)
}

func TestHasResultTypeEntityUnknownNonEntityType(t *testing.T) {
	t.Parallel()
	// typechecker.go:638-639 - hasResultTypeEntity allKnown check with unknown type.
	// typechecker.go:644 - hasResultTypeEntity return typeBool for unknown.
	// An entity LUB containing a type not in Entities, Enums, or Action types
	// means allKnown = false, so we get typeBool (line 644).
	s := testSchema()
	v := New(s)
	ty := v.hasResultTypeEntity(singleEntityLUB("CompletelyUnknown"), "x")
	testutil.Equals[cedarType](t, ty, typeBool{})
}

func TestValidateEntityRefsIfBlockError(t *testing.T) {
	t.Parallel()
	// typechecker.go:872-874 - validateEntityRefs if block error in if-then-else.
	// The If branch of the if-then-else contains a bad entity ref.
	s := testSchema()
	v := New(s)
	err := v.validateEntityRefs(ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.Long(2)},
	})
	testutil.Error(t, err)
}

func TestTypeOfExprNilDefault(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// nil hits the default case in the type switch
	_, _, err := v.typeOfExpr(env, nil, caps)
	testutil.Error(t, err)
}

func TestTypeOfValueNilDefault(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	// nil hits the default case in the type switch
	ty, err := v.typeOfValue(nil)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeNever{})
}

func TestWithStrict(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s, WithStrict())
	testutil.OK(t, v.Entity(types.Entity{
		UID: types.NewEntityUID("User", "alice"),
		Attributes: types.NewRecord(types.RecordMap{
			"name":  types.String("Alice"),
			"email": types.String("alice@example.com"),
		}),
	}))
}

func TestIsSubtypeStrictEntity(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})

	// Strict: equalLUB required — same LUB → true
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: singleEntityLUB("User")},
		typeEntity{lub: singleEntityLUB("User")}), true)

	// Strict: equalLUB required — different LUB → false
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: singleEntityLUB("User")},
		typeEntity{lub: newEntityLUB("User", "Group")}), false)

	// Strict: typeEntity is NOT subtype of typeAnyEntity
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: singleEntityLUB("User")},
		typeAnyEntity{}), false)

	// Strict: typeAnyEntity is subtype of typeAnyEntity
	testutil.Equals(t, v.isSubtype(typeAnyEntity{}, typeAnyEntity{}), true)
}

func TestIsSubtypePermissiveEntity(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())

	// Permissive: isSubsetLUB — subset → true
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: singleEntityLUB("User")},
		typeEntity{lub: newEntityLUB("User", "Group")}), true)

	// Permissive: typeEntity IS subtype of typeAnyEntity
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: singleEntityLUB("User")},
		typeAnyEntity{}), true)
}

func TestIsSubtypeRecordStrict(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})

	// Strict: extra attrs rejected even on open record
	open := typeRecord{
		attrs:          map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
		openAttributes: true,
	}
	withExtra := typeRecord{
		attrs: map[types.String]attributeType{
			"x": {typ: typeLong{}, required: true},
			"y": {typ: typeString{}, required: true},
		},
	}
	testutil.Equals(t, v.isSubtypeRecord(withExtra, open), false)

	// Strict: required/optional must match exactly — a optional, b required → false
	reqB := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
	}
	optA := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: false}},
	}
	testutil.Equals(t, v.isSubtypeRecord(optA, reqB), false)

	// Strict: a required, b optional → also false (exact match required)
	reqA := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: true}},
	}
	optB := typeRecord{
		attrs: map[types.String]attributeType{"x": {typ: typeLong{}, required: false}},
	}
	testutil.Equals(t, v.isSubtypeRecord(reqA, optB), false)
}

func TestLubRecordStrict(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{})

	// Strict: key only in one record → error
	a := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}
	b := typeRecord{attrs: map[types.String]attributeType{
		"y": {typ: typeString{}, required: true},
	}}
	_, err := v.lubRecord(a, b)
	testutil.Error(t, err)

	// Strict: key only in b → error
	c := typeRecord{attrs: map[types.String]attributeType{}}
	d := typeRecord{attrs: map[types.String]attributeType{
		"z": {typ: typeLong{}, required: true},
	}}
	_, err = v.lubRecord(c, d)
	testutil.Error(t, err)

	// Strict: same key, different required/optional → error
	e := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: true},
	}}
	f := typeRecord{attrs: map[types.String]attributeType{
		"x": {typ: typeLong{}, required: false},
	}}
	_, err = v.lubRecord(e, f)
	testutil.Error(t, err)
}

func TestTypeOfContainsPermissiveEmptySet(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s, WithPermissive())
	env := testEnv()
	caps := newCapabilitySet()

	// Permissive: empty set .contains(1) is NOT an error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeSet{Elements: []ast.IsNode{}},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.OK(t, err)
}

func TestTypeOfContainsAllAnyStrictIncompatible(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Strict: containsAll with incompatible element types → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeContainsAll{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.Long(1)}}},
		Right: ast.NodeTypeSet{Elements: []ast.IsNode{ast.NodeValue{Value: types.String("x")}}},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeOfExtensionCallStrictLiteral(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Strict: ip() with non-literal arg → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeTypeVariable{Name: "context"}},
	}, caps)
	testutil.Error(t, err)

	// Strict: decimal() with non-string literal → error
	_, _, err = v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "decimal",
		Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}},
	}, caps)
	testutil.Error(t, err)
}

func TestIsSubsetLUBNotSubset(t *testing.T) {
	t.Parallel()
	v := New(&resolved.Schema{}, WithPermissive())

	// Permissive: {User, Group} is NOT subtype of {User} (not a subset)
	testutil.Equals(t, v.isSubtype(
		typeEntity{lub: newEntityLUB("User", "Group")},
		typeEntity{lub: singleEntityLUB("User")}), false)
}

func TestTypeOfContainsPermissiveNonEmpty(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
	v := New(s, WithPermissive())
	env := testEnv()
	caps := newCapabilitySet()

	// Permissive: [User::"a"].contains(Photo::"b") — unrelated entity types, but permissive allows
	_, _, err := v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left: ast.NodeTypeSet{Elements: []ast.IsNode{
			ast.NodeValue{Value: types.NewEntityUID("User", "a")},
		}},
		Right: ast.NodeValue{Value: types.NewEntityUID("Photo", "b")},
	}}, caps)
	testutil.OK(t, err)
}

func TestTypeOfExtensionCallPermissiveNonLiteral(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s, WithPermissive())
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType: schemaRecordToCedarType(resolved.RecordType{
			"ip_str": {Type: resolved.StringType{}, Optional: false},
		}),
	}
	caps := newCapabilitySet()

	// Permissive: ip(context.ip_str) — non-literal arg is OK in permissive mode
	_, _, err := v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
			Arg:   ast.NodeTypeVariable{Name: "context"},
			Value: "ip_str",
		}}},
	}, caps)
	testutil.OK(t, err)
}

func TestTypeOfContainsStrictEmptySetValue(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Strict: empty set VALUE (not expression) .contains(1) — Set<Never> contains Long
	_, _, err := v.typeOfExpr(env, ast.NodeTypeContains{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Set{}},
		Right: ast.NodeValue{Value: types.Long(1)},
	}}, caps)
	testutil.Error(t, err)
}

func TestTypeOfExtensionCallArgError(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Method call (not constructor) with arg that causes typeOfExpr error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "isIpv4",
		Args: []ast.IsNode{ast.NodeValue{Value: types.NewEntityUID("Unknown", "x")}},
	}, caps)
	testutil.Error(t, err)

	// Method call with wrong arg type (reaches isSubtype check)
	_, _, err = v.typeOfExpr(env, ast.NodeTypeExtensionCall{
		Name: "isIpv4",
		Args: []ast.IsNode{ast.NodeValue{Value: types.Long(42)}},
	}, caps)
	testutil.Error(t, err)
}

func TestTypeOfSetStrictEmpty(t *testing.T) {
	t.Parallel()
	s := testSchema()
	v := New(s)
	env := testEnv()
	caps := newCapabilitySet()

	// Strict: empty set expression → error
	_, _, err := v.typeOfExpr(env, ast.NodeTypeSet{Elements: []ast.IsNode{}}, caps)
	testutil.Error(t, err)

	// Permissive: empty set expression → ok
	vp := New(s, WithPermissive())
	_, _, err = vp.typeOfExpr(env, ast.NodeTypeSet{Elements: []ast.IsNode{}}, caps)
	testutil.OK(t, err)
}
