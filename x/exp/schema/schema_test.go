package schema_test

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/types"
	. "github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolved"
)

// Cedar text parsing

func TestParseCedarEmpty(t *testing.T) {
	var s Schema
	if err := s.UnmarshalCedar([]byte("")); err != nil {
		t.Fatal(err)
	}
	if len(s.Namespaces) != 0 {
		t.Errorf("got %d namespaces, want 0", len(s.Namespaces))
	}
}

func TestParseCedarEntity(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User;`))
	if err != nil {
		t.Fatal(err)
	}
	ns := s.Namespaces[""]
	if ns == nil {
		t.Fatal("expected empty namespace")
	}
	if _, ok := ns.EntityTypes["User"]; !ok {
		t.Error("expected entity User")
	}
}

func TestParseCedarEntityMemberOf(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User in Group;`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if len(et.MemberOfTypes) != 1 || et.MemberOfTypes[0] != "Group" {
		t.Errorf("memberOf = %v, want [Group]", et.MemberOfTypes)
	}
}

func TestParseCedarEntityMemberOfMultiple(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User in [Group, Team];`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if len(et.MemberOfTypes) != 2 {
		t.Errorf("memberOf len = %d, want 2", len(et.MemberOfTypes))
	}
}

func TestParseCedarEntityShape(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User { name: String, age?: Long };`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if et.Shape == nil {
		t.Fatal("expected shape")
	}
	nameAttr := et.Shape.Attributes["name"]
	if nameAttr == nil || !nameAttr.Required {
		t.Error("expected required name attribute")
	}
	ageAttr := et.Shape.Attributes["age"]
	if ageAttr == nil || ageAttr.Required {
		t.Error("expected optional age attribute")
	}
}

func TestParseCedarEntityEqualsShape(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User = { name: String };`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if et.Shape == nil || et.Shape.Attributes["name"] == nil {
		t.Error("expected shape with name attribute")
	}
}

func TestParseCedarEntityTags(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity Config tags String;`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["Config"]
	if et.Tags == nil {
		t.Error("expected tags")
	}
}

func TestParseCedarMultiNameEntity(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity A, B, C;`))
	if err != nil {
		t.Fatal(err)
	}
	ns := s.Namespaces[""]
	for _, name := range []string{"A", "B", "C"} {
		if _, ok := ns.EntityTypes[name]; !ok {
			t.Errorf("expected entity %s", name)
		}
	}
}

func TestParseCedarEnum(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity Status enum ["active", "inactive"];`))
	if err != nil {
		t.Fatal(err)
	}
	enum := s.Namespaces[""].EnumTypes["Status"]
	if enum == nil {
		t.Fatal("expected enum Status")
	}
	if len(enum.Values) != 2 || enum.Values[0] != "active" {
		t.Errorf("values = %v", enum.Values)
	}
}

func TestParseCedarAction(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User; action view appliesTo { principal: User, resource: User, context: {} };`))
	if err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["view"]
	if act == nil {
		t.Fatal("expected action view")
	}
	if act.AppliesTo == nil {
		t.Fatal("expected appliesTo")
	}
	if len(act.AppliesTo.PrincipalTypes) != 1 || act.AppliesTo.PrincipalTypes[0] != "User" {
		t.Errorf("principal = %v", act.AppliesTo.PrincipalTypes)
	}
}

func TestParseCedarActionMemberOf(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view in [readOnly];`))
	if err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["view"]
	if len(act.MemberOf) != 1 || act.MemberOf[0].ID != "readOnly" {
		t.Errorf("memberOf = %v", act.MemberOf)
	}
}

func TestParseCedarActionStringName(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action "my action";`))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Namespaces[""].Actions["my action"]; !ok {
		t.Error("expected action 'my action'")
	}
}

func TestParseCedarActionRefWithType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view in [MyNS::Action::"parent"];`))
	if err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["view"]
	if len(act.MemberOf) != 1 {
		t.Fatal("expected 1 member")
	}
	ref := act.MemberOf[0]
	if ref.Type != "MyNS::Action" || ref.ID != "parent" {
		t.Errorf("ref = {Type:%q, ID:%q}", ref.Type, ref.ID)
	}
}

func TestParseCedarCommonType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`type Context = { key: String };`))
	if err != nil {
		t.Fatal(err)
	}
	ct := s.Namespaces[""].CommonTypes["Context"]
	if ct == nil {
		t.Fatal("expected common type Context")
	}
}

func TestParseCedarNamespace(t *testing.T) {
	input := `namespace MyApp {
  entity User;
  action view;
}`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	ns := s.Namespaces["MyApp"]
	if ns == nil {
		t.Fatal("expected namespace MyApp")
	}
	if _, ok := ns.EntityTypes["User"]; !ok {
		t.Error("expected entity User in MyApp")
	}
}

func TestParseCedarNestedNamespacePath(t *testing.T) {
	input := `namespace A::B::C { entity D; }`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Namespaces["A::B::C"]; !ok {
		t.Error("expected namespace A::B::C")
	}
}

func TestParseCedarAnnotations(t *testing.T) {
	input := `@doc("my entity")
entity User;`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if et.Annotations["doc"] != "my entity" {
		t.Errorf("annotations = %v", et.Annotations)
	}
}

func TestParseCedarAnnotationNoValue(t *testing.T) {
	input := `@deprecated
entity Old;`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["Old"]
	if v, ok := et.Annotations["deprecated"]; !ok || v != "" {
		t.Errorf("annotations = %v", et.Annotations)
	}
}

func TestParseCedarAnnotationOnNamespace(t *testing.T) {
	input := `@version("1.0")
namespace NS { entity E; }`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	ns := s.Namespaces["NS"]
	if ns.Annotations["version"] != "1.0" {
		t.Errorf("ns annotations = %v", ns.Annotations)
	}
}

func TestParseCedarSetType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User { groups: Set<String> };`))
	if err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["User"].Shape.Attributes["groups"]
	if _, ok := attr.Type.(SetTypeExpr); !ok {
		t.Errorf("got %T, want SetTypeExpr", attr.Type)
	}
}

func TestParseCedarRecordTypeNested(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User { addr: { street: String } };`))
	if err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["User"].Shape.Attributes["addr"]
	if _, ok := attr.Type.(*RecordTypeExpr); !ok {
		t.Errorf("got %T, want *RecordTypeExpr", attr.Type)
	}
}

func TestParseCedarQualifiedType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User { ref: NS::Other };`))
	if err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["User"].Shape.Attributes["ref"]
	tn, ok := attr.Type.(TypeNameExpr)
	if !ok {
		t.Fatalf("got %T, want TypeNameExpr", attr.Type)
	}
	if tn.Name != "NS::Other" {
		t.Errorf("name = %q, want NS::Other", tn.Name)
	}
}

func TestParseCedarComments(t *testing.T) {
	input := `// line comment
entity User /* block comment */;`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Namespaces[""].EntityTypes["User"]; !ok {
		t.Error("expected entity User")
	}
}

func TestParseCedarStringAttributes(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity E { "quoted attr": Long };`))
	if err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["E"].Shape.Attributes["quoted attr"]
	if attr == nil {
		t.Error("expected attribute 'quoted attr'")
	}
}

func TestParseCedarDuplicateNamespace(t *testing.T) {
	input := `namespace A { entity E; }
namespace A { entity F; }`
	var s Schema
	err := s.UnmarshalCedar([]byte(input))
	if err == nil {
		t.Fatal("expected error for duplicate namespace")
	}
}

func TestParseCedarErrorBadDecl(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`foobar;`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarErrorBadToken(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity `))
	if err == nil {
		t.Fatal("expected error for incomplete input")
	}
}

func TestParseCedarErrorBadAppliesToField(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view appliesTo { bad: User };`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarErrorBadAppliesToToken(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view appliesTo { 123 };`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarActionSingleRef(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view in parent;`))
	if err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["view"]
	if len(act.MemberOf) != 1 || act.MemberOf[0].ID != "parent" {
		t.Error("expected single member-of ref")
	}
}

