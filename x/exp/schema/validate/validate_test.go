package validate

import (
	"testing"

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

// --- Entity validation tests ---

func TestEntityValid(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID:     types.NewEntityUID("User", "alice"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Group", "admins")),
		Attributes: types.NewRecord(types.RecordMap{
			"name":  types.String("Alice"),
			"email": types.String("alice@example.com"),
		}),
	}
	if err := Entity(s, entity); err != nil {
		t.Fatalf("expected valid entity, got error: %v", err)
	}
}

func TestEntityUnknownType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("Unknown", "x"),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for unknown entity type")
	}
}

func TestEntityInvalidParentType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID:     types.NewEntityUID("User", "alice"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Document", "doc1")),
		Attributes: types.NewRecord(types.RecordMap{
			"name":  types.String("Alice"),
			"email": types.String("alice@example.com"),
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for invalid parent type")
	}
}

func TestEntityMissingRequiredAttr(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("User", "alice"),
		Attributes: types.NewRecord(types.RecordMap{
			"name": types.String("Alice"),
			// missing "email"
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for missing required attribute")
	}
}

func TestEntityUnexpectedAttr(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("User", "alice"),
		Attributes: types.NewRecord(types.RecordMap{
			"name":    types.String("Alice"),
			"email":   types.String("alice@example.com"),
			"unknown": types.String("extra"),
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for unexpected attribute")
	}
}

func TestEntityWrongAttrType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("User", "alice"),
		Attributes: types.NewRecord(types.RecordMap{
			"name":  types.Long(42), // wrong type
			"email": types.String("alice@example.com"),
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for wrong attribute type")
	}
}

func TestEntityTagsValid(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("Document", "doc1"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Folder", "folder1")),
		Attributes: types.NewRecord(types.RecordMap{
			"title":  types.String("My Doc"),
			"public": types.Boolean(true),
		}),
		Tags: types.NewRecord(types.RecordMap{
			"category": types.String("report"),
		}),
	}
	if err := Entity(s, entity); err != nil {
		t.Fatalf("expected valid entity with tags, got error: %v", err)
	}
}

func TestEntityTagsNotAllowed(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("User", "alice"),
		Attributes: types.NewRecord(types.RecordMap{
			"name":  types.String("Alice"),
			"email": types.String("alice@example.com"),
		}),
		Tags: types.NewRecord(types.RecordMap{
			"tag1": types.String("val"),
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for tags on entity that doesn't allow them")
	}
}

func TestEntityTagsWrongType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	entity := types.Entity{
		UID: types.NewEntityUID("Document", "doc1"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Folder", "folder1")),
		Attributes: types.NewRecord(types.RecordMap{
			"title":  types.String("My Doc"),
			"public": types.Boolean(true),
		}),
		Tags: types.NewRecord(types.RecordMap{
			"category": types.Long(42), // wrong type, should be String
		}),
	}
	err := Entity(s, entity)
	if err == nil {
		t.Fatal("expected error for wrong tag type")
	}
}

func TestEntityEnum(t *testing.T) {
	t.Parallel()
	s := testSchema()

	// Valid enum
	entity := types.Entity{
		UID: types.NewEntityUID("Color", "red"),
	}
	if err := Entity(s, entity); err != nil {
		t.Fatalf("expected valid enum entity, got error: %v", err)
	}

	// Invalid enum ID
	entity = types.Entity{
		UID: types.NewEntityUID("Color", "purple"),
	}
	if err := Entity(s, entity); err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}

func TestEntityAction(t *testing.T) {
	t.Parallel()
	s := testSchema()

	// Valid action entity
	entity := types.Entity{
		UID:     types.NewEntityUID("Action", "edit"),
		Parents: types.NewEntityUIDSet(types.NewEntityUID("Action", "view")),
	}
	if err := Entity(s, entity); err != nil {
		t.Fatalf("expected valid action entity, got error: %v", err)
	}

	// Invalid action - unknown
	entity = types.Entity{
		UID: types.NewEntityUID("Action", "delete"),
	}
	if err := Entity(s, entity); err == nil {
		t.Fatal("expected error for unknown action")
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
	if err := Entities(s, entities); err != nil {
		t.Fatalf("expected valid entities, got error: %v", err)
	}
}

// --- Request validation tests ---

func TestRequestValid(t *testing.T) {
	t.Parallel()
	s := testSchema()
	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context: types.NewRecord(types.RecordMap{
			"ip": types.IPAddr{},
		}),
	}
	if err := Request(s, req); err != nil {
		t.Fatalf("expected valid request, got error: %v", err)
	}
}

func TestRequestUnknownAction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "delete"),
		Resource:  types.NewEntityUID("Document", "doc1"),
	}
	err := Request(s, req)
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestRequestWrongPrincipalType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	req := types.Request{
		Principal: types.NewEntityUID("Document", "doc1"), // wrong type
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context: types.NewRecord(types.RecordMap{
			"ip": types.IPAddr{},
		}),
	}
	err := Request(s, req)
	if err == nil {
		t.Fatal("expected error for wrong principal type")
	}
}

func TestRequestWrongResourceType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("User", "bob"), // wrong type
		Context: types.NewRecord(types.RecordMap{
			"ip": types.IPAddr{},
		}),
	}
	err := Request(s, req)
	if err == nil {
		t.Fatal("expected error for wrong resource type")
	}
}

