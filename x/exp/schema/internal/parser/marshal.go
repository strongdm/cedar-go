package parser

import (
	"bytes"
	"fmt"
	"sort"

	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

// marshalMapSorted extracts keys from a map, sorts them, and calls the marshal function for each key-value pair.
// It handles the "first" flag to add newlines between items.
func marshalMapSorted[K ~string, V any](
	buf *bytes.Buffer,
	m map[K]V,
	first *bool,
	marshalFunc func(*bytes.Buffer, K, V, string),
	indent string,
) {
	if len(m) == 0 {
		return
	}

	// Extract and sort keys
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)

	// Marshal each item
	for _, keyStr := range keys {
		if !*first {
			buf.WriteString("\n")
		}
		*first = false
		key := K(keyStr)
		marshalFunc(buf, key, m[key], indent)
	}
}

// MarshalSchema converts the schema to Cedar human-readable format.
func MarshalSchema(s *ast.Schema) []byte {
	var buf bytes.Buffer
	first := true

	// Marshal top-level declarations in order
	marshalMapSorted(&buf, s.CommonTypes, &first, marshalCommonType, "")
	marshalMapSorted(&buf, s.Entities, &first, marshalEntity, "")
	marshalMapSorted(&buf, s.Enums, &first, marshalEnum, "")
	marshalMapSorted(&buf, s.Actions, &first, marshalAction, "")
	marshalMapSorted(&buf, s.Namespaces, &first, marshalNamespace, "")

	return buf.Bytes()
}

func marshalAnnotations(buf *bytes.Buffer, annotations ast.Annotations, indent string) {
	// Sort annotation keys for consistent output
	keys := make([]string, 0, len(annotations))
	for key := range annotations {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)

	for _, key := range keys {
		buf.WriteString(indent)
		buf.WriteString("@")
		buf.WriteString(key)
		value := annotations[types.Ident(key)]
		if value != "" {
			buf.WriteString("(")
			buf.WriteString(quoteString(string(value)))
			buf.WriteString(")")
		}
		buf.WriteString("\n")
	}
}

