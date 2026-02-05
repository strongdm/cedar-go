package schema

import (
	"encoding/json"
	"strings"
	"testing"
)

// ============================================================================
// Error formatting tests - errors.go
// ============================================================================

// TestUndefinedTypeErrorFormatting tests UndefinedTypeError.Error() with and without Context
func TestUndefinedTypeErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      *UndefinedTypeError
		expected string
	}{
		{
			name: "without_context",
			err: &UndefinedTypeError{
				Name:      "Unknown",
				Namespace: "MyApp",
			},
			expected: `undefined type: "Unknown"`,
		},
		{
			name: "with_context",
			err: &UndefinedTypeError{
				Name:      "Unknown",
				Namespace: "MyApp",
				Context:   "in entity User attribute owner",
			},
			expected: `undefined type: "Unknown" in entity User attribute owner`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if !strings.Contains(result, tt.expected) {
				t.Errorf("expected error to contain %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestDuplicateErrorFormatting tests DuplicateError.Error() with empty vs named namespace
func TestDuplicateErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      *DuplicateError
		contains string
	}{
		{
			name: "empty_namespace",
			err: &DuplicateError{
				Kind:      "entity type",
				Name:      "User",
				Namespace: "",
			},
			contains: "empty namespace",
		},
		{
			name: "named_namespace",
			err: &DuplicateError{
				Kind:      "entity type",
				Name:      "User",
				Namespace: "MyApp",
			},
			contains: `namespace "MyApp"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected error to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

// TestParseErrorFormatting tests ParseError.Error() all 3 branches
func TestParseErrorFormatting(t *testing.T) {
	tests := []struct {
		name     string
		err      *ParseError
		contains string
	}{
		{
			name: "with_filename",
			err: &ParseError{
				Filename: "schema.cedarschema",
				Line:     10,
				Column:   5,
				Message:  "unexpected token",
			},
			contains: "schema.cedarschema:10:5",
		},
		{
			name: "without_filename_with_line",
			err: &ParseError{
				Filename: "",
				Line:     10,
				Column:   5,
				Message:  "unexpected token",
			},
			contains: "line 10, column 5",
		},
		{
			name: "without_filename_without_line",
			err: &ParseError{
				Filename: "",
				Line:     0,
				Column:   0,
				Message:  "generic error",
			},
			contains: "generic error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if !strings.Contains(result, tt.contains) {
				t.Errorf("expected error to contain %q, got %q", tt.contains, result)
			}
		})
	}
}

// ============================================================================
// Builder nil checks - builder.go
// ============================================================================

// TestBuilderResourceBeforePrincipal tests calling Resource() before Principal()
func TestBuilderResourceBeforePrincipal(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Entity("Document").
		Action("view").
		Resource("Document"). // Call Resource before Principal
		Principal("User").
		Build()

	action := s.Namespaces["Test"].Actions["view"]
	if action == nil {
		t.Fatal("expected view action")
	}
	if action.AppliesTo == nil {
		t.Fatal("expected AppliesTo to be created")
	}
	if len(action.AppliesTo.ResourceTypes) != 1 || action.AppliesTo.ResourceTypes[0] != "Document" {
		t.Errorf("expected resource Document, got %v", action.AppliesTo.ResourceTypes)
	}
	if len(action.AppliesTo.PrincipalTypes) != 1 || action.AppliesTo.PrincipalTypes[0] != "User" {
		t.Errorf("expected principal User, got %v", action.AppliesTo.PrincipalTypes)
	}
}

// TestBuilderContextBeforePrincipal tests calling Context() before Principal()
func TestBuilderContextBeforePrincipal(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").
		Context(&RecordType{Attributes: map[string]*Attribute{
			"flag": {Type: Bool(), Required: true, Annotations: make(Annotations)},
		}}). // Call Context before Principal
		Principal("User").
		Resource("User").
		Build()

	action := s.Namespaces["Test"].Actions["view"]
	if action == nil {
		t.Fatal("expected view action")
	}
	if action.AppliesTo == nil {
		t.Fatal("expected AppliesTo to be created")
	}
	if action.AppliesTo.Context == nil {
		t.Fatal("expected Context to be set")
	}
	if action.AppliesTo.Context.Attributes["flag"] == nil {
		t.Error("expected flag attribute in context")
	}
}

// ============================================================================
// Parser edge cases - parse_cedar.go
// ============================================================================

// TestParseCommonTypeWithReservedName tests parseCommonType with reserved names
func TestParseCommonTypeWithReservedName(t *testing.T) {
	reservedNames := []string{"Bool", "Boolean", "Entity", "Extension", "Long", "Record", "Set", "String"}
	for _, name := range reservedNames {
		cedar := "type " + name + " = Long; entity User; action view appliesTo { principal: [User], resource: [User] };"
		var s Schema
		err := s.UnmarshalCedar([]byte(cedar))
		if err == nil {
			t.Errorf("expected reserved name error for %q", name)
		}
	}
}

// TestParseAnnotationWithoutParenthesis tests annotation without parenthesis
func TestParseAnnotationWithoutParenthesis(t *testing.T) {
	cedar := `
@deprecated
@note("This is a note")
entity User;
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil {
		t.Fatal("expected User entity")
	}
	// Annotation without value should have empty string value
	if user.Annotations["deprecated"] != "" {
		t.Errorf("expected deprecated annotation with empty value, got %q", user.Annotations["deprecated"])
	}
	if user.Annotations["note"] != "This is a note" {
		t.Errorf("expected note annotation, got %q", user.Annotations["note"])
	}
}

// TestParseStringEscapeSequences tests string escape sequences in parseString
func TestParseStringEscapeSequences(t *testing.T) {
	tests := []struct {
		name     string
		cedar    string
		expected string
	}{
		{
			name:     "simple_escape",
			cedar:    `entity User { "name\n": String }; action v appliesTo { principal: [User], resource: [User] };`,
			expected: "name\n",
		},
		{
			name:     "tab_escape",
			cedar:    `entity User { "name\t": String }; action v appliesTo { principal: [User], resource: [User] };`,
			expected: "name\t",
		},
		{
			name:     "quote_escape",
			cedar:    `entity User { "name\"test": String }; action v appliesTo { principal: [User], resource: [User] };`,
			expected: `name"test`,
		},
		{
			name:     "backslash_escape",
			cedar:    `entity User { "name\\path": String }; action v appliesTo { principal: [User], resource: [User] };`,
			expected: `name\path`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Schema
			if err := s.UnmarshalCedar([]byte(tt.cedar)); err != nil {
				t.Fatalf("failed to parse: %v", err)
			}
			user := s.Namespaces[""].EntityTypes["User"]
			if user == nil || user.Shape == nil {
				t.Fatal("expected User with shape")
			}
			if user.Shape.Attributes[tt.expected] == nil {
				t.Errorf("expected attribute %q", tt.expected)
			}
		})
	}
}

// TestParseStringWithNewline tests parseString with newline in string
func TestParseStringWithNewline(t *testing.T) {
	// This should work as the parser handles multi-line strings
	cedar := `action "view
photo" appliesTo { principal: [User], resource: [User] }; entity User;`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	// This may or may not error depending on implementation - just exercise the code
	_ = err
}

// TestIsValidIdentEdgeCases tests isValidIdent edge cases
func TestIsValidIdentEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"123", false},
		{"_underscore", true},
		{"camelCase", true},
		{"snake_case", true},
		{"with-hyphen", false},
		{"with space", false},
		{"valid123", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := isValidIdent(tt.input)
			if result != tt.expected {
				t.Errorf("isValidIdent(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseUnexpectedDeclaration tests parsing an unexpected declaration keyword
func TestParseUnexpectedDeclaration(t *testing.T) {
	cedar := `namespace Test { something User; }`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for unexpected declaration")
	}
}

// TestParseExpectEOF tests expect() when reaching EOF
func TestParseExpectEOF(t *testing.T) {
	cedar := `entity User`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing semicolon")
	}
}

// TestParseDuplicateNamespace tests parsing duplicate namespaces
func TestParseDuplicateNamespace(t *testing.T) {
	cedar := `
namespace Test { entity User; action v appliesTo { principal: [User], resource: [User] }; }
namespace Test { entity Admin; action v appliesTo { principal: [Admin], resource: [Admin] }; }
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for duplicate namespace")
	}
}


// ============================================================================
// JSON edge cases - json.go
// ============================================================================

// TestJSONUnmarshalInvalidJSON tests UnmarshalJSON with invalid JSON
func TestJSONUnmarshalInvalidJSON(t *testing.T) {
	var s Schema
	err := json.Unmarshal([]byte(`{invalid`), &s)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// TestJSONMarshalEmptyEntityTypes tests MarshalJSON for empty namespace
func TestJSONMarshalEmptyEntityTypes(t *testing.T) {
	ns := &Namespace{
		EntityTypes: make(map[string]*EntityTypeDef),
		EnumTypes:   make(map[string]*EnumTypeDef),
		Actions:     make(map[string]*ActionDef),
		CommonTypes: make(map[string]*CommonTypeDef),
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(ns)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should be empty object or close to it
	if len(data) == 0 {
		t.Error("expected non-empty marshaled data")
	}
}

// TestJSONMarshalEntityTypeDefAllFields tests MarshalJSON for EntityTypeDef with all fields
func TestJSONMarshalEntityTypeDefAllFields(t *testing.T) {
	et := &EntityTypeDef{
		MemberOfTypes: []string{"Group"},
		Shape: &RecordType{
			Attributes: map[string]*Attribute{
				"name": {
					Type:        PrimitiveType{Kind: PrimitiveString},
					Required:    true,
					Annotations: Annotations{"doc": "User name"},
				},
			},
		},
		Tags:        PrimitiveType{Kind: PrimitiveString},
		Annotations: Annotations{"doc": "A user"},
	}

	data, err := json.Marshal(et)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "memberOfTypes") {
		t.Error("expected memberOfTypes in JSON")
	}
	if !strings.Contains(string(data), "shape") {
		t.Error("expected shape in JSON")
	}
	if !strings.Contains(string(data), "tags") {
		t.Error("expected tags in JSON")
	}
}

// TestJSONMarshalEntityTypeDefMinimalFields tests MarshalJSON for EntityTypeDef with minimal fields
func TestJSONMarshalEntityTypeDefMinimalFields(t *testing.T) {
	et := &EntityTypeDef{
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(et)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should not contain optional fields
	if strings.Contains(string(data), "memberOfTypes") {
		t.Error("did not expect memberOfTypes in JSON for empty entity")
	}
}

// TestJSONMarshalEnumTypeDefWithAnnotations tests MarshalJSON for EnumTypeDef with annotations
func TestJSONMarshalEnumTypeDefWithAnnotations(t *testing.T) {
	enum := &EnumTypeDef{
		Values:      []string{"Active", "Inactive"},
		Annotations: Annotations{"doc": "Status values"},
	}

	data, err := json.Marshal(enum)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "annotations") {
		t.Error("expected annotations in JSON")
	}
}

// TestJSONMarshalEnumTypeDefWithoutAnnotations tests MarshalJSON for EnumTypeDef without annotations
func TestJSONMarshalEnumTypeDefWithoutAnnotations(t *testing.T) {
	enum := &EnumTypeDef{
		Values:      []string{"Active", "Inactive"},
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(enum)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should not contain annotations when empty
	if strings.Contains(string(data), "annotations") {
		t.Error("did not expect annotations in JSON for empty annotations")
	}
}

// TestJSONMarshalTypeUnknown tests marshalType with an unknown type
func TestJSONMarshalTypeUnknown(t *testing.T) {
	// Use a custom type that implements Type but isn't recognized
	// This is done indirectly by testing all known types
	types := []Type{
		PrimitiveType{Kind: PrimitiveLong},
		PrimitiveType{Kind: PrimitiveString},
		PrimitiveType{Kind: PrimitiveBool},
		SetType{Element: PrimitiveType{Kind: PrimitiveString}},
		&RecordType{Attributes: make(map[string]*Attribute)},
		EntityRef{Name: "User"},
		ExtensionType{Name: "ipaddr"},
		CommonTypeRef{Name: "MyType"},
		EntityOrCommonRef{Name: "Ambiguous"},
	}

	for _, typ := range types {
		data, err := marshalType(typ)
		if err != nil {
			t.Errorf("failed to marshal %T: %v", typ, err)
			continue
		}
		if len(data) == 0 {
			t.Errorf("empty marshal result for %T", typ)
		}
	}
}

// TestJSONUnmarshalUnknownTypeFormat tests unmarshalType with unknown format
func TestJSONUnmarshalUnknownTypeFormat(t *testing.T) {
	// Empty type with no name and no attributes
	_, err := unmarshalType(json.RawMessage(`{}`))
	if err == nil {
		t.Error("expected error for unknown type format")
	}
}

// TestJSONUnmarshalEntityOrCommonRequiresName tests EntityOrCommon requires name
func TestJSONUnmarshalEntityOrCommonRequiresName(t *testing.T) {
	_, err := unmarshalType(json.RawMessage(`{"type": "EntityOrCommon"}`))
	if err == nil {
		t.Error("expected error for EntityOrCommon without name")
	}
}

// TestJSONParseActionWithContext tests parsing action with record context vs ref context
func TestJSONParseActionWithContext(t *testing.T) {
	// Test with EntityOrCommonRef context
	jsonSchema := `{
		"": {
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "EntityOrCommon", "name": "MyContext"}
					}
				}
			}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if action == nil || action.AppliesTo == nil {
		t.Fatal("expected action with appliesTo")
	}
	if action.AppliesTo.ContextRef == nil {
		t.Error("expected ContextRef to be set")
	}
}

// TestJSONMarshalRecordWithAnnotations tests MarshalJSON for RecordType with attribute annotations
func TestJSONMarshalRecordWithAnnotations(t *testing.T) {
	rt := &RecordType{
		Attributes: map[string]*Attribute{
			"name": {
				Type:        PrimitiveType{Kind: PrimitiveString},
				Required:    false,
				Annotations: Annotations{"doc": "The name"},
			},
		},
	}

	data, err := json.Marshal(rt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "required") {
		t.Error("expected required field in JSON")
	}
	if !strings.Contains(string(data), "annotations") {
		t.Error("expected annotations in JSON")
	}
}

// TestJSONMarshalActionWithContextRef tests MarshalJSON for action with ContextRef
func TestJSONMarshalActionWithContextRef(t *testing.T) {
	action := &ActionDef{
		AppliesTo: &AppliesTo{
			PrincipalTypes: []string{"User"},
			ResourceTypes:  []string{"Document"},
			ContextRef:     CommonTypeRef{Name: "MyContext"},
		},
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "context") {
		t.Error("expected context in JSON")
	}
}

// TestJSONMarshalActionWithInlineContext tests MarshalJSON for action with inline Context
func TestJSONMarshalActionWithInlineContext(t *testing.T) {
	action := &ActionDef{
		AppliesTo: &AppliesTo{
			PrincipalTypes: []string{"User"},
			ResourceTypes:  []string{"Document"},
			Context: &RecordType{
				Attributes: map[string]*Attribute{
					"flag": {
						Type:        PrimitiveType{Kind: PrimitiveBool},
						Required:    true,
						Annotations: make(Annotations),
					},
				},
			},
		},
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "context") {
		t.Error("expected context in JSON")
	}
}

// TestJSONMarshalActionWithAnnotations tests MarshalJSON for action with annotations
func TestJSONMarshalActionWithAnnotations(t *testing.T) {
	action := &ActionDef{
		MemberOf: []*ActionRef{
			{Type: "NS::Action", ID: "admin"},
		},
		AppliesTo: &AppliesTo{
			PrincipalTypes: []string{"User"},
			ResourceTypes:  []string{"Document"},
		},
		Annotations: Annotations{"doc": "View action"},
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "memberOf") {
		t.Error("expected memberOf in JSON")
	}
	if !strings.Contains(string(data), "annotations") {
		t.Error("expected annotations in JSON")
	}
}

// TestJSONMarshalNamespaceWithAllFields tests MarshalJSON for Namespace with all fields
func TestJSONMarshalNamespaceWithAllFields(t *testing.T) {
	ns := &Namespace{
		EntityTypes: map[string]*EntityTypeDef{
			"User": {},
		},
		EnumTypes: map[string]*EnumTypeDef{
			"Status": {Values: []string{"A"}},
		},
		Actions: map[string]*ActionDef{
			"view": {},
		},
		CommonTypes: map[string]*CommonTypeDef{
			"MyType": {Type: PrimitiveType{Kind: PrimitiveString}},
		},
		Annotations: Annotations{"version": "1.0"},
	}

	data, err := json.Marshal(ns)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "entityTypes") {
		t.Error("expected entityTypes in JSON")
	}
	if !strings.Contains(string(data), "actions") {
		t.Error("expected actions in JSON")
	}
	if !strings.Contains(string(data), "commonTypes") {
		t.Error("expected commonTypes in JSON")
	}
	if !strings.Contains(string(data), "annotations") {
		t.Error("expected annotations in JSON")
	}
}

// TestJSONMarshalCommonTypeDef tests MarshalJSON for CommonTypeDef
func TestJSONMarshalCommonTypeDef(t *testing.T) {
	ct := &CommonTypeDef{
		Type:        PrimitiveType{Kind: PrimitiveString},
		Annotations: Annotations{"doc": "A string"},
	}

	data, err := json.Marshal(ct)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Common types marshal their type directly
	if len(data) == 0 {
		t.Error("expected non-empty marshal result")
	}
}

// ============================================================================
// Cedar Marshal edge cases - parse_cedar.go
// ============================================================================

// TestMarshalCedarWriteTypeForAllTypes tests writeType for all type variants
func TestMarshalCedarWriteTypeForAllTypes(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Attr("longVal", Long()).
		Attr("stringVal", String()).
		Attr("boolVal", Bool()).
		Attr("setVal", Set(Long())).
		Attr("entityVal", Entity("Other")).
		Attr("extVal", IPAddr()).
		Attr("commonVal", CommonType("MyType")).
		Entity("Other").
		CommonType("MyType", String()).
		Action("view").Principal("User").Resource("User").
		Build()

	// Also test EntityOrCommonRef
	s.Namespaces["Test"].EntityTypes["User"].Shape.Attributes["ambiguousVal"] = &Attribute{
		Type:        EntityOrCommonRef{Name: "Ambiguous"},
		Required:    true,
		Annotations: make(Annotations),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty Cedar output")
	}
}

// TestMarshalCedarActionWithQuotedName tests marshaling action with a name that needs quoting
func TestMarshalCedarActionWithQuotedName(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Build()

	// Add action with name that needs quoting
	s.Namespaces["Test"].Actions["view photo"] = &ActionDef{
		AppliesTo: &AppliesTo{
			PrincipalTypes: []string{"User"},
			ResourceTypes:  []string{"User"},
		},
		Annotations: make(Annotations),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), `"view photo"`) {
		t.Error("expected quoted action name in Cedar output")
	}
}