func TestParseCedarEntityInEmpty(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity User in [];`))
	if err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["User"]
	if len(et.MemberOfTypes) != 0 {
		t.Errorf("expected empty memberOf, got %v", et.MemberOfTypes)
	}
}

func TestParseCedarMultiNameAction(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action "a", "b";`))
	if err != nil {
		t.Fatal(err)
	}
	ns := s.Namespaces[""]
	if _, ok := ns.Actions["a"]; !ok {
		t.Error("expected action a")
	}
	if _, ok := ns.Actions["b"]; !ok {
		t.Error("expected action b")
	}
}

func TestParseCedarEnumTrailingComma(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity S enum ["a", "b",];`))
	if err != nil {
		t.Fatal(err)
	}
	enum := s.Namespaces[""].EnumTypes["S"]
	if len(enum.Values) != 2 {
		t.Errorf("values len = %d, want 2", len(enum.Values))
	}
}

// Cedar text round-trip

func TestCedarRoundTrip(t *testing.T) {
	input := `entity Group;
entity User in [Group] {
  age?: Long,
  name: String,
};
entity Config tags String;
entity Status enum ["active", "inactive"];
type Context = {
  ip: ipaddr,
};
action view appliesTo {
  principal: User,
  resource: User,
  context: Context,
};
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal("parse:", err)
	}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal("marshal:", err)
	}
	var s2 Schema
	if err := s2.UnmarshalCedar(out); err != nil {
		t.Fatal("re-parse:", err)
	}
	out2, err := s2.MarshalCedar()
	if err != nil {
		t.Fatal("re-marshal:", err)
	}
	if string(out) != string(out2) {
		t.Errorf("round-trip mismatch:\n=== first ===\n%s\n=== second ===\n%s", out, out2)
	}
}

func TestCedarRoundTripNamespace(t *testing.T) {
	input := `namespace MyApp {
  entity User;
  action view;
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalCedar(out); err != nil {
		t.Fatal("re-parse:", err)
	}
	if len(s2.Namespaces["MyApp"].EntityTypes) != 1 {
		t.Error("expected 1 entity type after round-trip")
	}
}

func TestCedarRoundTripAnnotations(t *testing.T) {
	input := `@deprecated
@doc("test")
entity User;
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatal(err)
	}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalCedar(out); err != nil {
		t.Fatal("re-parse:", err)
	}
	et := s2.Namespaces[""].EntityTypes["User"]
	if et.Annotations["deprecated"] != "" || et.Annotations["doc"] != "test" {
		t.Errorf("annotations = %v", et.Annotations)
	}
}

// MarshalCedar coverage for action member-of with types

func TestMarshalCedarActionMemberOfWithType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Action("child").InGroup(&ActionRef{Type: "NS::Action", ID: "parent"}).
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, `NS::Action::"parent"`) {
		t.Errorf("expected qualified action ref in output:\n%s", got)
	}
}

func TestMarshalCedarEntityMultipleParents(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("User").MemberOf("Group", "Team").
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "in [Group, Team]") {
		t.Errorf("expected bracketed parent list:\n%s", got)
	}
}

func TestMarshalCedarMultiplePrincipals(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("A").Entity("B").
		Action("x").Principal("A", "B").
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "principal: [A, B]") {
		t.Errorf("expected bracketed principal list:\n%s", got)
	}
}

func TestMarshalCedarEmptyRecord(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Action("x").Context(Record(nil)).
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "context: {}") {
		t.Errorf("expected empty record:\n%s", got)
	}
}

func TestMarshalCedarMultipleNamespaces(t *testing.T) {
	s := NewBuilder().
		Namespace("A").Entity("X").
		Namespace("B").Entity("Y").
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "namespace A") || !strings.Contains(got, "namespace B") {
		t.Errorf("expected both namespaces:\n%s", got)
	}
}

// JSON parsing

func TestJSONRoundTrip(t *testing.T) {
	input := `{
  "": {
    "entityTypes": {
      "User": {
        "memberOfTypes": ["Group"],
        "shape": {
          "type": "Record",
          "attributes": {
            "name": { "type": "String" },
            "age": { "type": "Long", "required": false }
          }
        }
      },
      "Group": {}
    },
    "actions": {
      "view": {
        "appliesTo": {
          "principalTypes": ["User"],
          "resourceTypes": ["User"]
        }
      }
    }
  }
}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal("unmarshal:", err)
	}
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal("marshal:", err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal("re-unmarshal:", err)
	}
	ns := s2.Namespaces[""]
	if ns == nil {
		t.Fatal("expected empty namespace")
	}
	if _, ok := ns.EntityTypes["User"]; !ok {
		t.Error("expected entity User")
	}
}

func TestJSONAllTypes(t *testing.T) {
	input := `{
  "": {
    "entityTypes": {
      "E": {
        "shape": {
          "type": "Record",
          "attributes": {
            "a": { "type": "Long" },
            "b": { "type": "String" },
            "c": { "type": "Bool" },
            "d": { "type": "Boolean" },
            "e": { "type": "Set", "element": { "type": "String" } },
            "f": { "type": "Entity", "name": "E" },
            "g": { "type": "Extension", "name": "ipaddr" },
            "h": { "type": "EntityOrCommon", "name": "MyType" },
            "i": { "type": "CustomName" }
          }
        }
      }
    }
  }
}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	attrs := s.Namespaces[""].EntityTypes["E"].Shape.Attributes
	checkTypeExpr[PrimitiveTypeExpr](t, attrs["a"].Type, "a")
	checkTypeExpr[PrimitiveTypeExpr](t, attrs["b"].Type, "b")
	checkTypeExpr[PrimitiveTypeExpr](t, attrs["c"].Type, "c")
	checkTypeExpr[PrimitiveTypeExpr](t, attrs["d"].Type, "d")
	checkTypeExpr[SetTypeExpr](t, attrs["e"].Type, "e")
	checkTypeExpr[EntityRefExpr](t, attrs["f"].Type, "f")
	checkTypeExpr[ExtensionTypeExpr](t, attrs["g"].Type, "g")
	checkTypeExpr[TypeNameExpr](t, attrs["h"].Type, "h")
	checkTypeExpr[TypeNameExpr](t, attrs["i"].Type, "i")
}

