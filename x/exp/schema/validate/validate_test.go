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
			err := Entity(s, tt.entity)
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
	testutil.OK(t, Entities(s, entities))
}

func TestRequest(t *testing.T) {
	t.Parallel()
	s := testSchema()

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
			err := Request(s, tt.req)
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
			err := Policy(s, p)
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
	ty, _, err := typeOfExpr(env, s, ast.NodeTypeVariable{Name: "principal"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeEntity{lub: singleEntityLUB("User")})

	// context → typeRecord
	ty, _, err = typeOfExpr(env, s, ast.NodeTypeVariable{Name: "context"}, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, env.contextType)
}

func TestTypeCheckArithmetic(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// 1 + 2 → typeLong
	expr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})

	// 1 + "two" → error
	badExpr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.String("two")},
	}}
	_, _, err = typeOfExpr(env, s, badExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckLogical(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// true && false → typeBool
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Boolean(false)},
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)

	// true && 42 → error
	badExpr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Long(42)},
	}}
	_, _, err = typeOfExpr(env, s, badExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckAttributeAccess(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// principal.name → typeString (required attr)
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "name",
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeString{})

	// principal.age → error (optional attr without has guard)
	optExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	_, _, err = typeOfExpr(env, s, optExpr, caps)
	testutil.Error(t, err)
}

func TestTypeCheckHasGuard(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// principal has age && principal.age > 18
	hasExpr := ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}

	_, hasCaps, err := typeOfExpr(env, s, hasExpr, caps)
	testutil.OK(t, err)

	// Now with hasCaps, principal.age should work
	accessExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	ty, _, err := typeOfExpr(env, s, accessExpr, hasCaps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})
}

func TestTypeCheckExtensionFunction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// ip("127.0.0.1") → typeExtension{ipaddr}
	expr := ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeValue{Value: types.String("127.0.0.1")}},
	}
	ty, _, err := typeOfExpr(env, s, expr, caps)
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
	ty, _, err = typeOfExpr(env, s, expr2, caps)
	testutil.OK(t, err)
	testutil.Equals(t, isBoolType(ty), true)
}

func TestTypeCheckIfThenElse(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := testEnv()
	caps := newCapabilitySet()

	// if true then 1 else 2 → typeLong
	expr := ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.Long(2)},
	}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeLong{})
}

func TestTypeCheckSetRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
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
	ty, _, err := typeOfExpr(env, s, setExpr, caps)
	testutil.OK(t, err)
	testutil.Equals[cedarType](t, ty, typeSet{element: typeLong{}})

	// {"a": 1, "b": "hello"} → typeRecord
	recExpr := ast.NodeTypeRecord{
		Elements: []ast.RecordElementNode{
			{Key: "a", Value: ast.NodeValue{Value: types.Long(1)}},
			{Key: "b", Value: ast.NodeValue{Value: types.String("hello")}},
		},
	}
	ty, _, err = typeOfExpr(env, s, recExpr, caps)
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
	env := testEnv()
	caps := newCapabilitySet()

	// principal.nonexistent → error
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "nonexistent",
	}}
	_, _, err := typeOfExpr(env, s, expr, caps)
	testutil.Error(t, err)
}

func TestIsSubtype(t *testing.T) {
	t.Parallel()

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
			testutil.Equals(t, isSubtype(tt.a, tt.b), tt.want)
		})
	}
}

func TestLeastUpperBound(t *testing.T) {
	t.Parallel()

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
			got, err := leastUpperBound(tt.a, tt.b)
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
			_, _, err := typeOfExpr(env, s, tt.expr, caps)
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
			testutil.Equals(t, hasResultType(s, tt.target, tt.attr), tt.want)
		})
	}
}

func TestTypeCheckIs(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()
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
			want: typeBool{},
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
			ty, _, err := typeOfExpr(env, s, tt.expr, caps)
			testutil.OK(t, err)
			testutil.Equals(t, ty, tt.want)
		})
	}
}

func TestPolicyScopeIsInInvalid(t *testing.T) {
	t.Parallel()
	s := testSchemaWithPhoto()

	// principal is Photo in Group::"admins" — Photo can never be "in" Group
	p := ast.Permit()
	p.PrincipalIsIn("Photo", types.NewEntityUID("Group", "admins"))
	p.ActionEq(types.NewEntityUID("Action", "view"))
	testutil.Error(t, Policy(s, p))
}