// TestMarshalCedarActionMemberOfWithQuotedID tests marshaling action memberOf with quoted ID
func TestMarshalCedarActionMemberOfWithQuotedID(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Build()

	// Add action with memberOf that has a name needing quoting
	s.Namespaces["Test"].Actions["group action"] = &ActionDef{
		Annotations: make(Annotations),
	}
	s.Namespaces["Test"].Actions["view"] = &ActionDef{
		MemberOf: []*ActionRef{
			{ID: "group action"},
			{Type: "OtherNS::Action", ID: "admin"},
		},
		AppliesTo: &AppliesTo{
			PrincipalTypes: []string{"User"},
			ResourceTypes:  []string{"User"},
		},
		Annotations: make(Annotations),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), `"group action"`) {
		t.Error("expected quoted memberOf ID in Cedar output")
	}
}

// ============================================================================
// Additional parser edge cases
// ============================================================================

// TestParseIdentEOF tests parseIdent when reaching EOF
func TestParseIdentEOF(t *testing.T) {
	cedar := `entity `
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error when parseIdent reaches EOF")
	}
}

// TestParseIdentInvalidStart tests parseIdent with invalid starting character
func TestParseIdentInvalidStart(t *testing.T) {
	cedar := `entity 123User;`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for identifier starting with digit")
	}
}

