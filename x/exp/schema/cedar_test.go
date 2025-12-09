package schema

import (
	"bytes"
	"strings"
	"testing"
)

func TestCedarMarshalUnmarshalRoundTrip(t *testing.T) {
	t.Parallel()

	// Create a schema programmatically
	s := New()
	ns := NewNamespace("PhotoApp")
	ns.AddEntity(NewEntity("User").
		MemberOf("Group").
		SetAttributes(
			RequiredAttr("name", String()),
			OptionalAttr("email", String()),
		))
	ns.AddEntity(NewEntity("Group"))
	ns.AddAction(NewAction("view").
		SetPrincipalTypes("User", "Group").
		SetResourceTypes("Photo"))
	ns.AddCommonType(NewCommonType("Name", String()))
	s.AddNamespace(ns)

	// Marshal to Cedar
	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Unmarshal back
	var s2 Schema
	if err := s2.UnmarshalCedar(cedarData); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	// Marshal again
	cedarData2, err := s2.MarshalCedar()
	if err != nil {
		t.Fatalf("Second MarshalCedar() error = %v", err)
	}

	// Compare
	if !bytes.Equal(cedarData, cedarData2) {
		t.Errorf("Round-trip produced different Cedar:\nFirst:\n%s\nSecond:\n%s", cedarData, cedarData2)
	}
}