func TestRequestInvalidContext(t *testing.T) {
	t.Parallel()
	s := testSchema()
	req := types.Request{
		Principal: types.NewEntityUID("User", "alice"),
		Action:    types.NewEntityUID("Action", "view"),
		Resource:  types.NewEntityUID("Document", "doc1"),
		Context:   types.NewRecord(types.RecordMap{}), // missing required "ip"
	}
	err := Request(s, req)
	if err == nil {
		t.Fatal("expected error for invalid context")
	}
}

// --- Policy validation tests ---

func TestPolicyRBACUnknownEntityType(t *testing.T) {
	t.Parallel()
	s := testSchema()
	p := ast.Permit()
	p.PrincipalIs("Unknown")
	err := Policy(s, p)
	if err == nil {
		t.Fatal("expected error for unknown entity type in scope")
	}
}

func TestPolicyRBACUnknownAction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	p := ast.Permit()
	p.ActionEq(types.NewEntityUID("Action", "delete"))
	err := Policy(s, p)
	if err == nil {
		t.Fatal("expected error for unknown action in scope")
	}
}

func TestPolicyRBACInvalidActionApplication(t *testing.T) {
	t.Parallel()
	s := testSchema()
	p := ast.Permit()
	p.PrincipalIs("Group")
	p.ActionEq(types.NewEntityUID("Action", "view"))
	err := Policy(s, p)
	if err == nil {
		t.Fatal("expected error for invalid action application")
	}
}

func TestPolicyValidSimple(t *testing.T) {
	t.Parallel()
	s := testSchema()
	p := ast.Permit()
	p.PrincipalIs("User")
	p.ActionEq(types.NewEntityUID("Action", "view"))
	p.ResourceIs("Document")
	if err := Policy(s, p); err != nil {
		t.Fatalf("expected valid policy, got error: %v", err)
	}
}

// --- Expression type checking tests ---

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if et, ok := ty.(typeEntity); !ok || len(et.lub.elements) != 1 || et.lub.elements[0] != "User" {
		t.Fatalf("expected typeEntity{User}, got %T", ty)
	}

	// context → typeRecord
	ty, _, err = typeOfExpr(env, s, ast.NodeTypeVariable{Name: "context"}, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeRecord); !ok {
		t.Fatalf("expected typeRecord, got %T", ty)
	}
}

func TestTypeCheckArithmetic(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// 1 + 2 → typeLong
	expr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.Long(2)},
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeLong); !ok {
		t.Fatalf("expected typeLong, got %T", ty)
	}

	// 1 + "two" → error
	badExpr := ast.NodeTypeAdd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Long(1)},
		Right: ast.NodeValue{Value: types.String("two")},
	}}
	_, _, err = typeOfExpr(env, s, badExpr, caps)
	if err == nil {
		t.Fatal("expected error for Long + String")
	}
}

func TestTypeCheckLogical(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// true && false → typeBool
	expr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Boolean(false)},
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBoolType(ty) {
		t.Fatalf("expected bool type, got %T", ty)
	}

	// true && 42 → error
	badExpr := ast.NodeTypeAnd{BinaryNode: ast.BinaryNode{
		Left:  ast.NodeValue{Value: types.Boolean(true)},
		Right: ast.NodeValue{Value: types.Long(42)},
	}}
	_, _, err = typeOfExpr(env, s, badExpr, caps)
	if err == nil {
		t.Fatal("expected error for true && Long")
	}
}

func TestTypeCheckAttributeAccess(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// principal.name → typeString (required attr)
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "name",
	}}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeString); !ok {
		t.Fatalf("expected typeString, got %T", ty)
	}

	// principal.age → error (optional attr without has guard)
	optExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	_, _, err = typeOfExpr(env, s, optExpr, caps)
	if err == nil {
		t.Fatal("expected error for accessing optional attr without has guard")
	}
}