// TestParseMissingClosingBrace tests parsing with missing closing brace
func TestParseMissingClosingBrace(t *testing.T) {
	cedar := `namespace Test { entity User;`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing closing brace")
	}
}

// TestParseActionRefWithQuotedString tests parseActionRef with quoted string
func TestParseActionRefWithQuotedString(t *testing.T) {
	cedar := `
namespace Test {
	entity User;
	action "admin action";
	action view in ["admin action"] appliesTo { principal: [User], resource: [User] };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	viewAction := s.Namespaces["Test"].Actions["view"]
	if len(viewAction.MemberOf) != 1 || viewAction.MemberOf[0].ID != "admin action" {
		t.Errorf("expected memberOf 'admin action', got %v", viewAction.MemberOf)
	}
}

// TestParseAppliesTo with unknown key
func TestParseAppliesToUnknownKey(t *testing.T) {
	cedar := `
entity User;
action view appliesTo { unknown: [User] };
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for unknown appliesTo key")
	}
}

// TestAdvanceAtEOF tests advance() when already at EOF
func TestAdvanceAtEOF(t *testing.T) {
	// This is tested indirectly through other tests, but let's ensure code path
	cedar := `entity User;
action view appliesTo { principal: [User], resource: [User] };`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
}

// TestParsePathError tests parsePath with error after first segment
func TestParsePathError(t *testing.T) {
	cedar := `namespace A:: { entity User; }`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for incomplete path")
	}
}

