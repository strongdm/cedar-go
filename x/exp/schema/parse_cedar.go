package schema

import (
	"bytes"
	"fmt"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/cedar-policy/cedar-go/x/exp/schema/internal/parse"
)

// UnmarshalCedar parses a Cedar schema from Cedar text format.
func (s *Schema) UnmarshalCedar(data []byte) error {
	p := parse.New(data, s.filename)

	parsed, err := p.Parse()
	if err != nil {
		// Convert internal parse error to schema parse error
		if pe, ok := err.(*parse.ParseError); ok {
			return &ParseError{
				Filename: pe.Filename,
				Line:     pe.Line,
				Column:   pe.Column,
				Message:  pe.Message,
			}
		}
		if re, ok := err.(*parse.ReservedNameError); ok {
			return &ReservedNameError{
				Name: re.Name,
				Kind: re.Kind,
			}
		}
		return err
	}

	s.Namespaces = convertNamespaces(parsed.Namespaces)
	return nil
}

func convertNamespaces(internal map[string]*parse.Namespace) map[string]*Namespace {
	result := make(map[string]*Namespace, len(internal))
	for name, ns := range internal {
		result[name] = convertNamespace(ns)
	}
	return result
}

func convertNamespace(internal *parse.Namespace) *Namespace {
	ns := &Namespace{
		EntityTypes: make(map[string]*EntityTypeDef),
		EnumTypes:   make(map[string]*EnumTypeDef),
		Actions:     make(map[string]*ActionDef),
		CommonTypes: make(map[string]*CommonTypeDef),
		Annotations: Annotations(internal.Annotations),
	}

	for name, et := range internal.EntityTypes {
		ns.EntityTypes[name] = convertEntityTypeDef(et)
	}

	for name, enum := range internal.EnumTypes {
		ns.EnumTypes[name] = convertEnumTypeDef(enum)
	}

	for name, act := range internal.Actions {
		ns.Actions[name] = convertActionDef(act)
	}

	for name, ct := range internal.CommonTypes {
		ns.CommonTypes[name] = convertCommonTypeDef(ct)
	}

	return ns
}

func convertEntityTypeDef(internal *parse.EntityTypeDef) *EntityTypeDef {
	et := &EntityTypeDef{
		MemberOfTypes: internal.MemberOfTypes,
		Annotations:   Annotations(internal.Annotations),
	}

	if internal.Shape != nil {
		et.Shape = convertRecordType(internal.Shape)
	}

	if internal.Tags != nil {
		et.Tags = convertType(internal.Tags)
	}

	return et
}

func convertEnumTypeDef(internal *parse.EnumTypeDef) *EnumTypeDef {
	return &EnumTypeDef{
		Values:      internal.Values,
		Annotations: Annotations(internal.Annotations),
	}
}

func convertActionDef(internal *parse.ActionDef) *ActionDef {
	act := &ActionDef{
		Annotations: Annotations(internal.Annotations),
	}

	if len(internal.MemberOf) > 0 {
		act.MemberOf = make([]*ActionRef, len(internal.MemberOf))
		for i, ref := range internal.MemberOf {
			act.MemberOf[i] = &ActionRef{
				Type: ref.Type,
				ID:   ref.ID,
			}
		}
	}

	if internal.AppliesTo != nil {
		act.AppliesTo = convertAppliesTo(internal.AppliesTo)
	}

	return act
}

func convertAppliesTo(internal *parse.AppliesTo) *AppliesTo {
	at := &AppliesTo{
		PrincipalTypes: internal.PrincipalTypes,
		ResourceTypes:  internal.ResourceTypes,
	}

	if internal.Context != nil {
		at.Context = convertRecordType(internal.Context)
	}

	if internal.ContextRef != nil {
		at.ContextRef = convertType(internal.ContextRef)
	}

	return at
}

func convertCommonTypeDef(internal *parse.CommonTypeDef) *CommonTypeDef {
	return &CommonTypeDef{
		Type:        convertType(internal.Type),
		Annotations: Annotations(internal.Annotations),
	}
}

func convertRecordType(internal *parse.RecordType) *RecordType {
	rt := &RecordType{
		Attributes: make(map[string]*Attribute, len(internal.Attributes)),
	}

	for name, attr := range internal.Attributes {
		rt.Attributes[name] = &Attribute{
			Type:        convertType(attr.Type),
			Required:    attr.Required,
			Annotations: Annotations(attr.Annotations),
		}
	}

	return rt
}