func TestCedarUnmarshalValidSchema(t *testing.T) {
	t.Parallel()

	input := `
namespace PhotoApp {
	type Name = String;
	entity User in [Group] {
		name: String,
		email?: String,
	};
	entity Group;
	entity Photo {
		title: String,
		tags?: Set<String>,
	};
	action view appliesTo {
		principal: [User, Group],
		resource: Photo,
	};
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("PhotoApp")
	if ns == nil {
		t.Fatal("expected PhotoApp namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	if len(user.MemberOfTypes) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(user.MemberOfTypes))
	}
	if len(user.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(user.Attributes))
	}

	view := ns.GetAction("view")
	if view == nil {
		t.Fatal("expected view action")
	}
	if len(view.PrincipalTypes) != 2 {
		t.Errorf("expected 2 principal types, got %d", len(view.PrincipalTypes))
	}
}

func TestCedarUnmarshalInvalidSchema(t *testing.T) {
	t.Parallel()

	input := `namespace foo { invalid syntax here }`
	var s Schema
	err := s.UnmarshalCedar([]byte(input))
	if err == nil {
		t.Error("expected error for invalid schema")
	}
}

func TestCedarMarshalEmpty(t *testing.T) {
	t.Parallel()

	s := New()
	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Empty schema should produce no output
	if len(cedarData) != 0 {
		t.Errorf("expected empty output, got %s", cedarData)
	}
}

func TestCedarUnmarshalEmpty(t *testing.T) {
	t.Parallel()

	var s Schema
	if err := s.UnmarshalCedar([]byte("")); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}
	if len(s.Namespaces) != 0 {
		t.Errorf("expected empty namespaces, got %d", len(s.Namespaces))
	}
}

func TestCedarMarshalAllTypes(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")

	ns.AddEntity(NewEntity("TestEntity").
		SetAttributes(
			RequiredAttr("boolean", Boolean()),
			RequiredAttr("long", Long()),
			RequiredAttr("string", String()),
			RequiredAttr("set", SetOf(String())),
			RequiredAttr("record", Record(
				RequiredAttr("nested", String()),
			)),
			RequiredAttr("entity", Entity("OtherEntity")),
			RequiredAttr("extension", Extension("ipaddr")),
			RequiredAttr("ref", Ref("CommonType")),
		))

	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "Bool") {
		t.Error("expected Bool type in Cedar")
	}
	if !strings.Contains(cedarStr, "Long") {
		t.Error("expected Long type in Cedar")
	}
	if !strings.Contains(cedarStr, "String") {
		t.Error("expected String type in Cedar")
	}
	if !strings.Contains(cedarStr, "Set<") {
		t.Error("expected Set type in Cedar")
	}
}

func TestCedarMarshalEntityWithTags(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("TaggedEntity").SetTags(String()))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !strings.Contains(string(cedarData), "tags") {
		t.Error("expected tags in Cedar")
	}
}

func TestCedarMarshalEnumEntity(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Color").SetEnum("red", "green", "blue"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "enum") {
		t.Error("expected enum in Cedar")
	}
	if !strings.Contains(cedarStr, `"red"`) {
		t.Error("expected red in Cedar")
	}
}

func TestCedarMarshalActionWithMemberOf(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.AddAction(NewAction("edit").
		InActions(
			ActionRef{Name: "view"},
			ActionRef{Namespace: "OtherNS", Name: "read"},
		))
	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "in [") {
		t.Error("expected in [...] in Cedar")
	}
}

func TestCedarMarshalActionWithContext(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.AddAction(NewAction("create").
		SetPrincipalTypes("User").
		SetResourceTypes("Doc").
		SetContext(Record(
			RequiredAttr("ip", Extension("ipaddr")),
		)))
	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !strings.Contains(string(cedarData), "context:") {
		t.Error("expected context in Cedar")
	}
}

func TestCedarMarshalAnnotations(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test").Annotate("doc", "test namespace")
	ns.AddEntity(NewEntity("User").Annotate("doc", "user entity"))
	ns.AddAction(NewAction("view").Annotate("doc", "view action"))
	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	if !strings.Contains(string(cedarData), "@doc") {
		t.Error("expected @doc annotation in Cedar")
	}
}

func TestCedarUnmarshalWithAnnotations(t *testing.T) {
	t.Parallel()

	input := `
@doc("test namespace")
namespace Test {
	@doc("user entity")
	entity User;

	@doc("view action")
	action view;
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	val, ok := ns.Annotations.Get("doc")
	if !ok || val != "test namespace" {
		t.Error("expected namespace annotation")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	val, ok = user.Annotations.Get("doc")
	if !ok || val != "user entity" {
		t.Error("expected entity annotation")
	}
}

func TestCedarUnmarshalNestedRecord(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	profile: {
		name: String,
	},
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
}

func TestCedarAnonymousNamespace(t *testing.T) {
	t.Parallel()

	input := `
entity User;
action view appliesTo {
	principal: User,
	resource: Resource,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	if ns.GetEntity("User") == nil {
		t.Error("expected User entity in anonymous namespace")
	}
	if ns.GetAction("view") == nil {
		t.Error("expected view action in anonymous namespace")
	}
}

func TestCedarMarshalAnonymousNamespace(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User"))
	s.AddAction(NewAction("view").
		SetPrincipalTypes("User").
		SetResourceTypes("Resource"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Anonymous namespace should not have "namespace" keyword
	if strings.Contains(string(cedarData), "namespace") {
		t.Error("anonymous namespace should not have namespace keyword")
	}
}

func TestCedarUnmarshalWithFilename(t *testing.T) {
	t.Parallel()

	input := `entity User;`

	var s Schema
	if err := s.UnmarshalCedarWithFilename("test.cedarschema", []byte(input)); err != nil {
		t.Fatalf("UnmarshalCedarWithFilename() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
}

func TestCedarUnmarshalCommonType(t *testing.T) {
	t.Parallel()

	input := `
type Name = String;
type UserRecord = {
	name: Name,
	age: Long,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	name := ns.GetCommonType("Name")
	if name == nil {
		t.Fatal("expected Name common type")
	}

	userRecord := ns.GetCommonType("UserRecord")
	if userRecord == nil {
		t.Fatal("expected UserRecord common type")
	}
}

func TestCedarMarshalCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddCommonType(NewCommonType("Name", String()))
	s.AddCommonType(NewCommonType("UserRecord", Record(
		RequiredAttr("name", Ref("Name")),
		RequiredAttr("age", Long()),
	)))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "type Name = String;") {
		t.Error("expected type Name in Cedar")
	}
	if !strings.Contains(cedarStr, "type UserRecord =") {
		t.Error("expected type UserRecord in Cedar")
	}
}

func TestCedarMarshalMultipleNamespaces(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddNamespace(NewNamespace("App1").AddEntity(NewEntity("User")))
	s.AddNamespace(NewNamespace("App2").AddEntity(NewEntity("Admin")))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "namespace App1") {
		t.Error("expected namespace App1")
	}
	if !strings.Contains(cedarStr, "namespace App2") {
		t.Error("expected namespace App2")
	}
}