func marshalNamespace(buf *bytes.Buffer, name types.Path, ns ast.Namespace, indent string) {
	marshalAnnotations(buf, ns.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("namespace ")
	buf.WriteString(string(name))
	buf.WriteString(" {\n")

	innerIndent := indent + "  "
	first := true

	// Marshal namespace declarations in order
	marshalMapSorted(buf, ns.CommonTypes, &first, marshalCommonType, innerIndent)
	marshalMapSorted(buf, ns.Entities, &first, marshalEntity, innerIndent)
	marshalMapSorted(buf, ns.Enums, &first, marshalEnum, innerIndent)
	marshalMapSorted(buf, ns.Actions, &first, marshalAction, innerIndent)

	buf.WriteString(indent)
	buf.WriteString("}\n")
}

func marshalCommonType(buf *bytes.Buffer, name types.Ident, ct ast.CommonType, indent string) {
	marshalAnnotations(buf, ct.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("type ")
	buf.WriteString(string(name))
	buf.WriteString(" = ")
	marshalTypeIndented(buf, ct.Type, indent)
	buf.WriteString(";\n")
}

func marshalEntity(buf *bytes.Buffer, name types.EntityType, e ast.Entity, indent string) {
	marshalAnnotations(buf, e.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("entity ")
	buf.WriteString(string(name))

	if len(e.MemberOf) > 0 {
		buf.WriteString(" in ")
		marshalEntityTypeRefs(buf, e.MemberOf)
	}

	if e.Shape != nil && len(e.Shape.Attributes) > 0 {
		buf.WriteString(" = ")
		marshalRecordType(buf, *e.Shape, indent)
	}

	if e.Tags != nil {
		buf.WriteString(" tags ")
		marshalTypeIndented(buf, e.Tags, indent+"  ")
	}

	buf.WriteString(";\n")
}

func marshalEnum(buf *bytes.Buffer, name types.EntityType, e ast.Enum, indent string) {
	marshalAnnotations(buf, e.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("entity ")
	buf.WriteString(string(name))
	buf.WriteString(" enum [")

	for i, v := range e.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(quoteString(string(v)))
	}

	buf.WriteString("];\n")
}

func marshalAction(buf *bytes.Buffer, name types.String, a ast.Action, indent string) {
	marshalAnnotations(buf, a.Annotations, indent)
	buf.WriteString(indent)
	buf.WriteString("action ")
	marshalActionName(buf, string(name))

	if len(a.MemberOf) > 0 {
		buf.WriteString(" in ")
		marshalEntityRefs(buf, a.MemberOf)
	}

	if a.AppliesTo != nil {
		buf.WriteString(" appliesTo {\n")
		innerIndent := indent + "  "

		if len(a.AppliesTo.PrincipalTypes) > 0 {
			buf.WriteString(innerIndent)
			buf.WriteString("principal: ")
			marshalEntityTypeRefs(buf, a.AppliesTo.PrincipalTypes)
			buf.WriteString(",\n")
		}

		if len(a.AppliesTo.ResourceTypes) > 0 {
			buf.WriteString(innerIndent)
			buf.WriteString("resource: ")
			marshalEntityTypeRefs(buf, a.AppliesTo.ResourceTypes)
			buf.WriteString(",\n")
		}

		if a.AppliesTo.Context != nil {
			buf.WriteString(innerIndent)
			buf.WriteString("context: ")
			marshalTypeIndented(buf, a.AppliesTo.Context, innerIndent)
			buf.WriteString("\n")
		}

		buf.WriteString(indent)
		buf.WriteString("}")
	}

	buf.WriteString(";\n")
}

func marshalActionName(buf *bytes.Buffer, name string) {
	if needsQuoting(name) {
		buf.WriteString(quoteString(name))
	} else {
		buf.WriteString(name)
	}
}

func marshalEntityTypeRefs(buf *bytes.Buffer, refs []ast.EntityTypeRef) {
	buf.WriteString("[")
	for i, ref := range refs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(string(ref.Name))
	}
	buf.WriteString("]")
}

func marshalEntityRefs(buf *bytes.Buffer, refs []ast.EntityRef) {
	if len(refs) == 1 {
		marshalEntityRef(buf, refs[0])
		return
	}

	buf.WriteString("[")
	for i, ref := range refs {
		if i > 0 {
			buf.WriteString(", ")
		}
		marshalEntityRef(buf, ref)
	}
	buf.WriteString("]")
}

func marshalEntityRef(buf *bytes.Buffer, ref ast.EntityRef) {
	if ref.Type.Name != "" {
		buf.WriteString(string(ref.Type.Name))
		buf.WriteString("::")
	}
	buf.WriteString(quoteString(string(ref.ID)))
}

func marshalTypeIndented(buf *bytes.Buffer, t ast.IsType, indent string) {
	switch v := t.(type) {
	case ast.StringType:
		buf.WriteString("String")
	case ast.LongType:
		buf.WriteString("Long")
	case ast.BoolType:
		buf.WriteString("Bool")
	case ast.ExtensionType:
		buf.WriteString("__cedar::")
		buf.WriteString(string(v.Name))
	case ast.SetType:
		buf.WriteString("Set<")
		marshalTypeIndented(buf, v.Element, indent+"  ")
		buf.WriteString(">")
	case ast.RecordType:
		marshalRecordType(buf, v, indent)
	case ast.EntityTypeRef:
		buf.WriteString(string(v.Name))
	case ast.TypeRef:
		buf.WriteString(string(v.Name))
	}
}

func marshalRecordType(buf *bytes.Buffer, r ast.RecordType, indent string) {
	buf.WriteString("{")
	if len(r.Attributes) == 0 {
		buf.WriteString("}")
		return
	}

	// Sort attribute keys for consistent output
	keys := make([]string, 0, len(r.Attributes))
	for key := range r.Attributes {
		keys = append(keys, string(key))
	}
	sort.Strings(keys)

	buf.WriteString("\n")
	innerIndent := indent + "  "
	for i, key := range keys {
		if i > 0 {
			buf.WriteString(",\n")
		}
		attr := r.Attributes[types.String(key)]
		marshalAnnotations(buf, attr.Annotations, innerIndent)
		buf.WriteString(innerIndent)
		buf.WriteString(quoteString(key))
		if attr.Optional {
			buf.WriteString("?")
		}
		buf.WriteString(": ")
		marshalTypeIndented(buf, attr.Type, innerIndent)
	}
	buf.WriteString(",\n")
	buf.WriteString(indent)
	buf.WriteString("}")
}

func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	// Check if it's a reserved keyword
	if s == "in" {
		return true
	}
	for i, c := range s {
		if i == 0 {
			if !isIdentStart(c) {
				return true
			}
		} else {
			if !isIdentChar(c) {
				return true
			}
		}
	}
	return false
}

func isIdentStart(c rune) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') || c == '_'
}

func isIdentChar(c rune) bool {
	return isIdentStart(c) || (c >= '0' && c <= '9')
}

func quoteString(s string) string {
	// Use Cedar-compatible escape sequences only
	var buf bytes.Buffer
	buf.WriteByte('"')
	for _, r := range s {
		switch r {
		case '\n':
			buf.WriteString("\\n")
		case '\r':
			buf.WriteString("\\r")
		case '\t':
			buf.WriteString("\\t")
		case '\\':
			buf.WriteString("\\\\")
		case '\x00':
			buf.WriteString("\\0")
		case '\'':
			buf.WriteString("\\'")
		case '"':
			buf.WriteString("\\\"")
		default:
			if r < 0x20 || r == 0x7F {
				// Control characters: use \xNN hex escape (2 hex digits)
				buf.WriteString(fmt.Sprintf("\\x%02x", r))
			} else if r > 0x7E && r < 0xA0 {
				// Extended control characters: use \xNN
				buf.WriteString(fmt.Sprintf("\\x%02x", r))
			} else {
				// Printable character
				buf.WriteRune(r)
			}
		}
	}
	buf.WriteByte('"')
	return buf.String()
}
