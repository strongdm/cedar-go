package schema

// Action represents a Cedar action definition.
// Actions define operations that can be performed, along with constraints on
// which principals and resources they apply to.
type Action struct {
	name        string
	namespace   *Namespace
	memberOf    []*ActionRef
	appliesTo   *AppliesTo
	annotations map[string]string
}

// ActionRef represents a reference to an action group.
type ActionRef struct {
	id       string
	typeName string // optional type qualifier
}

// AppliesTo defines constraints on an action.
type AppliesTo struct {
	principals []string
	resources  []string
	context    Type
}

// NewAction creates a new action with the given name.
// The name can be a simple identifier (e.g., "viewPhoto") or a quoted string.
func NewAction(name string) *Action {
	return &Action{
		name: name,
	}
}

// MemberOf specifies that this action is a member of the given action groups.
// Each parent can be a simple action name or a fully-qualified reference.
func (a *Action) MemberOf(parents ...*ActionRef) *Action {
	a.memberOf = append(a.memberOf, parents...)
	return a
}

// AppliesTo sets the principals, resources, and optional context for this action.
func (a *Action) AppliesTo(principals, resources []string, context Type) *Action {
	a.appliesTo = &AppliesTo{
		principals: principals,
		resources:  resources,
		context:    context,
	}
	return a
}

// WithAnnotation adds an annotation to the action.
func (a *Action) WithAnnotation(key, value string) *Action {
	if a.annotations == nil {
		a.annotations = make(map[string]string)
	}
	a.annotations[key] = value
	return a
}

// addToNamespace implements the Declaration interface.
func (a *Action) addToNamespace(ns *Namespace) {
	a.namespace = ns
	if ns.actions == nil {
		ns.actions = make(map[string]*Action)
	}
	ns.actions[a.name] = a
}

// ActionGroup creates a reference to an action group.
// Use this with MemberOf to specify action hierarchies.
func ActionGroup(id string) *ActionRef {
	return &ActionRef{
		id: id,
	}
}

// QualifiedActionGroup creates a fully-qualified reference to an action group.
// The typeName should be the namespace-qualified action type (e.g., "PhotoApp::Action").
func QualifiedActionGroup(id, typeName string) *ActionRef {
	return &ActionRef{
		id:       id,
		typeName: typeName,
	}
}

// Principals creates a list of principal type names for use with AppliesTo.
func Principals(types ...string) []string {
	return types
}

// Resources creates a list of resource type names for use with AppliesTo.
func Resources(types ...string) []string {
	return types
}