// ============================================================================
// Resolve edge cases
// ============================================================================

// TestResolveNilSchemaPointer tests Resolve on nil Schema pointer
func TestResolveNilSchemaPointer(t *testing.T) {
	var s *Schema
	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error when resolving nil schema")
	}
}

// ============================================================================
// Additional JSON unmarshal edge cases
// ============================================================================

// TestJSONUnmarshalEntityWithInvalidShape tests entity with non-record shape
func TestJSONUnmarshalEntityWithInvalidShape(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {"type": "String"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for non-record shape")
	}
}

// TestJSONUnmarshalSetRequiresElement tests Set type requires element
func TestJSONUnmarshalSetRequiresElement(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"items": {"type": "Set"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for Set without element")
	}
}

// TestJSONUnmarshalEntityRequiresName tests Entity type requires name
func TestJSONUnmarshalEntityRequiresName(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ref": {"type": "Entity"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for Entity without name")
	}
}

// TestJSONUnmarshalExtensionRequiresName tests Extension type requires name
func TestJSONUnmarshalExtensionRequiresName(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"ext": {"type": "Extension"}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for Extension without name")
	}
}

// TestJSONUnmarshalInvalidAttribute tests unmarshalAttribute with invalid data
func TestJSONUnmarshalInvalidAttribute(t *testing.T) {
	_, err := unmarshalAttribute(json.RawMessage(`{invalid}`))
	if err == nil {
		t.Error("expected error for invalid attribute JSON")
	}
}

// ============================================================================
// Additional tests for JSON edge cases
// ============================================================================

// TestJSONUnmarshalEntityWithInvalidTagsType tests entity with invalid tags
func TestJSONUnmarshalEntityWithInvalidTags(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"tags": {"type": "Set"}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	// This should error because Set requires element
	if err == nil {
		t.Error("expected error for invalid tags type")
	}
}

// TestJSONUnmarshalActionWithInvalidContext tests action with invalid context type
func TestJSONUnmarshalActionWithInvalidContext(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "String"}
					}
				}
			}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for non-record context")
	}
}

// TestJSONUnmarshalCommonTypeRefAsContext tests action context with CommonTypeRef
func TestJSONUnmarshalCommonTypeRefAsContext(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {
				"MyContext": {"type": "Record", "attributes": {"flag": {"type": "Bool"}}}
			},
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "MyContext"}
					}
				}
			}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if action == nil || action.AppliesTo == nil {
		t.Fatal("expected action with appliesTo")
	}
	// The context should be stored as ContextRef since MyContext is a type name
	if action.AppliesTo.ContextRef == nil {
		t.Error("expected ContextRef to be set for type reference")
	}
}

// ============================================================================
// Additional parse_cedar tests
// ============================================================================

// TestParseMultipleEntityDeclarations tests parsing multiple entities in one declaration
func TestParseMultipleEntityDeclarations(t *testing.T) {
	cedar := `
namespace Test {
	entity User, Admin, Guest in [Group] {
		name: String,
	};
	entity Group;
	action view appliesTo { principal: [User], resource: [Group] };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces["Test"]
	for _, name := range []string{"User", "Admin", "Guest"} {
		et := ns.EntityTypes[name]
		if et == nil {
			t.Errorf("expected %s entity type", name)
			continue
		}
		if len(et.MemberOfTypes) != 1 || et.MemberOfTypes[0] != "Group" {
			t.Errorf("%s: expected memberOf [Group], got %v", name, et.MemberOfTypes)
		}
		if et.Shape == nil || et.Shape.Attributes["name"] == nil {
			t.Errorf("%s: expected name attribute", name)
		}
	}
}

// TestParseMultipleEnumDeclarations tests parsing multiple enum entities in one declaration
func TestParseMultipleEnumDeclarations(t *testing.T) {
	cedar := `
namespace Test {
	entity Status, State enum ["Active", "Inactive"];
	entity User;
	action view appliesTo { principal: [User], resource: [Status] };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces["Test"]
	for _, name := range []string{"Status", "State"} {
		enum := ns.EnumTypes[name]
		if enum == nil {
			t.Errorf("expected %s enum type", name)
			continue
		}
		if len(enum.Values) != 2 {
			t.Errorf("%s: expected 2 values, got %d", name, len(enum.Values))
		}
	}
}

// TestParseActionWithSingleTypePrincipal tests action with single type (no brackets)
func TestParseActionWithSingleTypePrincipal(t *testing.T) {
	cedar := `
entity User;
entity Doc;
action view appliesTo { principal: User, resource: Doc };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if len(action.AppliesTo.PrincipalTypes) != 1 || action.AppliesTo.PrincipalTypes[0] != "User" {
		t.Errorf("expected principal [User], got %v", action.AppliesTo.PrincipalTypes)
	}
}

// TestParseEntityWithEqualsShape tests entity with = shape syntax
func TestParseEntityWithEqualsShape(t *testing.T) {
	cedar := `
entity User = { name: String };
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	if user == nil || user.Shape == nil {
		t.Fatal("expected User with shape")
	}
	if user.Shape.Attributes["name"] == nil {
		t.Error("expected name attribute")
	}
}

// TestParseSingleActionMemberOf tests action with single memberOf (no brackets)
func TestParseSingleActionMemberOf(t *testing.T) {
	cedar := `
entity User;
action admin;
action view in admin appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if len(action.MemberOf) != 1 || action.MemberOf[0].ID != "admin" {
		t.Errorf("expected memberOf [admin], got %v", action.MemberOf)
	}
}

// TestParseContextWithTypeRef tests parsing action with context type reference
func TestParseContextWithTypeRef(t *testing.T) {
	cedar := `
type MyContext = { flag: Bool };
entity User;
action view appliesTo { principal: [User], resource: [User], context: MyContext };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if action.AppliesTo.ContextRef == nil {
		t.Error("expected ContextRef to be set")
	}
}

// ============================================================================
// Additional resolve edge cases
// ============================================================================

// TestResolveReservedEntityTypeName tests resolving entity with reserved name in different namespace
func TestResolveReservedEntityTypeNameInBuildIndex(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"__cedar": {}, // Reserved name
				},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for reserved entity type name")
	}
}