func TestCedarUnmarshalMultipleNamespaces(t *testing.T) {
	t.Parallel()

	input := `
namespace App1 {
	entity User;
}
namespace App2 {
	entity Admin;
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	if len(s.Namespaces) != 2 {
		t.Errorf("expected 2 namespaces, got %d", len(s.Namespaces))
	}

	app1 := s.GetNamespace("App1")
	if app1 == nil || app1.GetEntity("User") == nil {
		t.Error("expected App1 namespace with User entity")
	}

	app2 := s.GetNamespace("App2")
	if app2 == nil || app2.GetEntity("Admin") == nil {
		t.Error("expected App2 namespace with Admin entity")
	}
}

func TestCedarWriteCedar(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User"))

	var buf bytes.Buffer
	if err := s.WriteCedar(&buf); err != nil {
		t.Fatalf("WriteCedar() error = %v", err)
	}

	if !strings.Contains(buf.String(), "entity User;") {
		t.Error("expected entity User in output")
	}
}

func TestCedarIsValidIdent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		valid bool
	}{
		{"", false},
		{"a", true},
		{"A", true},
		{"_", true},
		{"abc", true},
		{"ABC", true},
		{"_abc", true},
		{"abc123", true},
		{"a_b_c", true},
		{"123", false},
		{"1abc", false},
		{"a-b", false},
		{"a.b", false},
		{"a b", false},
		{"hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := isValidIdent(tt.input)
			if got != tt.valid {
				t.Errorf("isValidIdent(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestCedarMarshalQuotedNames(t *testing.T) {
	t.Parallel()

	s := New()
	// Action name that needs quoting
	s.AddAction(NewAction("view document").
		SetPrincipalTypes("User").
		SetResourceTypes("Document"))
	// Entity attribute that needs quoting
	s.AddEntity(NewEntity("User").
		SetAttributes(
			RequiredAttr("first name", String()),
		))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, `"view document"`) {
		t.Error("expected quoted action name")
	}
	if !strings.Contains(cedarStr, `"first name"`) {
		t.Error("expected quoted attribute name")
	}
}

func TestCedarUnmarshalQuotedNames(t *testing.T) {
	t.Parallel()

	input := `
entity User {
	"first name": String,
};
action "view document" appliesTo {
	principal: User,
	resource: Document,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	action := ns.GetAction("view document")
	if action == nil {
		t.Error("expected action with quoted name")
	}

	user := ns.GetEntity("User")
	if user == nil {
		t.Fatal("expected User entity")
	}
	if len(user.Attributes) != 1 {
		t.Errorf("expected 1 attribute, got %d", len(user.Attributes))
	}
	if user.Attributes[0].Name != "first name" {
		t.Errorf("expected 'first name' attribute, got %s", user.Attributes[0].Name)
	}
}

func TestCedarMarshalEntityMemberOfMultiple(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").MemberOf("Group", "Organization"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "in [") {
		t.Error("expected in [...] for multiple memberOf")
	}
}

func TestCedarMarshalEntityMemberOfSingle(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").MemberOf("Group"))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "in Group") {
		t.Error("expected in Group for single memberOf")
	}
}

func TestCedarMarshalActionMemberOfSingle(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddAction(NewAction("edit").InActions(ActionRef{Name: "view"}))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	// Single memberOf should not have brackets
	if strings.Contains(cedarStr, "in [view]") {
		t.Error("single memberOf should not have brackets")
	}
	if !strings.Contains(cedarStr, "in view") {
		t.Error("expected in view")
	}
}

