package schema

import (
	"bytes"

	"github.com/cedar-policy/cedar-go/types"
)

// Action defines an action in a Cedar schema.
type Action struct {
	Name        types.Ident // may be quoted
	Annotations Annotations
	MemberOf    []ActionRef  // action groups this action is a member of
	AppliesTo   *AppliesTo   // optional applies-to clause
}

// ActionRef is a reference to an action, optionally with a namespace.
type ActionRef struct {
	Namespace types.Path   // optional namespace
	Name      types.String // action name (always treated as a string)
}

// NewAction creates a new action definition with the given name.
func NewAction(name types.Ident) *Action {
	return &Action{Name: name}
}

// Annotate adds an annotation to the action.
func (a *Action) Annotate(key types.Ident, value types.String) *Action {
	a.Annotations = append(a.Annotations, Annotation{Key: key, Value: value})
	return a
}

// In specifies that this action is a member of the given action groups.
func (a *Action) In(refs ...ActionRef) *Action {
	a.MemberOf = append(a.MemberOf, refs...)
	return a
}

// InAction is a convenience method to add an unqualified action reference.
func (a *Action) InAction(name types.String) *Action {
	return a.In(ActionRef{Name: name})
}

// InQualifiedAction adds a qualified action reference.
func (a *Action) InQualifiedAction(namespace types.Path, name types.String) *Action {
	return a.In(ActionRef{Namespace: namespace, Name: name})
}

// WithAppliesTo sets the applies-to clause for this action.
func (a *Action) WithAppliesTo(appliesTo AppliesTo) *Action {
	a.AppliesTo = &appliesTo
	return a
}

// MarshalCedar returns the Cedar format representation of the action.
func (a *Action) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write annotations
	for _, ann := range a.Annotations {
		buf.Write(ann.MarshalCedar())
		buf.WriteByte('\n')
	}

	// Write action declaration
	buf.WriteString("action ")
	buf.WriteString(string(a.Name))

	// Write memberOf clause
	if len(a.MemberOf) > 0 {
		buf.WriteString(" in ")
		if len(a.MemberOf) == 1 {
			buf.Write(a.MemberOf[0].MarshalCedar())
		} else {
			buf.WriteByte('[')
			for i, ref := range a.MemberOf {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.Write(ref.MarshalCedar())
			}
			buf.WriteByte(']')
		}
	}

	// Write applies-to clause
	if a.AppliesTo != nil {
		buf.WriteString(" appliesTo {\n")
		buf.Write(a.AppliesTo.MarshalCedar())
		buf.WriteString("}")
	}

	buf.WriteByte(';')
	return buf.Bytes()
}

// MarshalCedar returns the Cedar format representation of the action reference.
func (r ActionRef) MarshalCedar() []byte {
	var buf bytes.Buffer
	if r.Namespace != "" {
		buf.WriteString(string(r.Namespace))
		buf.WriteString("::")
	}
	buf.Write(marshalString(r.Name))
	return buf.Bytes()
}

// AppliesTo defines what principals and resources an action can apply to.
type AppliesTo struct {
	PrincipalTypes []types.Path
	ResourceTypes  []types.Path
	Context        Type // optional context type
}

// NewAppliesTo creates a new applies-to clause.
func NewAppliesTo() AppliesTo {
	return AppliesTo{}
}

// WithPrincipals sets the principal types.
func (a AppliesTo) WithPrincipals(principals ...types.Path) AppliesTo {
	a.PrincipalTypes = append(a.PrincipalTypes, principals...)
	return a
}

// WithResources sets the resource types.
func (a AppliesTo) WithResources(resources ...types.Path) AppliesTo {
	a.ResourceTypes = append(a.ResourceTypes, resources...)
	return a
}

// WithContext sets the context type.
func (a AppliesTo) WithContext(ctx Type) AppliesTo {
	a.Context = ctx
	return a
}

// MarshalCedar returns the Cedar format representation of the applies-to clause.
func (a AppliesTo) MarshalCedar() []byte {
	var buf bytes.Buffer

	// Write principal types
	if len(a.PrincipalTypes) > 0 {
		buf.WriteString("  principal: ")
		if len(a.PrincipalTypes) == 1 {
			buf.WriteString(string(a.PrincipalTypes[0]))
		} else {
			buf.WriteByte('[')
			for i, p := range a.PrincipalTypes {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(string(p))
			}
			buf.WriteByte(']')
		}
		buf.WriteString(",\n")
	}

	// Write resource types
	if len(a.ResourceTypes) > 0 {
		buf.WriteString("  resource: ")
		if len(a.ResourceTypes) == 1 {
			buf.WriteString(string(a.ResourceTypes[0]))
		} else {
			buf.WriteByte('[')
			for i, r := range a.ResourceTypes {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(string(r))
			}
			buf.WriteByte(']')
		}
		buf.WriteString(",\n")
	}

	// Write context type
	if a.Context != nil {
		buf.WriteString("  context: ")
		// Indent the context type properly if it's a record
		contextBytes := a.Context.MarshalCedar()
		// Replace newlines with newline + indent for record types
		indented := bytes.ReplaceAll(contextBytes, []byte("\n"), []byte("\n  "))
		buf.Write(indented)
		buf.WriteString(",\n")
	}

	return buf.Bytes()
}