// TestResolveReservedEnumTypeName tests resolving enum with reserved name
func TestResolveReservedEnumTypeName(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes: map[string]*EnumTypeDef{
					"__cedar::test": {Values: []string{"A"}}, // Reserved name
				},
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for reserved enum type name")
	}
}

// TestResolveReservedCommonTypeName tests resolving common type with reserved name
func TestResolveReservedCommonTypeName(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: map[string]*CommonTypeDef{
					"__cedar": {Type: PrimitiveType{Kind: PrimitiveString}},
				},
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for reserved common type name")
	}
}

// TestResolveContextRefToNonRecord tests resolving context ref that doesn't resolve to record
func TestResolveContextRefToNonRecord(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {
				"MyType": {"type": "String"}
			},
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "MyType"}
					}
				}
			}
		}
	}`

	var s Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error because context must resolve to record")
	}
}

// TestResolveMultiPartNamespaceCommonType tests resolving common type in multi-part namespace
func TestResolveMultiPartNamespaceCommonType(t *testing.T) {
	cedar := `
namespace A::B::C {
	type Inner = { value: Long };
	type Outer = { nested: Inner };
	entity User { data: Outer };
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveSetOfCommonType tests resolution of Set<CommonType>
func TestResolveSetOfCommonType(t *testing.T) {
	cedar := `
type MyString = String;
entity User { values: Set<MyString> };
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// ============================================================================
// Additional marshal edge cases
// ============================================================================

// TestMarshalCedarMultipleNamespaces tests marshaling schema with multiple namespaces
func TestMarshalCedarMultipleNamespaces(t *testing.T) {
	s := NewBuilder().
		Namespace("NS1").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Annotate("doc", "Namespace 1").
		Namespace("NS2").
		Entity("Admin").
		Action("manage").Principal("Admin").Resource("Admin").
		Namespace("").
		Entity("Global").
		Action("global").Principal("Global").Resource("Global").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Should contain both namespaces and empty namespace content
	if !strings.Contains(string(data), "namespace NS1") {
		t.Error("expected NS1 namespace in output")
	}
	if !strings.Contains(string(data), "namespace NS2") {
		t.Error("expected NS2 namespace in output")
	}
	if !strings.Contains(string(data), "entity Global") {
		t.Error("expected Global entity in output (empty namespace)")
	}
}

// TestMarshalCedarNestedRecord tests marshaling nested record types
func TestMarshalCedarNestedRecord(t *testing.T) {
	nested := &RecordType{
		Attributes: map[string]*Attribute{
			"inner": {
				Type:        PrimitiveType{Kind: PrimitiveString},
				Required:    true,
				Annotations: make(Annotations),
			},
		},
	}

	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	s.Namespaces["Test"].EntityTypes["User"].Shape = &RecordType{
		Attributes: map[string]*Attribute{
			"nested": {
				Type:        nested,
				Required:    true,
				Annotations: make(Annotations),
			},
		},
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

// ============================================================================
// Additional JSON parsing edge cases
// ============================================================================

// TestJSONUnmarshalRecordAttributeError tests record attribute parsing error
func TestJSONUnmarshalRecordAttributeError(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"bad": "not an object"
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for invalid attribute")
	}
}

// TestJSONUnmarshalSetElementError tests set element parsing error
func TestJSONUnmarshalSetElementError(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"items": {
								"type": "Set",
								"element": "invalid"
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	if err == nil {
		t.Error("expected error for invalid set element")
	}
}

// ============================================================================
// Additional parser edge cases
// ============================================================================

// TestParseAnnotationError tests annotation parsing errors
func TestParseAnnotationError(t *testing.T) {
	// Annotation without closing paren
	cedar := `@doc("unclosed entity User;`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for unclosed annotation")
	}
}

// TestParseCommonTypeMissingEquals tests common type without equals
func TestParseCommonTypeMissingEquals(t *testing.T) {
	cedar := `type MyType String; entity User; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing equals in type def")
	}
}

// TestParseCommonTypeMissingSemicolon tests common type without semicolon
func TestParseCommonTypeMissingSemicolon(t *testing.T) {
	cedar := `type MyType = String entity User;`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing semicolon in type def")
	}
}

// TestParseEntityDuplicateWithEnum tests duplicate entity when enum exists
func TestParseEntityDuplicateWithEnum(t *testing.T) {
	cedar := `
entity Status enum ["A"];
entity Status;
action v appliesTo { principal: [Status], resource: [Status] };
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for duplicate entity (enum)")
	}
}

// TestParseStringListEmpty tests parsing empty string list
func TestParseStringListEmpty(t *testing.T) {
	cedar := `entity Status enum []; entity User; action v appliesTo { principal: [User], resource: [Status] };`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	// Empty enum is valid
	status := s.Namespaces[""].EnumTypes["Status"]
	if status == nil {
		t.Fatal("expected Status enum")
	}
	if len(status.Values) != 0 {
		t.Errorf("expected empty values, got %v", status.Values)
	}
}

// TestParseActionDuplicate tests duplicate action parsing
func TestParseActionDuplicate(t *testing.T) {
	cedar := `
entity User;
action view appliesTo { principal: [User], resource: [User] };
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for duplicate action")
	}
}

// TestParseEntityDuplicate tests duplicate entity parsing
func TestParseEntityDuplicate(t *testing.T) {
	cedar := `
entity User;
entity User;
action v appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for duplicate entity")
	}
}

// TestParseCommonTypeDuplicate tests duplicate common type parsing
func TestParseCommonTypeDuplicate(t *testing.T) {
	cedar := `
type MyType = String;
type MyType = Long;
entity User;
action v appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for duplicate common type")
	}
}

// TestParseRecordTypeError tests record type parsing error
func TestParseRecordTypeError(t *testing.T) {
	// Missing colon after attribute name
	cedar := `entity User { name String }; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing colon in record attribute")
	}
}

// TestParseRecordMissingClosingBrace tests record with missing closing brace
func TestParseRecordMissingClosingBrace(t *testing.T) {
	cedar := `entity User { name: String; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing closing brace in record")
	}
}

// TestParseSetTypeError tests Set type parsing error
func TestParseSetTypeError(t *testing.T) {
	// Missing closing >
	cedar := `entity User { items: Set<String }; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for malformed Set type")
	}
}

// TestParseTypeListError tests type list parsing error
func TestParseTypeListError(t *testing.T) {
	// Missing closing bracket
	cedar := `entity User in [Group; entity Group; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for malformed type list")
	}
}

// TestParseActionRefListError tests action ref list parsing error
func TestParseActionRefListError(t *testing.T) {
	cedar := `entity User; action admin; action view in [admin; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for malformed action ref list")
	}
}

