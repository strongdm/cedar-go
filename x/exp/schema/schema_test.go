package schema

import (
	"testing"
)

func TestNewSchema(t *testing.T) {
	t.Parallel()

	s := New()
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if s.Namespaces == nil {
		t.Error("Namespaces should not be nil")
	}
	if len(s.Namespaces) != 0 {
		t.Error("Namespaces should be empty")
	}
}

func TestSchemaAddNamespace(t *testing.T) {
	t.Parallel()

	s := New()
	ns := NewNamespace("MyApp")

	result := s.AddNamespace(ns)
	if result != s {
		t.Error("AddNamespace should return schema for chaining")
	}

	got := s.GetNamespace("MyApp")
	if got == nil {
		t.Error("expected namespace to be added")
	}
	if got.Name != "MyApp" {
		t.Errorf("expected MyApp, got %s", got.Name)
	}
}

func TestSchemaAddEntity(t *testing.T) {
	t.Parallel()

	s := New()
	e := NewEntity("User")

	s.AddEntity(e)

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace to be created")
	}

	got := ns.GetEntity("User")
	if got == nil {
		t.Error("expected entity to be added")
	}
}

func TestSchemaAddAction(t *testing.T) {
	t.Parallel()

	s := New()
	a := NewAction("view")

	s.AddAction(a)

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace to be created")
	}

	got := ns.GetAction("view")
	if got == nil {
		t.Error("expected action to be added")
	}
}

func TestSchemaAddCommonType(t *testing.T) {
	t.Parallel()

	s := New()
	ct := NewCommonType("Name", String())

	s.AddCommonType(ct)

	ns := s.GetNamespace("")
	if ns == nil {
		t.Fatal("expected anonymous namespace to be created")
	}

	got := ns.GetCommonType("Name")
	if got == nil {
		t.Error("expected common type to be added")
	}
}

func TestNamespaceNew(t *testing.T) {
	t.Parallel()

	ns := NewNamespace("MyApp")
	if ns.Name != "MyApp" {
		t.Errorf("expected MyApp, got %s", ns.Name)
	}
	if ns.Entities == nil {
		t.Error("Entities should not be nil")
	}
	if ns.Actions == nil {
		t.Error("Actions should not be nil")
	}
	if ns.CommonTypes == nil {
		t.Error("CommonTypes should not be nil")
	}
}

func TestNamespaceAnnotations(t *testing.T) {
	t.Parallel()

	ns := NewNamespace("MyApp").
		Annotate("doc", "My application namespace")

	if len(ns.Annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(ns.Annotations))
	}

	val, ok := ns.Annotations.Get("doc")
	if !ok {
		t.Error("expected annotation to exist")
	}
	if val != "My application namespace" {
		t.Errorf("unexpected annotation value: %s", val)
	}
}

func TestNamespaceGetters(t *testing.T) {
	t.Parallel()

	ns := NewNamespace("MyApp")

	if ns.GetEntity("User") != nil {
		t.Error("expected nil for non-existent entity")
	}
	if ns.GetAction("view") != nil {
		t.Error("expected nil for non-existent action")
	}
	if ns.GetCommonType("Name") != nil {
		t.Error("expected nil for non-existent common type")
	}

	ns.AddEntity(NewEntity("User"))
	ns.AddAction(NewAction("view"))
	ns.AddCommonType(NewCommonType("Name", String()))

	if ns.GetEntity("User") == nil {
		t.Error("expected entity to exist")
	}
	if ns.GetAction("view") == nil {
		t.Error("expected action to exist")
	}
	if ns.GetCommonType("Name") == nil {
		t.Error("expected common type to exist")
	}
}

func TestEntityDecl(t *testing.T) {
	t.Parallel()

	e := NewEntity("User").
		MemberOf("Group", "Organization").
		SetAttributes(
			RequiredAttr("name", String()),
			OptionalAttr("email", String()),
		).
		SetTags(String()).
		Annotate("doc", "A user entity")

	if e.Name != "User" {
		t.Errorf("expected User, got %s", e.Name)
	}
	if len(e.MemberOfTypes) != 2 {
		t.Errorf("expected 2 memberOf types, got %d", len(e.MemberOfTypes))
	}
	if len(e.Attributes) != 2 {
		t.Errorf("expected 2 attributes, got %d", len(e.Attributes))
	}
	if e.Tags.v == nil {
		t.Error("expected tags to be set")
	}
	if len(e.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(e.Annotations))
	}
}

func TestEntityDeclEnum(t *testing.T) {
	t.Parallel()

	e := NewEntity("Color").
		SetEnum("red", "green", "blue")

	if len(e.Enum) != 3 {
		t.Errorf("expected 3 enum values, got %d", len(e.Enum))
	}
	if e.Enum[0] != "red" || e.Enum[1] != "green" || e.Enum[2] != "blue" {
		t.Error("unexpected enum values")
	}
}