func convertType(internal parse.Type) Type {
	switch v := internal.(type) {
	case parse.PrimitiveType:
		return PrimitiveType{Kind: PrimitiveKind(v.Kind)}
	case parse.SetType:
		return SetType{Element: convertType(v.Element)}
	case *parse.RecordType:
		return convertRecordType(v)
	case parse.EntityRef:
		return EntityRef{Name: v.Name}
	case parse.ExtensionType:
		return ExtensionType{Name: v.Name}
	case parse.CommonTypeRef:
		return CommonTypeRef{Name: v.Name}
	case parse.EntityOrCommonRef:
		return EntityOrCommonRef{Name: v.Name}
	default:
		panic(fmt.Sprintf("convertType: unexpected type %T", internal))
	}
}

// MarshalCedar serializes the schema to Cedar text format.
// The output is deterministic: namespaces, types, actions, and attributes
// are sorted alphabetically, with the empty namespace written first.
func (s *Schema) MarshalCedar() ([]byte, error) {
	var buf bytes.Buffer

	// Sort namespaces for deterministic output (empty namespace first, then alphabetical)
	nsNames := slices.Collect(maps.Keys(s.Namespaces))
	slices.Sort(nsNames)

	// Write empty namespace declarations first (without namespace block)
	if ns, ok := s.Namespaces[""]; ok {
		if err := writeNamespaceContents(&buf, ns, ""); err != nil {
			return nil, err
		}
	}

	// Write named namespaces
	for _, name := range nsNames {
		if name == "" {
			continue // Already handled above
		}
		ns := s.Namespaces[name]

		// Write annotations
		writeAnnotations(&buf, "", ns.Annotations)

		fmt.Fprintf(&buf, "namespace %s {\n", name)
		if err := writeNamespaceContents(&buf, ns, "  "); err != nil {
			return nil, err
		}
		fmt.Fprintf(&buf, "}\n")
	}

	return buf.Bytes(), nil
}

// writeAnnotations writes annotation declarations to the buffer in sorted order.
func writeAnnotations(buf *bytes.Buffer, indent string, ann Annotations) {
	keys := slices.Collect(maps.Keys(ann))
	slices.Sort(keys)
	for _, k := range keys {
		fmt.Fprintf(buf, "%s@%s(%s)\n", indent, k, quoteString(ann[k]))
	}
}

func writeNamespaceContents(buf *bytes.Buffer, ns *Namespace, indent string) error {
	// Write entity types in sorted order
	entityNames := slices.Collect(maps.Keys(ns.EntityTypes))
	slices.Sort(entityNames)
	for _, name := range entityNames {
		writeEntityType(buf, name, ns.EntityTypes[name], indent)
	}

	// Write enum types in sorted order
	enumNames := slices.Collect(maps.Keys(ns.EnumTypes))
	slices.Sort(enumNames)
	for _, name := range enumNames {
		writeEnumType(buf, name, ns.EnumTypes[name], indent)
	}

	// Write actions in sorted order
	actionNames := slices.Collect(maps.Keys(ns.Actions))
	slices.Sort(actionNames)
	for _, name := range actionNames {
		writeAction(buf, name, ns.Actions[name], indent)
	}

	// Write common types in sorted order
	commonNames := slices.Collect(maps.Keys(ns.CommonTypes))
	slices.Sort(commonNames)
	for _, name := range commonNames {
		writeCommonType(buf, name, ns.CommonTypes[name], indent)
	}

	return nil
}

func writeEntityType(buf *bytes.Buffer, name string, et *EntityTypeDef, indent string) {
	writeAnnotations(buf, indent, et.Annotations)

	fmt.Fprintf(buf, "%sentity %s", indent, name)

	if len(et.MemberOfTypes) > 0 {
		fmt.Fprintf(buf, " in [%s]", strings.Join(et.MemberOfTypes, ", "))
	}

	if et.Shape != nil && len(et.Shape.Attributes) > 0 {
		buf.WriteString(" {\n")
		writeRecordAttributes(buf, et.Shape, indent+"  ")
		fmt.Fprintf(buf, "%s}", indent)
	}

	if et.Tags != nil {
		buf.WriteString(" tags ")
		writeType(buf, et.Tags)
	}

	buf.WriteString(";\n")
}