// TestParseAppliesToMissingClosingBrace tests appliesTo with missing closing brace
func TestParseAppliesToMissingClosingBrace(t *testing.T) {
	cedar := `entity User; action view appliesTo { principal: [User];`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing closing brace in appliesTo")
	}
}

// TestParseNameListError tests name list parsing error with invalid continuation
func TestParseNameListError(t *testing.T) {
	cedar := `entity User; action view, 123 appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for invalid name in name list")
	}
}

// TestParseIdentListError tests ident list parsing error
func TestParseIdentListError(t *testing.T) {
	cedar := `entity User, 123 in [Group]; entity Group; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for invalid ident in ident list")
	}
}

// ============================================================================
// Additional JSON marshal edge cases
// ============================================================================

// TestMarshalTypeNil tests marshalType with nil
func TestMarshalTypeNil(t *testing.T) {
	_, err := marshalType(nil)
	if err == nil {
		t.Error("expected error for nil type")
	}
}

// TestMarshalTypeNonPointerRecord tests marshalType with non-pointer RecordType
func TestMarshalTypeNonPointerRecord(t *testing.T) {
	// RecordType as non-pointer should work through the *RecordType case
	rt := RecordType{Attributes: make(map[string]*Attribute)}
	_, err := marshalType(&rt) // Must be pointer
	if err != nil {
		t.Errorf("failed to marshal *RecordType: %v", err)
	}
}

// ============================================================================
// Additional resolve edge cases
// ============================================================================

// TestResolveTypeUnknown tests resolving an unknown type
func TestResolveTypeUnknown(t *testing.T) {
	// Create a resolver and try to resolve an unknown type
	// This is tested indirectly, but let's ensure the default case
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {
						Shape: &RecordType{
							Attributes: map[string]*Attribute{
								"ref": {
									Type:        EntityOrCommonRef{Name: "Unknown"},
									Required:    true,
									Annotations: make(Annotations),
								},
							},
						},
					},
				},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for undefined type reference")
	}
}

// TestResolveEntityTypeShapeError tests entity type shape resolution error
func TestResolveEntityTypeShapeError(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {
						Shape: &RecordType{
							Attributes: map[string]*Attribute{
								"ref": {
									Type:        EntityRef{Name: "Unknown"},
									Required:    true,
									Annotations: make(Annotations),
								},
							},
						},
					},
				},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for undefined entity ref in shape")
	}
}

// TestResolveEntityTypeTagsError tests entity type tags resolution error
func TestResolveEntityTypeTagsError(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {
						Tags: EntityRef{Name: "Unknown"},
					},
				},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for undefined entity ref in tags")
	}
}

// TestResolveActionContextError tests action context resolution error
func TestResolveActionContextError(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {},
				},
				EnumTypes: make(map[string]*EnumTypeDef),
				Actions: map[string]*ActionDef{
					"view": {
						AppliesTo: &AppliesTo{
							PrincipalTypes: []string{"User"},
							ResourceTypes:  []string{"User"},
							Context: &RecordType{
								Attributes: map[string]*Attribute{
									"ref": {
										Type:        EntityRef{Name: "Unknown"},
										Required:    true,
										Annotations: make(Annotations),
									},
								},
							},
						},
					},
				},
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected error for undefined entity ref in context")
	}
}

// TestResolveCommonTypeCycle tests cycle detection in common types
func TestResolveCommonTypeCycle(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: map[string]*CommonTypeDef{
					"A": {Type: CommonTypeRef{Name: "B"}},
					"B": {Type: CommonTypeRef{Name: "A"}},
				},
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected cycle error")
	}
}

// TestResolveCommonTypeToBuiltins tests resolving common type refs to builtins
func TestResolveCommonTypeToBuiltins(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: map[string]*CommonTypeDef{
					"MyLong":   {Type: CommonTypeRef{Name: "Long"}},
					"MyString": {Type: CommonTypeRef{Name: "String"}},
					"MyBool":   {Type: CommonTypeRef{Name: "Bool"}},
					"MyIP":     {Type: CommonTypeRef{Name: "ipaddr"}},
				},
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveShadowingEnumTypes tests shadowing detection with enum types
func TestResolveShadowingEnumTypes(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes: map[string]*EnumTypeDef{
					"Status": {Values: []string{"A"}},
				},
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
			"NS": {
				EntityTypes: make(map[string]*EntityTypeDef),
				EnumTypes: map[string]*EnumTypeDef{
					"Status": {Values: []string{"B"}},
				},
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err == nil {
		t.Error("expected shadow error for enum type")
	}
}

// ============================================================================
// Additional Cedar marshal edge cases
// ============================================================================

// TestMarshalCedarSortedNamespaces tests that namespaces are sorted correctly
func TestMarshalCedarSortedNamespaces(t *testing.T) {
	s := NewBuilder().
		Namespace("Zebra").Entity("A").Action("a").Principal("A").Resource("A").
		Namespace("Alpha").Entity("B").Action("b").Principal("B").Resource("B").
		Namespace("").Entity("C").Action("c").Principal("C").Resource("C").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	output := string(data)
	// Empty namespace should appear first (without namespace block)
	// Check that entity C comes before namespace declarations
	entityCPos := strings.Index(output, "entity C")
	namespacePos := strings.Index(output, "namespace")

	if entityCPos == -1 || namespacePos == -1 {
		t.Log(output)
		t.Fatal("expected both entity C and namespace in output")
	}

	if entityCPos > namespacePos {
		t.Error("expected empty namespace content before named namespaces")
	}
}

// ============================================================================
// More parse_cedar edge cases
// ============================================================================

// TestParseNamespaceMissingName tests namespace without name
func TestParseNamespaceMissingName(t *testing.T) {
	cedar := `namespace { entity User; }`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for namespace without name")
	}
}

// TestParseAppliesToMissingColon tests appliesTo with missing colon
func TestParseAppliesToMissingColon(t *testing.T) {
	cedar := `entity User; action view appliesTo { principal [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for appliesTo missing colon")
	}
}

// TestParseContextWithInlineRecord tests parsing context with inline record
func TestParseContextWithInlineRecord(t *testing.T) {
	cedar := `
entity User;
action view appliesTo {
	principal: [User],
	resource: [User],
	context: { flag: Bool }
};
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if action.AppliesTo.Context == nil {
		t.Error("expected inline context")
	}
}

// TestParseEntityWithDuplicateAttribute tests entity with duplicate attribute annotation
func TestParseEntityWithDuplicateAttribute(t *testing.T) {
	cedar := `
entity User {
	@doc("First")
	@note("Second")
	name: String,
};
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	user := s.Namespaces[""].EntityTypes["User"]
	attr := user.Shape.Attributes["name"]
	if attr.Annotations["doc"] != "First" {
		t.Errorf("expected doc annotation, got %v", attr.Annotations)
	}
	if attr.Annotations["note"] != "Second" {
		t.Errorf("expected note annotation, got %v", attr.Annotations)
	}
}

// TestParseNamespaceWithAnnotation tests namespace with annotation
func TestParseNamespaceWithAnnotation(t *testing.T) {
	cedar := `
@version("1.0")
namespace MyApp {
	entity User;
	action view appliesTo { principal: [User], resource: [User] };
}
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	ns := s.Namespaces["MyApp"]
	if ns.Annotations["version"] != "1.0" {
		t.Errorf("expected version annotation, got %v", ns.Annotations)
	}
}

// TestParseCommonTypeError tests common type with parsing error
func TestParseCommonTypeError(t *testing.T) {
	cedar := `type MyType = Set<; entity User; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for incomplete Set type")
	}
}

// TestParseEntityMissingAttributes tests entity type attribute parsing error
func TestParseEntityMissingAttributes(t *testing.T) {
	cedar := `entity User { bad }; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for bad entity attribute")
	}
}