func TestActionDecl(t *testing.T) {
	t.Parallel()

	a := NewAction("view").
		InActions(
			ActionRef{Namespace: "BaseActions", Name: "read"},
		).
		SetPrincipalTypes("User", "Group").
		SetResourceTypes("Document", "Folder").
		SetContext(Record(
			RequiredAttr("ip", Extension("ipaddr")),
		)).
		Annotate("doc", "View action")

	if a.Name != "view" {
		t.Errorf("expected view, got %s", a.Name)
	}
	if len(a.MemberOf) != 1 {
		t.Errorf("expected 1 memberOf, got %d", len(a.MemberOf))
	}
	if len(a.PrincipalTypes) != 2 {
		t.Errorf("expected 2 principal types, got %d", len(a.PrincipalTypes))
	}
	if len(a.ResourceTypes) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(a.ResourceTypes))
	}
	if a.Context.v == nil {
		t.Error("expected context to be set")
	}
	if len(a.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(a.Annotations))
	}
}

func TestCommonTypeDecl(t *testing.T) {
	t.Parallel()

	ct := NewCommonType("Name", String()).
		Annotate("doc", "A name type")

	if ct.Name != "Name" {
		t.Errorf("expected Name, got %s", ct.Name)
	}
	if ct.Type.v == nil {
		t.Error("expected type to be set")
	}
	if len(ct.Annotations) != 1 {
		t.Errorf("expected 1 annotation, got %d", len(ct.Annotations))
	}
}

func TestAnnotations(t *testing.T) {
	t.Parallel()

	var ann Annotations

	// Get from empty
	_, ok := ann.Get("key")
	if ok {
		t.Error("expected false for non-existent key")
	}

	// Set
	ann = ann.Set("key", "value")
	val, ok := ann.Get("key")
	if !ok {
		t.Error("expected key to exist")
	}
	if val != "value" {
		t.Errorf("expected value, got %s", val)
	}

	// Update existing
	ann = ann.Set("key", "new value")
	val, ok = ann.Get("key")
	if !ok {
		t.Error("expected key to still exist")
	}
	if val != "new value" {
		t.Errorf("expected new value, got %s", val)
	}

	// Add another
	ann = ann.Set("key2", "value2")
	if len(ann) != 2 {
		t.Errorf("expected 2 annotations, got %d", len(ann))
	}
}

func TestBuilderChaining(t *testing.T) {
	t.Parallel()

	// Test that all builders return the same pointer for chaining
	s := New()
	if s.AddNamespace(NewNamespace("A")) != s {
		t.Error("AddNamespace should return schema")
	}
	if s.AddEntity(NewEntity("E")) != s {
		t.Error("AddEntity should return schema")
	}
	if s.AddAction(NewAction("a")) != s {
		t.Error("AddAction should return schema")
	}
	if s.AddCommonType(NewCommonType("T", String())) != s {
		t.Error("AddCommonType should return schema")
	}

	ns := NewNamespace("B")
	if ns.AddEntity(NewEntity("E")) != ns {
		t.Error("Namespace.AddEntity should return namespace")
	}
	if ns.AddAction(NewAction("a")) != ns {
		t.Error("Namespace.AddAction should return namespace")
	}
	if ns.AddCommonType(NewCommonType("T", String())) != ns {
		t.Error("Namespace.AddCommonType should return namespace")
	}
	if ns.Annotate("k", "v") != ns {
		t.Error("Namespace.Annotate should return namespace")
	}

	e := NewEntity("E")
	if e.MemberOf("P") != e {
		t.Error("EntityDecl.MemberOf should return entity")
	}
	if e.SetAttributes() != e {
		t.Error("EntityDecl.SetAttributes should return entity")
	}
	if e.SetTags(String()) != e {
		t.Error("EntityDecl.SetTags should return entity")
	}
	if e.SetEnum("a") != e {
		t.Error("EntityDecl.SetEnum should return entity")
	}
	if e.Annotate("k", "v") != e {
		t.Error("EntityDecl.Annotate should return entity")
	}

	a := NewAction("a")
	if a.InActions(ActionRef{Name: "b"}) != a {
		t.Error("ActionDecl.InActions should return action")
	}
	if a.SetPrincipalTypes("P") != a {
		t.Error("ActionDecl.SetPrincipalTypes should return action")
	}
	if a.SetResourceTypes("R") != a {
		t.Error("ActionDecl.SetResourceTypes should return action")
	}
	if a.SetContext(Record()) != a {
		t.Error("ActionDecl.SetContext should return action")
	}
	if a.Annotate("k", "v") != a {
		t.Error("ActionDecl.Annotate should return action")
	}

	ct := NewCommonType("T", String())
	if ct.Annotate("k", "v") != ct {
		t.Error("CommonTypeDecl.Annotate should return common type")
	}
}