func writeEnumType(buf *bytes.Buffer, name string, enum *EnumTypeDef, indent string) {
	writeAnnotations(buf, indent, enum.Annotations)

	fmt.Fprintf(buf, "%sentity %s enum [", indent, name)
	for i, v := range enum.Values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(quoteString(v))
	}
	buf.WriteString("];\n")
}

func writeAction(buf *bytes.Buffer, name string, act *ActionDef, indent string) {
	writeAnnotations(buf, indent, act.Annotations)

	// Quote name if it's not a valid identifier
	actionName := name
	if !isValidIdent(name) {
		actionName = quoteString(name)
	}

	fmt.Fprintf(buf, "%saction %s", indent, actionName)

	if len(act.MemberOf) > 0 {
		buf.WriteString(" in [")
		for i, ref := range act.MemberOf {
			if i > 0 {
				buf.WriteString(", ")
			}
			if ref.Type != "" {
				fmt.Fprintf(buf, "%s::%s", ref.Type, quoteString(ref.ID))
			} else {
				if isValidIdent(ref.ID) {
					buf.WriteString(ref.ID)
				} else {
					buf.WriteString(quoteString(ref.ID))
				}
			}
		}
		buf.WriteString("]")
	}

	if act.AppliesTo != nil {
		buf.WriteString(" appliesTo {\n")

		if len(act.AppliesTo.PrincipalTypes) > 0 {
			fmt.Fprintf(buf, "%s  principal: [%s],\n", indent, strings.Join(act.AppliesTo.PrincipalTypes, ", "))
		}

		if len(act.AppliesTo.ResourceTypes) > 0 {
			fmt.Fprintf(buf, "%s  resource: [%s],\n", indent, strings.Join(act.AppliesTo.ResourceTypes, ", "))
		}

		if act.AppliesTo.ContextRef != nil {
			fmt.Fprintf(buf, "%s  context: ", indent)
			writeType(buf, act.AppliesTo.ContextRef)
			buf.WriteString(",\n")
		} else if act.AppliesTo.Context != nil && len(act.AppliesTo.Context.Attributes) > 0 {
			fmt.Fprintf(buf, "%s  context: {\n", indent)
			writeRecordAttributes(buf, act.AppliesTo.Context, indent+"    ")
			fmt.Fprintf(buf, "%s  },\n", indent)
		}

		fmt.Fprintf(buf, "%s}", indent)
	}

	buf.WriteString(";\n")
}

func writeCommonType(buf *bytes.Buffer, name string, ct *CommonTypeDef, indent string) {
	writeAnnotations(buf, indent, ct.Annotations)

	fmt.Fprintf(buf, "%stype %s = ", indent, name)
	writeType(buf, ct.Type)
	buf.WriteString(";\n")
}

func writeRecordAttributes(buf *bytes.Buffer, rt *RecordType, indent string) {
	// Sort attribute names for deterministic output
	attrNames := slices.Collect(maps.Keys(rt.Attributes))
	slices.Sort(attrNames)

	for _, name := range attrNames {
		attr := rt.Attributes[name]
		writeAnnotations(buf, indent, attr.Annotations)

		attrName := name
		if !isValidIdent(name) {
			attrName = quoteString(name)
		}

		fmt.Fprintf(buf, "%s%s", indent, attrName)
		if !attr.Required {
			buf.WriteString("?")
		}
		buf.WriteString(": ")
		writeType(buf, attr.Type)
		buf.WriteString(",\n")
	}
}

func writeType(buf *bytes.Buffer, t Type) {
	switch v := t.(type) {
	case PrimitiveType:
		buf.WriteString(v.Kind.String())
	case SetType:
		buf.WriteString("Set<")
		writeType(buf, v.Element)
		buf.WriteString(">")
	case *RecordType:
		buf.WriteString("{\n")
		writeRecordAttributes(buf, v, "  ")
		buf.WriteString("}")
	case EntityRef:
		buf.WriteString(v.Name)
	case ExtensionType:
		buf.WriteString(v.Name)
	case CommonTypeRef:
		buf.WriteString(v.Name)
	case EntityOrCommonRef:
		buf.WriteString(v.Name)
	}
}

func isValidIdent(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, r := range s {
		if i == 0 {
			if !unicode.IsLetter(r) && r != '_' {
				return false
			}
		} else {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				return false
			}
		}
	}
	return true
}

func quoteString(s string) string {
	return strconv.Quote(s)
}