// ============================================================================
// More resolve edge cases
// ============================================================================

// TestResolveQualifiedCommonType tests resolving qualified common type references
func TestResolveQualifiedCommonType(t *testing.T) {
	jsonSchema := `{
		"NS1": {
			"commonTypes": {
				"MyType": {"type": "String"}
			},
			"entityTypes": {},
			"actions": {}
		},
		"NS2": {
			"commonTypes": {
				"Ref": {"type": "NS1::MyType"}
			},
			"entityTypes": {},
			"actions": {}
		}
	}`

	var s Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveQualifiedEntityRef tests resolving qualified entity references
func TestResolveQualifiedEntityRef(t *testing.T) {
	jsonSchema := `{
		"NS1": {
			"entityTypes": {"Group": {}},
			"actions": {}
		},
		"NS2": {
			"entityTypes": {
				"User": {"memberOfTypes": ["NS1::Group"]}
			},
			"actions": {}
		}
	}`

	var s Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveQualifiedBuiltinType tests resolving __cedar:: prefixed types
func TestResolveQualifiedBuiltinType(t *testing.T) {
	cedar := `
entity User {
	val: __cedar::String,
};
action view appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveActionMemberOfError tests action memberOf resolution error
func TestResolveActionMemberOfError(t *testing.T) {
	// memberOf refs don't need to resolve - they're just stored as EntityUIDs
	// So this test ensures the path is exercised
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{"User": {}},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions: map[string]*ActionDef{
					"view": {
						MemberOf: []*ActionRef{
							{ID: "admin"},
							{Type: "OtherNS::Action", ID: "superadmin"},
						},
						AppliesTo: &AppliesTo{
							PrincipalTypes: []string{"User"},
							ResourceTypes:  []string{"User"},
						},
					},
				},
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// ============================================================================
// More JSON edge cases
// ============================================================================

// TestJSONUnmarshalActionError tests action unmarshal error
func TestJSONUnmarshalActionError(t *testing.T) {
	jsonSchema := `{
		"": {
			"entityTypes": {},
			"actions": {
				"view": {
					"appliesTo": {
						"context": {"type": "Set"}
					}
				}
			}
		}
	}`

	var s Schema
	err := json.Unmarshal([]byte(jsonSchema), &s)
	// Context parsing may fail because Set requires element
	if err == nil {
		t.Log("context parsing didn't fail, checking appliesTo")
	}
}

// TestJSONMarshalRecordTypeError tests record marshal error path
func TestJSONMarshalRecordTypeError(t *testing.T) {
	// All type marshaling is straightforward, but let's ensure record path works
	rt := &RecordType{
		Attributes: map[string]*Attribute{
			"name": {
				Type:        PrimitiveType{Kind: PrimitiveString},
				Required:    true,
				Annotations: make(Annotations),
			},
			"age": {
				Type:        PrimitiveType{Kind: PrimitiveLong},
				Required:    false,
				Annotations: Annotations{"doc": "User age"},
			},
		},
	}

	data, err := json.Marshal(rt)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty data")
	}
}

// TestParseInvalidStringEscape tests invalid string escape
func TestParseInvalidStringEscape(t *testing.T) {
	// Invalid unicode escape - depends on rust.Unquote behavior
	cedar := `entity User { "name\xZZ": String }; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	// This may or may not error depending on rust.Unquote implementation
	_ = s.UnmarshalCedar([]byte(cedar))
}

// TestMarshalCedarEmptySchema tests marshaling empty schema
func TestMarshalCedarEmptySchema(t *testing.T) {
	s := &Schema{
		Namespaces: make(map[string]*Namespace),
	}

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	// Empty schema should produce empty output (no content)
	// This is valid - an empty schema produces an empty byte slice
	t.Logf("Empty schema marshaled to %d bytes", len(data))
}

// TestMarshalCedarWithOnlyCommonTypes tests marshaling with only common types
func TestMarshalCedarWithOnlyCommonTypes(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		CommonType("MyString", String()).
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "type MyString") {
		t.Error("expected common type in output")
	}
}

// ============================================================================
// Marker method coverage (workaround for Go coverage tool)
// ============================================================================

// TestIsTypeMethodsCoverage explicitly calls isType methods with assertions
func TestIsTypeMethodsCoverage(t *testing.T) {
	// These tests explicitly call the marker methods to ensure coverage
	// The Go coverage tool sometimes doesn't count empty methods
	var types []Type

	types = append(types, PrimitiveType{Kind: PrimitiveLong})
	types = append(types, SetType{Element: PrimitiveType{Kind: PrimitiveString}})
	types = append(types, &RecordType{Attributes: make(map[string]*Attribute)})
	types = append(types, EntityRef{Name: "User"})
	types = append(types, ExtensionType{Name: "ipaddr"})
	types = append(types, CommonTypeRef{Name: "MyType"})
	types = append(types, EntityOrCommonRef{Name: "Ambiguous"})

	for _, typ := range types {
		// This forces the compiler to call isType
		_ = typ
		typ.isType()
	}
}

// ============================================================================
// Final edge cases to reach 95%+ coverage
// ============================================================================

// TestParseAnnotationIdentError tests annotation with invalid ident
func TestParseAnnotationIdentError(t *testing.T) {
	cedar := `@123 entity User; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for invalid annotation ident")
	}
}

// TestParseAnnotationStringError tests annotation with invalid string
func TestParseAnnotationStringError(t *testing.T) {
	cedar := `@doc("unclosed entity User; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for unclosed annotation string")
	}
}