func TestComplexSchema(t *testing.T) {
	t.Parallel()

	s := New()

	// Add a namespace with entities and actions
	ns := NewNamespace("PhotoApp").
		Annotate("version", "1.0")

	ns.AddEntity(NewEntity("User").
		MemberOf("Group").
		SetAttributes(
			RequiredAttr("name", String()),
			RequiredAttr("email", String()),
			OptionalAttr("age", Long()),
		))

	ns.AddEntity(NewEntity("Group").
		SetAttributes(
			RequiredAttr("name", String()),
		))

	ns.AddEntity(NewEntity("Photo").
		SetAttributes(
			RequiredAttr("title", String()),
			RequiredAttr("owner", Entity("User")),
			OptionalAttr("tags", SetOf(String())),
		))

	ns.AddAction(NewAction("view").
		SetPrincipalTypes("User", "Group").
		SetResourceTypes("Photo"))

	ns.AddAction(NewAction("edit").
		InActions(ActionRef{Name: "view"}).
		SetPrincipalTypes("User").
		SetResourceTypes("Photo").
		SetContext(Record(
			RequiredAttr("ip", Extension("ipaddr")),
		)))

	ns.AddCommonType(NewCommonType("Name", String()))

	s.AddNamespace(ns)

	// Verify structure
	if len(s.Namespaces) != 1 {
		t.Fatalf("expected 1 namespace, got %d", len(s.Namespaces))
	}

	gotNS := s.GetNamespace("PhotoApp")
	if gotNS == nil {
		t.Fatal("expected PhotoApp namespace")
	}

	if len(gotNS.Entities) != 3 {
		t.Errorf("expected 3 entities, got %d", len(gotNS.Entities))
	}

	if len(gotNS.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(gotNS.Actions))
	}

	if len(gotNS.CommonTypes) != 1 {
		t.Errorf("expected 1 common type, got %d", len(gotNS.CommonTypes))
	}
}

func TestSchemaEnsureNamespaceIdempotent(t *testing.T) {
	t.Parallel()

	s := New()

	// Add an entity to anonymous namespace
	s.AddEntity(NewEntity("A"))

	// Add another entity to anonymous namespace
	s.AddEntity(NewEntity("B"))

	// Should still have just one anonymous namespace
	if len(s.Namespaces) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(s.Namespaces))
	}

	ns := s.GetNamespace("")
	if len(ns.Entities) != 2 {
		t.Errorf("expected 2 entities, got %d", len(ns.Entities))
	}
}

func TestActionRef(t *testing.T) {
	t.Parallel()

	// Simple ref
	ref1 := ActionRef{Name: "view"}
	if ref1.Namespace != "" {
		t.Error("expected empty namespace")
	}
	if ref1.Name != "view" {
		t.Errorf("expected view, got %s", ref1.Name)
	}

	// Qualified ref
	ref2 := ActionRef{Namespace: "OtherApp", Name: "read"}
	if ref2.Namespace != "OtherApp" {
		t.Errorf("expected OtherApp, got %s", ref2.Namespace)
	}
	if ref2.Name != "read" {
		t.Errorf("expected read, got %s", ref2.Name)
	}
}

func TestSchemaWithNilNamespaces(t *testing.T) {
	t.Parallel()

	// Test that ensureNamespace handles nil map
	s := &Schema{}
	s.AddEntity(NewEntity("A"))

	if s.Namespaces == nil {
		t.Error("Namespaces should be initialized")
	}
	if len(s.Namespaces) != 1 {
		t.Errorf("expected 1 namespace, got %d", len(s.Namespaces))
	}
}

func TestEntityMemberOfMultipleCalls(t *testing.T) {
	t.Parallel()

	e := NewEntity("User")
	e.MemberOf("Group")
	e.MemberOf("Organization")

	if len(e.MemberOfTypes) != 2 {
		t.Errorf("expected 2 memberOf types, got %d", len(e.MemberOfTypes))
	}
}

func TestActionMultipleCalls(t *testing.T) {
	t.Parallel()

	a := NewAction("modify")
	a.InActions(ActionRef{Name: "view"})
	a.InActions(ActionRef{Name: "read"})
	a.SetPrincipalTypes("User")
	a.SetPrincipalTypes("Admin")
	a.SetResourceTypes("Doc")
	a.SetResourceTypes("File")

	if len(a.MemberOf) != 2 {
		t.Errorf("expected 2 memberOf, got %d", len(a.MemberOf))
	}
	if len(a.PrincipalTypes) != 2 {
		t.Errorf("expected 2 principal types, got %d", len(a.PrincipalTypes))
	}
	if len(a.ResourceTypes) != 2 {
		t.Errorf("expected 2 resource types, got %d", len(a.ResourceTypes))
	}
}

func TestGetNamespaceNonExistent(t *testing.T) {
	t.Parallel()

	s := New()
	if s.GetNamespace("NonExistent") != nil {
		t.Error("expected nil for non-existent namespace")
	}
}