func TestTypeCheckHasGuard(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// principal has age && principal.age > 18
	hasExpr := ast.NodeTypeHas{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}

	_, hasCaps, err := typeOfExpr(env, s, hasExpr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Now with hasCaps, principal.age should work
	accessExpr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "age",
	}}
	ty, _, err := typeOfExpr(env, s, accessExpr, hasCaps)
	if err != nil {
		t.Fatalf("expected access to succeed after has guard, got error: %v", err)
	}
	if _, ok := ty.(typeLong); !ok {
		t.Fatalf("expected typeLong, got %T", ty)
	}
}

func TestTypeCheckExtensionFunction(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// ip("127.0.0.1") → typeExtension{ipaddr}
	expr := ast.NodeTypeExtensionCall{
		Name: "ip",
		Args: []ast.IsNode{ast.NodeValue{Value: types.String("127.0.0.1")}},
	}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ext, ok := ty.(typeExtension); !ok || ext.name != "ipaddr" {
		t.Fatalf("expected typeExtension{ipaddr}, got %T", ty)
	}

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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !isBoolType(ty) {
		t.Fatalf("expected bool type, got %T", ty)
	}
}

func TestTypeCheckIfThenElse(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// if true then 1 else 2 → typeLong
	expr := ast.NodeTypeIfThenElse{
		If:   ast.NodeValue{Value: types.Boolean(true)},
		Then: ast.NodeValue{Value: types.Long(1)},
		Else: ast.NodeValue{Value: types.Long(2)},
	}
	ty, _, err := typeOfExpr(env, s, expr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeLong); !ok {
		t.Fatalf("expected typeLong, got %T", ty)
	}
}

func TestTypeCheckSetRecord(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
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
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st, ok := ty.(typeSet); !ok {
		t.Fatalf("expected typeSet, got %T", ty)
	} else if _, ok := st.element.(typeLong); !ok {
		t.Fatalf("expected set element type Long, got %T", st.element)
	}

	// {"a": 1, "b": "hello"} → typeRecord
	recExpr := ast.NodeTypeRecord{
		Elements: []ast.RecordElementNode{
			{Key: "a", Value: ast.NodeValue{Value: types.Long(1)}},
			{Key: "b", Value: ast.NodeValue{Value: types.String("hello")}},
		},
	}
	ty, _, err = typeOfExpr(env, s, recExpr, caps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec, ok := ty.(typeRecord); !ok {
		t.Fatalf("expected typeRecord, got %T", ty)
	} else {
		if a, ok := rec.attrs["a"]; !ok {
			t.Fatal("expected attr 'a'")
		} else if _, ok := a.typ.(typeLong); !ok {
			t.Fatalf("expected attr 'a' to be Long, got %T", a.typ)
		}
	}
}

func TestTypeCheckUnsafeAccess(t *testing.T) {
	t.Parallel()
	s := testSchema()
	env := &requestEnv{
		principalType: "User",
		actionUID:     types.NewEntityUID("Action", "view"),
		resourceType:  "Document",
		contextType:   schemaRecordToCedarType(resolved.RecordType{}),
	}
	caps := newCapabilitySet()

	// principal.nonexistent → error
	expr := ast.NodeTypeAccess{StrOpNode: ast.StrOpNode{
		Arg:   ast.NodeTypeVariable{Name: "principal"},
		Value: "nonexistent",
	}}
	_, _, err := typeOfExpr(env, s, expr, caps)
	if err == nil {
		t.Fatal("expected error for accessing nonexistent attribute")
	}
}

// --- Subtype and LUB tests ---

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
			got := isSubtype(tt.a, tt.b)
			if got != tt.want {
				t.Fatalf("isSubtype(%T, %T) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLeastUpperBound(t *testing.T) {
	t.Parallel()

	// True | False → Bool
	ty, err := leastUpperBound(typeTrue{}, typeFalse{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeBool); !ok {
		t.Fatalf("expected typeBool, got %T", ty)
	}

	// Long | Long → Long
	ty, err = leastUpperBound(typeLong{}, typeLong{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := ty.(typeLong); !ok {
		t.Fatalf("expected typeLong, got %T", ty)
	}

	// Long | String → error
	_, err = leastUpperBound(typeLong{}, typeString{})
	if err == nil {
		t.Fatal("expected error for incompatible types")
	}
}