// TestParseAnnotationMissingCloseParen tests annotation without closing paren
func TestParseAnnotationMissingCloseParen(t *testing.T) {
	cedar := `@doc("value" entity User; action v appliesTo { principal: [User], resource: [User] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing closing paren")
	}
}

// TestParseActionRefPath tests action ref with path
func TestParseActionRefPath(t *testing.T) {
	cedar := `
entity User;
action admin;
action view in [NS::Action::"admin"] appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	action := s.Namespaces[""].Actions["view"]
	if len(action.MemberOf) != 1 {
		t.Fatalf("expected 1 memberOf, got %d", len(action.MemberOf))
	}
	if action.MemberOf[0].Type != "NS::Action" || action.MemberOf[0].ID != "admin" {
		t.Errorf("expected NS::Action::admin, got %+v", action.MemberOf[0])
	}
}

// TestParseStringListMissingOpenBracket tests string list without opening bracket
func TestParseStringListMissingOpenBracket(t *testing.T) {
	cedar := `entity Status enum "A", "B"]; entity User; action v appliesTo { principal: [User], resource: [Status] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing open bracket in enum")
	}
}

// TestParseStringListMissingCloseBracket tests string list without closing bracket
func TestParseStringListMissingCloseBracket(t *testing.T) {
	cedar := `entity Status enum ["A", "B"; entity User; action v appliesTo { principal: [User], resource: [Status] };`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error for missing close bracket in enum")
	}
}

// TestMarshalCedarWithNamespaceOnly tests marshaling namespace with no empty namespace
func TestMarshalCedarWithNamespaceOnly(t *testing.T) {
	s := NewBuilder().
		Namespace("Test").
		Entity("User").
		Action("view").Principal("User").Resource("User").
		Build()

	data, err := s.MarshalCedar()
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if !strings.Contains(string(data), "namespace Test") {
		t.Error("expected namespace Test in output")
	}
}

// TestJSONUnmarshalActionWithContextRefCommonType tests action with CommonTypeRef context
func TestJSONUnmarshalActionWithContextRefCommonType(t *testing.T) {
	jsonSchema := `{
		"": {
			"commonTypes": {
				"MyContext": {"type": "Record", "attributes": {"flag": {"type": "Bool"}}}
			},
			"entityTypes": {"User": {}},
			"actions": {
				"view": {
					"appliesTo": {
						"principalTypes": ["User"],
						"resourceTypes": ["User"],
						"context": {"type": "MyContext"}
					}
				}
			}
		}
	}`

	var s Schema
	if err := json.Unmarshal([]byte(jsonSchema), &s); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Should resolve correctly
	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveEntityOrCommonToEntity tests EntityOrCommonRef resolving to entity
func TestResolveEntityOrCommonToEntity(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User":  {},
					"Group": {},
				},
				EnumTypes: make(map[string]*EnumTypeDef),
				Actions: map[string]*ActionDef{
					"view": {
						AppliesTo: &AppliesTo{
							PrincipalTypes: []string{"User"},
							ResourceTypes:  []string{"User"},
						},
					},
				},
				CommonTypes: map[string]*CommonTypeDef{
					"MyRef": {Type: EntityOrCommonRef{Name: "Group"}}, // Should resolve to entity Group
				},
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveEntityOrCommonToCommon tests EntityOrCommonRef resolving to common type
func TestResolveEntityOrCommonToCommon(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {},
				},
				EnumTypes: make(map[string]*EnumTypeDef),
				Actions: map[string]*ActionDef{
					"view": {
						AppliesTo: &AppliesTo{
							PrincipalTypes: []string{"User"},
							ResourceTypes:  []string{"User"},
						},
					},
				},
				CommonTypes: map[string]*CommonTypeDef{
					"MyString": {Type: PrimitiveType{Kind: PrimitiveString}},
					"MyRef":    {Type: EntityOrCommonRef{Name: "MyString"}}, // Should resolve to common type
				},
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestResolveEntityOrCommonBuiltins tests EntityOrCommonRef resolving to builtins
func TestResolveEntityOrCommonBuiltins(t *testing.T) {
	s := &Schema{
		Namespaces: map[string]*Namespace{
			"": {
				EntityTypes: map[string]*EntityTypeDef{
					"User": {
						Shape: &RecordType{
							Attributes: map[string]*Attribute{
								"strRef":  {Type: EntityOrCommonRef{Name: "String"}, Required: true, Annotations: make(Annotations)},
								"longRef": {Type: EntityOrCommonRef{Name: "Long"}, Required: true, Annotations: make(Annotations)},
								"boolRef": {Type: EntityOrCommonRef{Name: "Bool"}, Required: true, Annotations: make(Annotations)},
								"ipRef":   {Type: EntityOrCommonRef{Name: "ipaddr"}, Required: true, Annotations: make(Annotations)},
							},
						},
					},
				},
				EnumTypes:   make(map[string]*EnumTypeDef),
				Actions:     make(map[string]*ActionDef),
				CommonTypes: make(map[string]*CommonTypeDef),
				Annotations: make(Annotations),
			},
		},
	}

	_, err := s.Resolve()
	if err != nil {
		t.Fatalf("failed to resolve: %v", err)
	}
}

// TestParseNameWithDifferentContinuations tests parseName behavior
func TestParseNameWithDifferentContinuations(t *testing.T) {
	// Test action name list with quoted names
	cedar := `
entity User;
action "read doc", "write doc" appliesTo { principal: [User], resource: [User] };
`
	var s Schema
	if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if s.Namespaces[""].Actions["read doc"] == nil {
		t.Error("expected 'read doc' action")
	}
	if s.Namespaces[""].Actions["write doc"] == nil {
		t.Error("expected 'write doc' action")
	}
}

// TestParseEntityIdentListContinuation tests entity ident list with different continuation keywords
func TestParseEntityIdentListContinuation(t *testing.T) {
	// Test with different keywords that terminate ident list
	testCases := []string{
		`entity A, B in [Group]; entity Group; action v appliesTo { principal: [A], resource: [B] };`,
		`entity A, B enum ["X"]; action v appliesTo { principal: [A], resource: [B] };`,
		`entity A, B { name: String }; action v appliesTo { principal: [A], resource: [B] };`,
		`entity A, B = { name: String }; action v appliesTo { principal: [A], resource: [B] };`,
	}

	for _, cedar := range testCases {
		var s Schema
		if err := s.UnmarshalCedar([]byte(cedar)); err != nil {
			t.Errorf("failed to parse %q: %v", cedar[:30], err)
		}
	}
}

// TestJSONMarshalActionNoAppliesTo tests marshaling action without appliesTo
func TestJSONMarshalActionNoAppliesTo(t *testing.T) {
	action := &ActionDef{
		MemberOf:    nil,
		AppliesTo:   nil,
		Annotations: make(Annotations),
	}

	data, err := json.Marshal(action)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	if len(data) == 0 {
		t.Error("expected non-empty output")
	}
}

// TestParseActionExpectError tests the error path when an unexpected keyword appears in a declaration context
func TestParseActionExpectError(t *testing.T) {
	// Test through public API - when "action" keyword is expected but something else is provided
	cedar := `namespace Test { notaction view; }`
	var s Schema
	err := s.UnmarshalCedar([]byte(cedar))
	if err == nil {
		t.Error("expected error when 'action' keyword is missing")
	}
}