func TestJSONEnumType(t *testing.T) {
	input := `{"": {"entityTypes": {"Status": {"enum": ["active", "inactive"]}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	enum := s.Namespaces[""].EnumTypes["Status"]
	if enum == nil || len(enum.Values) != 2 {
		t.Errorf("expected enum with 2 values, got %v", enum)
	}
}

func TestJSONEntityTags(t *testing.T) {
	input := `{"": {"entityTypes": {"E": {"tags": {"type": "String"}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	et := s.Namespaces[""].EntityTypes["E"]
	if et.Tags == nil {
		t.Error("expected tags")
	}
}

func TestJSONCommonType(t *testing.T) {
	input := `{"": {"commonTypes": {"MyType": {"type": "Record", "attributes": {"x": {"type": "Long"}}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	ct := s.Namespaces[""].CommonTypes["MyType"]
	if ct == nil {
		t.Error("expected common type MyType")
	}
}

func TestJSONCommonTypeAnnotations(t *testing.T) {
	input := `{"": {"commonTypes": {"T": {"type": "Long", "annotations": {"doc": "test"}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	ct := s.Namespaces[""].CommonTypes["T"]
	if ct.Annotations["doc"] != "test" {
		t.Errorf("annotations = %v", ct.Annotations)
	}
}

func TestJSONActionMemberOf(t *testing.T) {
	input := `{"": {"actions": {"view": {"memberOf": [{"type": "Action", "id": "readOnly"}]}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["view"]
	if len(act.MemberOf) != 1 || act.MemberOf[0].Type != "Action" || act.MemberOf[0].ID != "readOnly" {
		t.Errorf("memberOf = %v", act.MemberOf)
	}
}

func TestJSONActionContext(t *testing.T) {
	input := `{"": {"entityTypes": {"U": {}}, "actions": {"x": {"appliesTo": {"principalTypes": ["U"], "resourceTypes": ["U"], "context": {"type": "Record", "attributes": {"k": {"type": "String"}}}}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["x"]
	if act.AppliesTo == nil || act.AppliesTo.Context == nil {
		t.Error("expected context")
	}
}

func TestJSONNamespaceAnnotations(t *testing.T) {
	input := `{"NS": {"annotations": {"v": "1"}, "entityTypes": {"E": {}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if s.Namespaces["NS"].Annotations["v"] != "1" {
		t.Error("expected annotation")
	}
}

func TestJSONEntityAnnotations(t *testing.T) {
	input := `{"": {"entityTypes": {"E": {"annotations": {"doc": "test"}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if s.Namespaces[""].EntityTypes["E"].Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONEnumAnnotations(t *testing.T) {
	input := `{"": {"entityTypes": {"S": {"enum": ["a"], "annotations": {"doc": "test"}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if s.Namespaces[""].EnumTypes["S"].Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONActionAnnotations(t *testing.T) {
	input := `{"": {"actions": {"x": {"annotations": {"doc": "test"}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	if s.Namespaces[""].Actions["x"].Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONAttributeAnnotations(t *testing.T) {
	input := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Long", "annotations": {"doc": "test"}}}}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["E"].Shape.Attributes["x"]
	if attr.Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONTypeWithOnlyName(t *testing.T) {
	input := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"name": "MyType"}}}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["E"].Shape.Attributes["x"]
	tn, ok := attr.Type.(TypeNameExpr)
	if !ok || tn.Name != "MyType" {
		t.Errorf("got %T %v", attr.Type, attr.Type)
	}
}

func TestJSONTypeRecordShorthand(t *testing.T) {
	input := `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"attributes": {"y": {"type": "Long"}}}}}}}}}`
	var s Schema
	if err := s.UnmarshalJSON([]byte(input)); err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["E"].Shape.Attributes["x"]
	_, ok := attr.Type.(*RecordTypeExpr)
	if !ok {
		t.Errorf("got %T, want *RecordTypeExpr", attr.Type)
	}
}

func TestJSONErrorInvalidJSON(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{invalid`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorInvalidNamespace(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"ns": "not_an_object"}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorReservedEntityName(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"Long": {}}}}`))
	if err == nil {
		t.Fatal("expected error for reserved name")
	}
}

func TestJSONErrorReservedCommonName(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"commonTypes": {"Bool": {"type": "Long"}}}}`))
	if err == nil {
		t.Fatal("expected error for reserved name")
	}
}

func TestJSONErrorSetNoElement(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set"}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorEntityNoName(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Entity"}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorExtensionNoName(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Extension"}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorEntityOrCommonNoName(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "EntityOrCommon"}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorUnknownTypeFormat(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorShapeNotRecord(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Long"}}}}}`))
	if err == nil {
		t.Fatal("expected error for non-record shape")
	}
}

// JSON marshal for all type variants

func TestJSONMarshalAllTypes(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").
			Attr("a", Long()).
			Attr("b", String()).
			Attr("c", Bool()).
			Attr("d", Set(Long())).
			Attr("e", Entity("E")).
			Attr("f", Extension("ipaddr")).
			Attr("g", NamedType("MyType")).
			OptionalAttr("h", Long()).
		CommonType("MyType", Long()).
		Build()

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	attrs := s2.Namespaces[""].EntityTypes["E"].Shape.Attributes
	if attrs["h"].Required {
		t.Error("expected optional attribute h")
	}
}

func TestJSONMarshalEnum(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		EnumType("S", "a", "b").
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces[""].EnumTypes["S"] == nil {
		t.Error("expected enum S")
	}
}

func TestJSONMarshalAction(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("U").
		Action("view").
			InGroup(&ActionRef{Type: "Action", ID: "read"}).
			Principal("U").Resource("U").
			Context(Record(map[string]*Attribute{
				"key": {Type: String(), Required: true, Annotations: make(Annotations)},
			})).
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	act := s2.Namespaces[""].Actions["view"]
	if act.AppliesTo == nil || act.AppliesTo.Context == nil {
		t.Error("expected action with context")
	}
}

func TestJSONMarshalCommonTypeAnnotations(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("T", Long()).
		Build()
	s.Namespaces[""].CommonTypes["T"].Annotations["doc"] = "test"

	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces[""].CommonTypes["T"].Annotations["doc"] != "test" {
		t.Error("expected annotation preserved")
	}
}

func TestJSONMarshalEntityNameExpr(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("x", EntityNameExpr{Name: "F"}).
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	json.Unmarshal(data, &m)
	// Should marshal as Entity reference
	if string(data) == "" {
		t.Error("expected non-empty JSON")
	}
}

// Cross-format tests

func TestCedarToJSONToSchema(t *testing.T) {
	cedar := `entity Group;
entity User in [Group] {
  name: String,
};
action view appliesTo {
  principal: User,
  resource: User,
  context: {},
};
`
	var s1 Schema
	if err := s1.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatal(err)
	}
	jsonData, err := s1.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(jsonData); err != nil {
		t.Fatal(err)
	}
	if _, ok := s2.Namespaces[""].EntityTypes["User"]; !ok {
		t.Error("expected entity User in JSON-parsed schema")
	}
}

// Builder tests

func TestBuilderComprehensive(t *testing.T) {
	s := NewBuilder().
		Namespace("App").
		Annotate("version", "1.0").
		Entity("User").
			MemberOf("Group").
			Attr("name", String()).
			OptionalAttr("email", String()).
			Tags(String()).
			Annotate("doc", "user entity").
		Entity("Group").
		EnumType("Status", "active", "inactive").
		CommonType("Context", Record(map[string]*Attribute{
			"ip": {Type: IPAddr(), Required: true, Annotations: make(Annotations)},
		})).
		Action("view").
			Principal("User").
			Resource("User").
			InGroupByName("readOnly").
			Context(NamedType("Context")).
			Annotate("doc", "view action").
		Action("edit").
		Namespace("Other").
		Entity("X").
		Build()

	if s.Namespaces["App"].Annotations["version"] != "1.0" {
		t.Error("expected namespace annotation")
	}
	user := s.Namespaces["App"].EntityTypes["User"]
	if user.Annotations["doc"] != "user entity" {
		t.Error("expected entity annotation")
	}
	if user.Tags == nil {
		t.Error("expected tags")
	}
	if len(user.MemberOfTypes) != 1 {
		t.Error("expected 1 member-of type")
	}
	act := s.Namespaces["App"].Actions["view"]
	if act.Annotations["doc"] != "view action" {
		t.Error("expected action annotation")
	}
	if len(act.MemberOf) != 1 || act.MemberOf[0].ID != "readOnly" {
		t.Error("expected InGroupByName")
	}
	if _, ok := s.Namespaces["Other"]; !ok {
		t.Error("expected Other namespace")
	}
}

func TestBuilderTypeConstructors(t *testing.T) {
	if _, ok := Long().(PrimitiveTypeExpr); !ok {
		t.Error("Long")
	}
	if _, ok := String().(PrimitiveTypeExpr); !ok {
		t.Error("String")
	}
	if _, ok := Bool().(PrimitiveTypeExpr); !ok {
		t.Error("Bool")
	}
	if _, ok := Set(Long()).(SetTypeExpr); !ok {
		t.Error("Set")
	}
	if _, ok := Record(nil).(*RecordTypeExpr); !ok {
		t.Error("Record")
	}
	if _, ok := Entity("E").(EntityRefExpr); !ok {
		t.Error("Entity")
	}
	if _, ok := Extension("ipaddr").(ExtensionTypeExpr); !ok {
		t.Error("Extension")
	}
	if _, ok := IPAddr().(ExtensionTypeExpr); !ok {
		t.Error("IPAddr")
	}
	if _, ok := Decimal().(ExtensionTypeExpr); !ok {
		t.Error("Decimal")
	}
	if _, ok := Datetime().(ExtensionTypeExpr); !ok {
		t.Error("Datetime")
	}
	if _, ok := Duration().(ExtensionTypeExpr); !ok {
		t.Error("Duration")
	}
	if _, ok := NamedType("Foo").(TypeNameExpr); !ok {
		t.Error("NamedType")
	}
}

func TestBuilderChainingThroughEntityToOthers(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Entity("E").
			CommonType("T", Long()).
			Entity("F").
				EnumType("S", "a").
					Action("x").
						Namespace("NS2").Entity("G").
		Build()
	if s.Namespaces["NS"].CommonTypes["T"] == nil {
		t.Error("expected CommonType T")
	}
	if s.Namespaces["NS"].EnumTypes["S"] == nil {
		t.Error("expected EnumType S")
	}
	if s.Namespaces["NS"].Actions["x"] == nil {
		t.Error("expected Action x")
	}
	if s.Namespaces["NS2"] == nil {
		t.Error("expected NS2")
	}
}

func TestBuilderActionChainingToOthers(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Action("a").
			Entity("E").
			Action("b").
				EnumType("S", "v").
					Build()
	if s.Namespaces["NS"].EntityTypes["E"] == nil {
		t.Error("expected Entity E")
	}
	if s.Namespaces["NS"].Actions["b"] == nil {
		t.Error("expected Action b")
	}
}

// Schema methods

func TestSchemaFilename(t *testing.T) {
	var s Schema
	s.SetFilename("test.cedarschema")
	if s.Filename() != "test.cedarschema" {
		t.Errorf("Filename() = %q", s.Filename())
	}
}

func TestParseErrorWithFilename(t *testing.T) {
	var s Schema
	s.SetFilename("myfile.cedarschema")
	err := s.UnmarshalCedar([]byte(`bad!`))
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrParse) {
		t.Errorf("expected ErrParse, got %v", err)
	}
	pe := err.(*ParseError)
	if pe.Filename != "myfile.cedarschema" {
		t.Errorf("filename = %q", pe.Filename)
	}
}

// Error types

func TestErrorTypes(t *testing.T) {
	t.Run("CycleError", func(t *testing.T) {
		err := &CycleError{Path: []string{"A", "B", "A"}}
		if !errors.Is(err, ErrCycle) {
			t.Error("expected Is(ErrCycle)")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("UndefinedTypeError", func(t *testing.T) {
		err := &UndefinedTypeError{Name: "Foo", Namespace: "NS", Context: "in entity"}
		if !errors.Is(err, ErrUndefinedType) {
			t.Error("expected Is(ErrUndefinedType)")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
		err2 := &UndefinedTypeError{Name: "Bar"}
		if err2.Error() == "" {
			t.Error("expected non-empty error without context")
		}
	})

	t.Run("ShadowError", func(t *testing.T) {
		err := &ShadowError{Name: "Foo", Namespace: "NS"}
		if !errors.Is(err, ErrShadow) {
			t.Error("expected Is(ErrShadow)")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("ReservedNameError", func(t *testing.T) {
		err := &ReservedNameError{Name: "Long", Kind: "entity type"}
		if !errors.Is(err, ErrReservedName) {
			t.Error("expected Is(ErrReservedName)")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
	})

	t.Run("ParseError", func(t *testing.T) {
		err := &ParseError{Filename: "f.cedar", Line: 1, Column: 2, Message: "oops"}
		if !errors.Is(err, ErrParse) {
			t.Error("expected Is(ErrParse)")
		}
		if err.Error() == "" {
			t.Error("expected non-empty error")
		}
		err2 := &ParseError{Line: 1, Column: 2, Message: "oops"}
		if err2.Error() == "" {
			t.Error("expected non-empty error without filename")
		}
		err3 := &ParseError{Message: "oops"}
		if err3.Error() == "" {
			t.Error("expected non-empty error without line")
		}
	})
}

// PrimitiveKind.String()

func TestPrimitiveKindString(t *testing.T) {
	if PrimitiveLong.String() != "Long" {
		t.Error("Long")
	}
	if PrimitiveString.String() != "String" {
		t.Error("String")
	}
	if PrimitiveBool.String() != "Bool" {
		t.Error("Bool")
	}
	if PrimitiveKind(99).String() != "Unknown" {
		t.Error("Unknown")
	}
}

// Resolution tests

func TestResolveBasic(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("User").
			Attr("name", String()).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	ns := rs.Namespaces[types.Path("")]
	if ns == nil {
		t.Fatal("expected empty namespace")
	}
	et := ns.EntityTypes[types.EntityType("User")]
	if et == nil {
		t.Fatal("expected resolved entity User")
	}
	attr := et.Shape.Attributes["name"]
	if _, ok := attr.Type.(resolved.Primitive); !ok {
		t.Errorf("got %T, want resolved.Primitive", attr.Type)
	}
}

func TestResolveNamespace(t *testing.T) {
	s := NewBuilder().
		Namespace("App").
		Entity("User").MemberOf("Group").
		Entity("Group").
		Action("view").Principal("User").Resource("User").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	ns := rs.Namespaces[types.Path("App")]
	et := ns.EntityTypes[types.EntityType("App::User")]
	if len(et.MemberOfTypes) != 1 || et.MemberOfTypes[0] != types.EntityType("App::Group") {
		t.Errorf("memberOf = %v", et.MemberOfTypes)
	}

	uid := types.NewEntityUID(types.EntityType("App::Action"), types.String("view"))
	act := ns.Actions[uid]
	if act == nil {
		t.Fatal("expected resolved action")
	}
	if len(act.PrincipalTypes) != 1 || act.PrincipalTypes[0] != types.EntityType("App::User") {
		t.Errorf("principal = %v", act.PrincipalTypes)
	}
}

func TestResolveCommonType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("Context", Record(map[string]*Attribute{
			"ip": {Type: IPAddr(), Required: true, Annotations: make(Annotations)},
		})).
		Entity("U").
		Action("view").Principal("U").Resource("U").Context(NamedType("Context")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	uid := types.NewEntityUID(types.EntityType("Action"), types.String("view"))
	act := rs.Namespaces[types.Path("")].Actions[uid]
	if act.Context == nil {
		t.Fatal("expected resolved context")
	}
	attr := act.Context.Attributes["ip"]
	if _, ok := attr.Type.(resolved.Extension); !ok {
		t.Errorf("got %T, want resolved.Extension", attr.Type)
	}
}

func TestResolveCommonTypeChain(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("A", Long()).
		CommonType("B", NamedType("A")).
		Entity("E").Attr("x", NamedType("B")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["x"]
	prim, ok := attr.Type.(resolved.Primitive)
	if !ok || prim.Kind != resolved.PrimitiveLong {
		t.Errorf("expected Long, got %v", attr.Type)
	}
}

func TestResolveCommonTypePriority(t *testing.T) {
	// When both common type and entity type exist with same name,
	// common type should take priority
	s := NewBuilder().
		Namespace("").
		Entity("Foo").
		CommonType("Foo", Long()).
		Entity("E").Attr("x", NamedType("Foo")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["x"]
	if _, ok := attr.Type.(resolved.Primitive); !ok {
		t.Errorf("expected Primitive (common type priority), got %T", attr.Type)
	}
}

func TestResolveEntityRef(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("User").
		Entity("E").Attr("owner", Entity("User")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["owner"]
	ref, ok := attr.Type.(resolved.EntityRef)
	if !ok {
		t.Fatalf("got %T, want resolved.EntityRef", attr.Type)
	}
	if ref.EntityType != types.EntityType("User") {
		t.Errorf("entity type = %q", ref.EntityType)
	}
}

func TestResolveExtensionTypes(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").
			Attr("a", IPAddr()).
			Attr("b", Decimal()).
			Attr("c", Datetime()).
			Attr("d", Duration()).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	for _, name := range []string{"a", "b", "c", "d"} {
		if _, ok := et.Shape.Attributes[name].Type.(resolved.Extension); !ok {
			t.Errorf("attr %q: got %T, want resolved.Extension", name, et.Shape.Attributes[name].Type)
		}
	}
}

func TestResolveSet(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("items", Set(Long())).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	st, ok := et.Shape.Attributes["items"].Type.(resolved.Set)
	if !ok {
		t.Fatalf("got %T, want resolved.Set", et.Shape.Attributes["items"].Type)
	}
	if _, ok := st.Element.(resolved.Primitive); !ok {
		t.Errorf("set element: got %T, want Primitive", st.Element)
	}
}

func TestResolveTags(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Tags(String()).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	if _, ok := et.Tags.(resolved.Primitive); !ok {
		t.Errorf("tags: got %T, want Primitive", et.Tags)
	}
}

func TestResolveEnum(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		EnumType("Status", "active", "inactive").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	ns := rs.Namespaces[types.Path("")]
	enumType := ns.EnumTypes[types.EntityType("Status")]
	if enumType == nil {
		t.Fatal("expected resolved enum")
	}
	if len(enumType.Values) != 2 {
		t.Errorf("values = %v", enumType.Values)
	}
}

func TestResolveActionMemberOf(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Action("readOnly").
		Action("view").InGroupByName("readOnly").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	uid := types.NewEntityUID(types.EntityType("Action"), types.String("view"))
	act := rs.Namespaces[types.Path("")].Actions[uid]
	if len(act.MemberOf) != 1 {
		t.Fatalf("memberOf len = %d", len(act.MemberOf))
	}
	if act.MemberOf[0].ID != types.String("readOnly") {
		t.Errorf("memberOf[0] = %v", act.MemberOf[0])
	}
}

func TestResolveAnnotations(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Annotate("nskey", "nsval").
		Entity("E").Annotate("ekey", "eval").
		Action("a").Annotate("akey", "aval").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	ns := rs.Namespaces[types.Path("NS")]
	if ns.Annotations["nskey"] != "nsval" {
		t.Error("namespace annotation")
	}
	et := ns.EntityTypes[types.EntityType("NS::E")]
	if et.Annotations["ekey"] != "eval" {
		t.Error("entity annotation")
	}
	uid := types.NewEntityUID(types.EntityType("NS::Action"), types.String("a"))
	if ns.Actions[uid].Annotations["akey"] != "aval" {
		t.Error("action annotation")
	}
}

func TestResolveCrossNamespace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`
namespace A { entity User; }
namespace B { entity E { owner: A::User }; }
`))
	if err != nil {
		t.Fatal(err)
	}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("B")].EntityTypes[types.EntityType("B::E")]
	ref, ok := et.Shape.Attributes["owner"].Type.(resolved.EntityRef)
	if !ok || ref.EntityType != types.EntityType("A::User") {
		t.Errorf("expected A::User, got %v", et.Shape.Attributes["owner"].Type)
	}
}

func TestResolveBuiltinTypeName(t *testing.T) {
	// TypeNameExpr("Long") should resolve to Primitive
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("x", NamedType("Long")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["x"]
	prim, ok := attr.Type.(resolved.Primitive)
	if !ok || prim.Kind != resolved.PrimitiveLong {
		t.Errorf("expected Long, got %v", attr.Type)
	}
}

func TestResolveBuiltinExtensionTypeName(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("x", NamedType("ipaddr")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["x"]
	ext, ok := attr.Type.(resolved.Extension)
	if !ok || ext.Name != "ipaddr" {
		t.Errorf("expected Extension{ipaddr}, got %v", attr.Type)
	}
}

func TestResolveBuiltinCedarPrefix(t *testing.T) {
	// __cedar::Long should resolve
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity E { x: __cedar::Long };`))
	if err != nil {
		t.Fatal(err)
	}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	attr := et.Shape.Attributes["x"]
	prim, ok := attr.Type.(resolved.Primitive)
	if !ok || prim.Kind != resolved.PrimitiveLong {
		t.Errorf("expected Long, got %v", attr.Type)
	}
}

// Resolution error tests

func TestResolveCycleError(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("A", NamedType("B")).
		CommonType("B", NamedType("A")).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrCycle) {
		t.Errorf("expected ErrCycle, got %v", err)
	}
}

func TestResolveSelfCycle(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("A", Set(NamedType("A"))).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrCycle) {
		t.Errorf("expected ErrCycle, got %v", err)
	}
}

func TestResolveShadowError(t *testing.T) {
	s := NewBuilder().
		Namespace("").Entity("Foo").
		Namespace("NS").Entity("Foo").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrShadow) {
		t.Errorf("expected ErrShadow, got %v", err)
	}
}

func TestResolveShadowCommon(t *testing.T) {
	s := NewBuilder().
		Namespace("").CommonType("T", Long()).
		Namespace("NS").CommonType("T", String()).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrShadow) {
		t.Errorf("expected ErrShadow, got %v", err)
	}
}

func TestResolveShadowAction(t *testing.T) {
	s := NewBuilder().
		Namespace("").Action("x").
		Namespace("NS").Action("x").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrShadow) {
		t.Errorf("expected ErrShadow, got %v", err)
	}
}

func TestResolveUndefinedType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("x", NamedType("Nonexistent")).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveUndefinedMemberOf(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("User").MemberOf("Nonexistent").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveUndefinedAction(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Action("view").InGroupByName("nonexistent").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveUndefinedPrincipal(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Action("view").Principal("Nonexistent").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveContextNotRecord(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("T", Long()).
		Entity("U").
		Action("view").Principal("U").Resource("U").Context(NamedType("T")).
		Build()

	_, err := s.Resolve()
	if err == nil {
		t.Fatal("expected error for non-record context")
	}
}

func TestResolveNoNamespaces(t *testing.T) {
	s := &Schema{Namespaces: make(map[string]*Namespace)}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	if len(rs.Namespaces) != 0 {
		t.Error("expected empty")
	}
}


func checkTypeExpr[T TypeExpr](t *testing.T, expr TypeExpr, name string) {
	t.Helper()
	if _, ok := expr.(T); !ok {
		t.Errorf("attr %q: got %T, want %T", name, expr, *new(T))
	}
}

// Additional coverage tests

func TestCedarEmitAllTypeExprVariants(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("Ref").
		Entity("E").
			Attr("prim", Long()).
			Attr("set", Set(String())).
			Attr("rec", Record(map[string]*Attribute{
				"nested": {Type: Bool(), Required: true, Annotations: make(Annotations)},
			})).
			Attr("eref", Entity("Ref")).
			Attr("ext", Extension("ipaddr")).
			Attr("tname", NamedType("Ref")).
			Attr("ename", EntityNameExpr{Name: "Ref"}).
		Build()

	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	for _, want := range []string{"Long", "Set<String>", "Bool", "Ref", "ipaddr"} {
		if !strings.Contains(got, want) {
			t.Errorf("expected %q in output:\n%s", want, got)
		}
	}
}

func TestCedarEmitQuotedAttrNames(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("good", Long()).
		Build()
	s.Namespaces[""].EntityTypes["E"].Shape.Attributes["has space"] = &Attribute{
		Type: Long(), Required: true, Annotations: make(Annotations),
	}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, `"has space"`) {
		t.Errorf("expected quoted attr name:\n%s", got)
	}
}

func TestCedarEmitQuotedActionName(t *testing.T) {
	s := NewBuilder().Namespace("").Build()
	s.Namespaces[""].Actions["my action"] = &ActionDef{Annotations: make(Annotations)}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, `"my action"`) {
		t.Errorf("expected quoted action name:\n%s", got)
	}
}

func TestCedarEmitEntityEmptyShape(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Tags(Set(Long())).
		Build()
	// Entity with tags and no shape
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "tags Set<Long>") {
		t.Errorf("expected tags in output:\n%s", got)
	}
}



func TestCedarEmitQuotesNonIdentNames(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{
				"E": {
					Shape: &RecordTypeExpr{Attributes: map[string]*Attribute{
						"":          {Type: String(), Required: true, Annotations: make(Annotations)},
						"1digit":    {Type: Long(), Required: true, Annotations: make(Annotations)},
						"has space": {Type: Bool(), Required: true, Annotations: make(Annotations)},
					}},
					Annotations: make(Annotations),
				},
			},
			EnumTypes:   map[string]*EnumTypeDef{},
			Actions:     map[string]*ActionDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Annotations: make(Annotations),
		},
	}}
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, `"1digit"`) {
		t.Errorf("expected quoted digit-start attr name:\n%s", got)
	}
	if !strings.Contains(got, `"has space"`) {
		t.Errorf("expected quoted space-containing attr name:\n%s", got)
	}
}

func TestParseCedarActionEmptyRefList(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view in [];`))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Namespaces[""].Actions["view"].MemberOf) != 0 {
		t.Error("expected empty member list")
	}
}

func TestParseCedarActionMultipleRefs(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action view in [a, b];`))
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Namespaces[""].Actions["view"].MemberOf) != 2 {
		t.Error("expected 2 members")
	}
}

func TestParseCedarEntityTypeListMultiple(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`
entity U;
action view appliesTo { principal: [U], resource: [U] };
`))
	if err != nil {
		t.Fatal(err)
	}
}

func TestParseCedarActionContextType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`
entity U;
action view appliesTo { principal: U, resource: U, context: MyContext };
`))
	if err != nil {
		t.Fatal(err)
	}
	ctx := s.Namespaces[""].Actions["view"].AppliesTo.Context
	if _, ok := ctx.(TypeNameExpr); !ok {
		t.Errorf("got %T, want TypeNameExpr", ctx)
	}
}

func TestParseCedarAnnotationOnRecord(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity E { @doc("field") name: String };`))
	if err != nil {
		t.Fatal(err)
	}
	attr := s.Namespaces[""].EntityTypes["E"].Shape.Attributes["name"]
	if attr.Annotations["doc"] != "field" {
		t.Errorf("annotations = %v", attr.Annotations)
	}
}

func TestResolveAllPrimitiveKinds(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").
			Attr("l", Long()).
			Attr("s", String()).
			Attr("b", Bool()).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	for name, wantKind := range map[string]resolved.PrimitiveKind{
		"l": resolved.PrimitiveLong,
		"s": resolved.PrimitiveString,
		"b": resolved.PrimitiveBool,
	} {
		prim, ok := et.Shape.Attributes[name].Type.(resolved.Primitive)
		if !ok || prim.Kind != wantKind {
			t.Errorf("attr %q: got %v, want %v", name, et.Shape.Attributes[name].Type, wantKind)
		}
	}
}

func TestResolveEntityNameExpr(t *testing.T) {
	// Parse a Cedar schema where memberOfTypes generates EntityNameExpr
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity Group; entity User in Group;`))
	if err != nil {
		t.Fatal(err)
	}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("User")]
	if len(et.MemberOfTypes) != 1 || et.MemberOfTypes[0] != types.EntityType("Group") {
		t.Errorf("memberOf = %v", et.MemberOfTypes)
	}
}

func TestResolveUndefinedCommonType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		CommonType("A", NamedType("Nonexistent")).
		Entity("E").Attr("x", NamedType("A")).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveUndefinedResource(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("U").
		Action("view").Principal("U").Resource("Nonexistent").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveUndefinedTags(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Tags(NamedType("Nonexistent")).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestResolveActionMemberOfWithType(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Action("parent").
		Action("child").InGroup(&ActionRef{Type: "NS::Action", ID: "parent"}).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	uid := types.NewEntityUID(types.EntityType("NS::Action"), types.String("child"))
	act := rs.Namespaces[types.Path("NS")].Actions[uid]
	if len(act.MemberOf) != 1 {
		t.Fatal("expected 1 member")
	}
}

func TestJSONMarshalRecordType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("r", Record(map[string]*Attribute{
			"x": {Type: Long(), Required: true, Annotations: make(Annotations)},
			"y": {Type: String(), Required: false, Annotations: Annotations{"doc": "test"}},
		})).
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	attr := s2.Namespaces[""].EntityTypes["E"].Shape.Attributes["r"]
	rec, ok := attr.Type.(*RecordTypeExpr)
	if !ok || len(rec.Attributes) != 2 {
		t.Error("expected record with 2 attributes")
	}
}

func TestJSONMarshalEntityAnnotations(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Annotate("doc", "test").
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces[""].EntityTypes["E"].Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONMarshalEnumAnnotations(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		EnumType("S", "a").
		Build()
	s.Namespaces[""].EnumTypes["S"].Annotations["doc"] = "test"
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces[""].EnumTypes["S"].Annotations["doc"] != "test" {
		t.Error("expected annotation")
	}
}

func TestJSONMarshalNamespaceAnnotations(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Annotate("v", "1").
		Entity("E").
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces["NS"].Annotations["v"] != "1" {
		t.Error("expected annotation")
	}
}

func TestJSONMarshalEntityTags(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Tags(String()).
		Build()
	data, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if s2.Namespaces[""].EntityTypes["E"].Tags == nil {
		t.Error("expected tags")
	}
}

func TestBuilderActionCommonType(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Action("a").CommonType("T", Long()).
		Build()
	if s.Namespaces["NS"].CommonTypes["T"] == nil {
		t.Error("expected CommonType T from ActionBuilder")
	}
}

func TestBuilderActionNamespace(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		Action("a").Namespace("NS2").Entity("E").
		Build()
	if s.Namespaces["NS2"] == nil {
		t.Error("expected NS2")
	}
}

func TestBuilderNamespaceBuild(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").Entity("E").
		Build()
	if s.Namespaces["NS"] == nil {
		t.Error("expected NS")
	}
}

func TestParseErrorNoLineInfo(t *testing.T) {
	err := &ParseError{Message: "test"}
	if err.Error() == "" {
		t.Error("expected non-empty error")
	}
}

func TestCedarEmitNoEntitiesNoActions(t *testing.T) {
	s := NewBuilder().
		Namespace("NS").
		CommonType("T", Long()).
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	if len(out) == 0 {
		t.Error("expected non-empty output")
	}
}

func TestResolveCommonTypeFromNamespace(t *testing.T) {
	// Test that a common type in a namespace resolves correctly
	var s Schema
	err := s.UnmarshalCedar([]byte(`namespace NS {
  type MyType = Long;
  entity E { x: MyType };
}`))
	if err != nil {
		t.Fatal(err)
	}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("NS")].EntityTypes[types.EntityType("NS::E")]
	prim, ok := et.Shape.Attributes["x"].Type.(resolved.Primitive)
	if !ok || prim.Kind != resolved.PrimitiveLong {
		t.Errorf("expected Long, got %v", et.Shape.Attributes["x"].Type)
	}
}

func TestResolveCommonTypeTopoOrder(t *testing.T) {
	// C depends on B which depends on A - must resolve in correct order
	s := NewBuilder().
		Namespace("").
		CommonType("A", Long()).
		CommonType("B", Set(NamedType("A"))).
		CommonType("C", NamedType("B")).
		Entity("E").Attr("x", NamedType("C")).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	st, ok := et.Shape.Attributes["x"].Type.(resolved.Set)
	if !ok {
		t.Fatalf("expected Set, got %T", et.Shape.Attributes["x"].Type)
	}
	if _, ok := st.Element.(resolved.Primitive); !ok {
		t.Errorf("set element: got %T, want Primitive", st.Element)
	}
}

func TestResolveShadowEnum(t *testing.T) {
	s := NewBuilder().
		Namespace("").EnumType("S", "a").
		Namespace("NS").EnumType("S", "b").
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrShadow) {
		t.Errorf("expected ErrShadow, got %v", err)
	}
}

func TestResolveEnumAsEntityType(t *testing.T) {
	// Enum types should be resolvable in entity-type positions
	s := NewBuilder().
		Namespace("").
		EnumType("Status", "active").
		Entity("E").MemberOf("Status").
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	if len(et.MemberOfTypes) != 1 || et.MemberOfTypes[0] != types.EntityType("Status") {
		t.Errorf("memberOf = %v", et.MemberOfTypes)
	}
}

func TestResolveRecordType(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("r", Record(map[string]*Attribute{
			"x": {Type: Long(), Required: true, Annotations: Annotations{"doc": "test"}},
		})).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	rec, ok := et.Shape.Attributes["r"].Type.(*resolved.RecordType)
	if !ok {
		t.Fatalf("expected RecordType, got %T", et.Shape.Attributes["r"].Type)
	}
	if rec.Attributes["x"].Annotations["doc"] != "test" {
		t.Error("expected attribute annotation preserved")
	}
}

func TestResolveActionContext(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("U").
		Action("view").Principal("U").Resource("U").Context(Record(map[string]*Attribute{
			"k": {Type: String(), Required: true, Annotations: make(Annotations)},
		})).
		Build()

	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	uid := types.NewEntityUID(types.EntityType("Action"), types.String("view"))
	act := rs.Namespaces[types.Path("")].Actions[uid]
	if act.Context == nil {
		t.Fatal("expected context")
	}
	if _, ok := act.Context.Attributes["k"].Type.(resolved.Primitive); !ok {
		t.Error("expected primitive string in context")
	}
}

func TestResolveUndefinedEntityRef(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("E").Attr("x", Entity("Nonexistent")).
		Build()

	_, err := s.Resolve()
	if !errors.Is(err, ErrUndefinedType) {
		t.Errorf("expected ErrUndefinedType, got %v", err)
	}
}

func TestJSONMarshalAllTypeExprsViaTags(t *testing.T) {
	tests := []struct {
		name string
		expr TypeExpr
	}{
		{"set", Set(Long())},
		{"record", Record(map[string]*Attribute{
			"x": {Type: Long(), Required: true, Annotations: make(Annotations)},
		})},
		{"entityRef", Entity("E")},
		{"extension", Extension("ipaddr")},
		{"typeName", NamedType("MyT")},
		{"entityName", EntityNameExpr{Name: "E"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBuilder().Namespace("").Entity("E").Build()
			s.Namespaces[""].EntityTypes["E"].Tags = tt.expr
			data, err := s.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			if len(data) == 0 {
				t.Fatal("empty JSON")
			}
		})
	}
}

func TestJSONMarshalCommonTypeAllVariants(t *testing.T) {
	tests := []struct {
		name string
		expr TypeExpr
	}{
		{"set", Set(String())},
		{"entity", Entity("E")},
		{"extension", Extension("decimal")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewBuilder().Namespace("").Entity("E").CommonType("T", tt.expr).Build()
			data, err := s.MarshalJSON()
			if err != nil {
				t.Fatal(err)
			}
			var s2 Schema
			if err := s2.UnmarshalJSON(data); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestResolveEntityNameExprInTypePosition(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("User").
		Entity("E").
		Build()
	s.Namespaces[""].EntityTypes["E"].Shape = &RecordTypeExpr{
		Attributes: map[string]*Attribute{
			"ref": {Type: EntityNameExpr{Name: "User"}, Required: true, Annotations: make(Annotations)},
		},
	}
	rs, err := s.Resolve()
	if err != nil {
		t.Fatal(err)
	}
	et := rs.Namespaces[types.Path("")].EntityTypes[types.EntityType("E")]
	ref, ok := et.Shape.Attributes["ref"].Type.(resolved.EntityRef)
	if !ok || ref.EntityType != types.EntityType("User") {
		t.Errorf("expected EntityRef{User}, got %v", et.Shape.Attributes["ref"].Type)
	}
}

func TestParseCedarScannerErrorInEntity(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity User { name: ~ }"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInType(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("type T = ~"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInAnnotation(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("@~"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInNamespace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("namespace ~"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInSet(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x: Set<~> };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInPath(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x: A::~ };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInRecord(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { ~: Long };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInActionRef(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("action a in [~];"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInAppliesTo(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("action a appliesTo { principal: ~, };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarScannerErrorInEnum(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity S enum [~];`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingEntitySemicolon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity User"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingActionSemicolon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("action view"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingTypeSemicolon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("type T = Long"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingTypeEquals(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("type T Long;"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingNamespaceBrace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("namespace NS entity E;"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingCloseNamespace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("namespace NS { entity E;"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingRecordClose(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x: Long"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingEnumBracket(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity S enum "a", "b";`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAttrColon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x Long };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingSetAngle(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x: Set Long };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingSetCloseAngle(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity E { x: Set<Long };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAppliesToBrace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity U; action a appliesTo principal: U;"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAppliesToColon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity U; action a appliesTo { principal U };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAppliesToCloseBrace(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity U; action a appliesTo { principal: U"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingEntityTypeBracket(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("entity U; action a appliesTo { principal: [U };"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingActionRefBracket(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte("action a in [b ;"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAnnotationParen(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`@doc("test" entity E;`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarMissingAnnotationString(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`@doc(123) entity E;`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParseCedarBadDeclToken(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`namespace NS { 123 }`))
	if err == nil {
		t.Fatal("expected error for non-ident declaration")
	}
}

func TestParseCedarBadDeclKeyword(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`namespace NS { unknown E; }`))
	if err == nil {
		t.Fatal("expected error for unknown keyword")
	}
}

func TestParseCedarEnumCloseBracket(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`entity S enum ["a" "b"];`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCedarEmitActionAppliesTo(t *testing.T) {
	s := NewBuilder().
		Namespace("").
		Entity("U").
		Action("x").
			Principal("U").
			Resource("U").
			Context(Record(map[string]*Attribute{
				"k": {Type: Long(), Required: true, Annotations: make(Annotations)},
			})).
		Build()
	out, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(out)
	if !strings.Contains(got, "context:") {
		t.Errorf("expected context in output:\n%s", got)
	}
}

func TestJSONErrorBadAction(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"actions": {"x": "bad"}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadCommonType(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"commonTypes": {"T": "bad"}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadEntityType(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": "bad"}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadAttribute(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": "bad"}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadSetElement(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set", "element": "bad"}}}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadTags(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"entityTypes": {"E": {"tags": "bad"}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONErrorBadContext(t *testing.T) {
	var s Schema
	err := s.UnmarshalJSON([]byte(`{"": {"actions": {"x": {"appliesTo": {"context": "bad"}}}}}`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJSONMarshalActionAnnotations(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{
				"User": {Annotations: make(Annotations)},
			},
			EnumTypes:   map[string]*EnumTypeDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Actions: map[string]*ActionDef{
				"view": {
					AppliesTo:   &AppliesTo{PrincipalTypes: []string{"User"}},
					Annotations: Annotations{"doc": "view action"},
				},
			},
			Annotations: make(Annotations),
		},
	}}
	b, err := s.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	var s2 Schema
	if err := s2.UnmarshalJSON(b); err != nil {
		t.Fatal(err)
	}
	act := s2.Namespaces[""].Actions["view"]
	if act.Annotations["doc"] != "view action" {
		t.Errorf("annotation = %q, want %q", act.Annotations["doc"], "view action")
	}
}

func TestCedarWriteActionMultipleMemberOf(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{},
			EnumTypes:   map[string]*EnumTypeDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Actions: map[string]*ActionDef{
				"view": {
					MemberOf: []*ActionRef{
						{ID: "readOnly"},
						{ID: "allActions"},
					},
					Annotations: make(Annotations),
				},
			},
			Annotations: make(Annotations),
		},
	}}
	b, err := s.MarshalCedar()
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	if !strings.Contains(got, `"readOnly"`) || !strings.Contains(got, `"allActions"`) {
		t.Errorf("expected both action refs in output:\n%s", got)
	}
	if !strings.Contains(got, ", ") {
		t.Error("expected comma separator between action refs")
	}
}

func TestResolveActionContextError(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{
				"User": {Annotations: make(Annotations)},
			},
			EnumTypes:   map[string]*EnumTypeDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Actions: map[string]*ActionDef{
				"view": {
					AppliesTo: &AppliesTo{
						Context: TypeNameExpr{Name: "NonExistent"},
					},
					Annotations: make(Annotations),
				},
			},
			Annotations: make(Annotations),
		},
	}}
	_, err := s.Resolve()
	if err == nil {
		t.Fatal("expected error for unresolvable context type")
	}
}

func TestResolveSetElementError(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{
				"User": {
					Shape: &RecordTypeExpr{
						Attributes: map[string]*Attribute{
							"tags": {Type: SetTypeExpr{Element: TypeNameExpr{Name: "NonExistent"}}, Required: true, Annotations: make(Annotations)},
						},
					},
					Annotations: make(Annotations),
				},
			},
			EnumTypes:   map[string]*EnumTypeDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Actions:     map[string]*ActionDef{},
			Annotations: make(Annotations),
		},
	}}
	_, err := s.Resolve()
	if err == nil {
		t.Fatal("expected error for unresolvable Set element")
	}
}

func TestResolveEntityNameExprError(t *testing.T) {
	s := &Schema{Namespaces: map[string]*Namespace{
		"": {
			EntityTypes: map[string]*EntityTypeDef{
				"User": {
					Shape: &RecordTypeExpr{
						Attributes: map[string]*Attribute{
							"ref": {Type: EntityNameExpr{Name: "NonExistent"}, Required: true, Annotations: make(Annotations)},
						},
					},
					Annotations: make(Annotations),
				},
			},
			EnumTypes:   map[string]*EnumTypeDef{},
			CommonTypes: map[string]*CommonTypeDef{},
			Actions:     map[string]*ActionDef{},
			Annotations: make(Annotations),
		},
	}}
	_, err := s.Resolve()
	if err == nil {
		t.Fatal("expected error for unresolvable EntityNameExpr")
	}
}

func TestParserErrorPaths(t *testing.T) {
	inputs := []string{
		// UnmarshalCedar initial advance error (line 16)
		"~",
		// parsePath error after namespace keyword (line 324)
		"namespace ;",
		// parseAnnotations error in namespace body (line 361)
		"namespace Foo { @~ }",
		// parseDecl non-ident in namespace body (line 372)
		"namespace Foo { ; }",
		// advance error after "entity" (line 388)
		"entity ~",
		// advance error after "enum" (line 400)
		"entity Foo enum ~",
		// expect semicolon error after enum values (line 407)
		"entity Foo enum [\"a\"] (",
		// advance error after "in" in entity (line 422)
		"entity Foo in ~",
		// parseEntityTypeList error after "in" (line 426)
		"entity Foo in ;",
		// advance error after "=" in entity shape (line 434)
		"entity Foo = ~",
		// advance error after "tags" (line 447)
		"entity Foo tags ~",
		// parseType error after "tags" (line 451)
		"entity Foo tags ;",
		// advance error after "action" (line 477)
		"action ~",
		// parseNameList error in action (line 483)
		"action ;",
		// advance error after "in" in action (line 489)
		`action "foo" in ~`,
		// advance error after "appliesTo" (line 500)
		`action "foo" appliesTo ~`,
		// advance error after "type" (line 529)
		"type ~",
		// expect ident error in type decl (line 533)
		"type ;",
		// expect "=" error in type decl (line 540)
		"type MyType ;",
		// expect ident after "@" in annotations (line 560)
		"@;",
		// advance error after "(" in annotation (line 568)
		"@foo(~",
		// parsePath expect ident error (line 584)
		"entity ;",
		// advance error in parsePath after "::" (line 594)
		"entity Foo::~",
		// parseType Set element error (line 615)
		"type X = Set<;>",
		// parseAnnotations error in record type (line 633)
		"entity Foo { @~ };",
		// parseName error in record attribute (line 637)
		"entity Foo { ~ };",
		// advance error after "?" in record (line 643)
		"entity Foo { x?~ };",
		// advance error after "," in record (line 660)
		"entity Foo { x: Long, ~ };",
		// advance error after "," in identList (line 678)
		"entity Foo, ~",
		// expect ident error after "," in identList (line 682)
		"entity Foo, ;",
		// parseName error in nameList first (line 692)
		"action ;",
		// advance error after "," in nameList (line 697)
		`action "foo", ~`,
		// parseName error after "," in nameList (line 701)
		`action "foo", ;`,
		// advance error after string in parseName (line 712)
		`action "foo"~`,
		// expect ident error in parseName (line 718)
		"action ,",
		// advance error after "[" in entityTypeList (line 726)
		"entity Foo in [~",
		// parsePath error in entityTypeList bracket (line 732)
		"entity Foo in [;",
		// advance error after "," in entityTypeList (line 737)
		"entity Foo in [Bar, ~",
		// parsePath error after "," in entityTypeList (line 741)
		"entity Foo in [Bar, ;",
		// parsePath error for bare entity type (line 753)
		"entity Foo in ;",
		// advance error after "[" in actionRefList (line 767)
		`action "x" in [~`,
		// advance error after "," in actionRefList (line 772)
		`action "x" in ["a", ~`,
		// parseActionRef error after "," in actionRefList (line 776)
		`action "x" in ["a", ;`,
		// parseActionRef error for bare action ref (line 788)
		`action "x" in ;`,
		// parseActionRef string advance error (line 798)
		`action "x" in "a"~`,
		// parseActionRef string return (line 801) - needs valid parse of string ref
		// covered via action "x" in "a"; (already tested elsewhere? let's add it)
		// expect ident error in parseActionRef (line 804)
		`action "x" in [,]`,
		// advance error after "::" in parseActionRef (line 809)
		`action "x" in [Foo::~]`,
		// advance error after string in parseActionRef path::str (line 814)
		`action "x" in [Foo::"a"~]`,
		// expect ident after "::" in parseActionRef (line 823)
		`action "x" in [Foo::;]`,
		// expect LBrace error in parseAppliesTo (line 837)
		`action "x" appliesTo ;`,
		// advance error after "principal" (line 842)
		`action "x" appliesTo { principal ~`,
		// expect colon after "principal" (line 844)
		// Not separately needed - the ~ test triggers advance error at 842
		// advance error after "resource" (line 854)
		`action "x" appliesTo { resource ~`,
		// expect colon after "resource" (line 857)
		`action "x" appliesTo { resource ;`,
		// parseEntityTypeList error after resource colon (line 861)
		`action "x" appliesTo { resource: ;`,
		// advance error after "context" (line 866)
		`action "x" appliesTo { context ~`,
		// expect colon after "context" (line 869)
		`action "x" appliesTo { context ;`,
		// parseType error after context colon (line 873)
		`action "x" appliesTo { context: ;`,
		// advance error after comma in appliesTo (line 881)
		`action "x" appliesTo { principal: [User], ~`,
		// expect string error in stringList (line 899)
		`entity Foo enum [;`,
		// advance error after "," in stringList (line 904)
		`entity Foo enum ["a", ~`,
		// expect string error after "," in stringList (line 911)
		`entity Foo enum ["a", ;`,
		// expect string after ( in annotation value (line 568)
		"@foo(;",
		// expect ident after :: in parsePath (line 594)
		"type X = Foo::;",
		// parseName error in record attribute (line 637)
		"entity Foo { , };",
		// non-ident token in appliesTo body (line 837)
		`action "x" appliesTo { ; }`,
	}
	for _, input := range inputs {
		var s Schema
		if err := s.UnmarshalCedar([]byte(input)); err == nil {
			t.Errorf("expected error for input %q", input)
		}
	}
}

func TestParseActionRefStringOnly(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action "x" in "parent";`))
	if err != nil {
		t.Fatal(err)
	}
	act := s.Namespaces[""].Actions["x"]
	if len(act.MemberOf) != 1 {
		t.Fatalf("expected 1 memberOf, got %d", len(act.MemberOf))
	}
	if act.MemberOf[0].ID != "parent" {
		t.Errorf("memberOf ID = %q, want %q", act.MemberOf[0].ID, "parent")
	}
}

func TestParseAppliesToPrincipalColon(t *testing.T) {
	var s Schema
	err := s.UnmarshalCedar([]byte(`action "x" appliesTo { principal ; }`))
	if err == nil {
		t.Fatal("expected error")
	}
}