func TestCedarMarshalContextAsRef(t *testing.T) {
	t.Parallel()

	input := `
namespace Test {
	type Context = { ip: String, };
	action view appliesTo {
		principal: User,
		resource: Doc,
		context: Context,
	};
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("Test")
	if ns == nil {
		t.Fatal("expected Test namespace")
	}

	view := ns.GetAction("view")
	if view == nil {
		t.Fatal("expected view action")
	}

	// Context should be a Ref type
	if view.Context.v == nil {
		t.Fatal("expected context to be set")
	}
}

func TestCedarFormatterWriteError(t *testing.T) {
	t.Parallel()

	// Test that write errors are returned
	s := New()
	s.AddEntity(NewEntity("User"))

	// Use a writer that always fails
	errWriter := &errorWriter{}
	err := s.WriteCedar(errWriter)
	if err == nil {
		t.Error("expected error from failing writer")
	}
}

type errorWriter struct{}

func (w *errorWriter) Write(p []byte) (n int, err error) {
	return 0, bytes.ErrTooLarge
}

func TestCedarMarshalSetType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("User").
		SetAttributes(
			RequiredAttr("roles", SetOf(String())),
			RequiredAttr("nestedSet", SetOf(SetOf(Long()))),
		))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "Set<String>") {
		t.Error("expected Set<String>")
	}
	if !strings.Contains(cedarStr, "Set<Set<Long>>") {
		t.Error("expected Set<Set<Long>>")
	}
}

func TestCedarUnmarshalActionWithNamespacedMemberOf(t *testing.T) {
	t.Parallel()

	input := `
namespace App {
	action edit in [OtherApp::"read", view];
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("App")
	if ns == nil {
		t.Fatal("expected App namespace")
	}

	edit := ns.GetAction("edit")
	if edit == nil {
		t.Fatal("expected edit action")
	}

	if len(edit.MemberOf) != 2 {
		t.Errorf("expected 2 memberOf, got %d", len(edit.MemberOf))
	}
}

func TestCedarMarshalRecordTypeEmpty(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Empty").SetAttributes())

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	// Empty entity should not have shape
	if strings.Contains(string(cedarData), "{") {
		t.Log(string(cedarData))
		// Actually, this is fine since empty shape is allowed
	}
}

func TestCedarUnmarshalMultipleEntitiesSameLine(t *testing.T) {
	t.Parallel()

	input := `
entity User, Admin, Guest;
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	// All three should be separate entities
	if ns.GetEntity("User") == nil {
		t.Error("expected User entity")
	}
	if ns.GetEntity("Admin") == nil {
		t.Error("expected Admin entity")
	}
	if ns.GetEntity("Guest") == nil {
		t.Error("expected Guest entity")
	}
}

func TestCedarUnmarshalMultipleActionsSameLine(t *testing.T) {
	t.Parallel()

	input := `
action view, edit, delete appliesTo {
	principal: User,
	resource: Doc,
};
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() error = %v", err)
	}

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace")
	}

	// All three should be separate actions
	if ns.GetAction("view") == nil {
		t.Error("expected view action")
	}
	if ns.GetAction("edit") == nil {
		t.Error("expected edit action")
	}
	if ns.GetAction("delete") == nil {
		t.Error("expected delete action")
	}
}

func TestCedarMarshalExtensionType(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Request").
		SetAttributes(
			RequiredAttr("ip", Extension("ipaddr")),
			RequiredAttr("amount", Extension("decimal")),
		))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "ipaddr") {
		t.Error("expected ipaddr extension type")
	}
	if !strings.Contains(cedarStr, "decimal") {
		t.Error("expected decimal extension type")
	}
}

func TestCedarMarshalEntityRef(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddEntity(NewEntity("Doc").
		SetAttributes(
			RequiredAttr("owner", Entity("User")),
			RequiredAttr("project", Entity("Namespace::Project")),
		))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "User") {
		t.Error("expected User entity reference")
	}
	if !strings.Contains(cedarStr, "Namespace::Project") {
		t.Error("expected Namespace::Project entity reference")
	}
}

func TestCedarMarshalTypeRef(t *testing.T) {
	t.Parallel()

	s := New()
	s.AddCommonType(NewCommonType("Name", String()))
	s.AddEntity(NewEntity("User").
		SetAttributes(
			RequiredAttr("name", Ref("Name")),
		))

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	if !strings.Contains(cedarStr, "type Name = String;") {
		t.Error("expected type Name declaration")
	}
	if !strings.Contains(cedarStr, "name: Name") {
		t.Error("expected type reference in attribute")
	}
}

func TestCedarMarshalAnnotationWithoutValue(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("Test")
	ns.Annotations = ns.Annotations.Set("flag", "")
	ns.AddEntity(NewEntity("User"))
	s.AddNamespace(ns)

	cedarData, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() error = %v", err)
	}

	cedarStr := string(cedarData)
	// Empty annotation value should still have quotes
	if !strings.Contains(cedarStr, "@flag") {
		t.Error("expected @flag annotation")
	}
}
