package batch

import (
	"context"
	"fmt"
	"maps"
	"slices"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/ast"
	"github.com/cedar-policy/cedar-go/internal/consts"
	"github.com/cedar-policy/cedar-go/internal/eval"
	"github.com/cedar-policy/cedar-go/types"
)

// Ignore returns a value that should be ignored during batch evaluation.
func Ignore() types.Value { return eval.Ignore() }

// Variable returns a named variable that is populated during batch evaluation.
func Variable(name types.String) types.Value { return eval.Variable(name) }

// Request defines the PARC and map of Variables to batch evaluate.
type Request struct {
	Principal types.Value
	Action    types.Value
	Resource  types.Value
	Context   types.Value
	Variables Variables
}

// Variables is a map of String to slice of Value.
type Variables map[types.String][]types.Value

// Values is a map of String to Value.  This structure is part of the result and
// reveals the current variable substitutions.
type Values map[types.String]types.Value

// Result is the result of a single batched authorization.  It includes a
// specific Request, the Values that were substituted, and the resulting
// Decision and Diagnostics.
type Result struct {
	Request    types.Request
	Values     Values
	Decision   types.Decision
	Diagnostic types.Diagnostic
}

// Callback is a function that is called for each single batch authorization with
// a Result.
type Callback func(Result)

// Option is an option to be passed to the Authorize function.  It should be created
// by one of the WithOption style factories.
type Option struct {
	ignoreForbid bool
	ignorePermit bool
}

// WithIgnoreForbid set the behavior of ignore to bias towards a deny decision.
// e.g. it best answers the question "When ignoring context could this request be denied?"
//
//  1. When a Permit Policy Condition refers to an ignored value, the Condition is dropped from the Policy.
//  2. When a Forbid Policy Condition refers to an ignored value, the Policy is dropped.
//  3. When a Scope clause refers to an ignored value, that scope clause is set to match any.
func WithIgnoreForbid() Option {
	return Option{ignoreForbid: true}
}

// WithIgnoreForbid set the behavior of ignore to bias towards an allow decision.
// e.g. it better answers the question "When ignoring context could this request be allowed?"
//
//  1. When a Forbid Policy Condition refers to an ignored value, the Condition is dropped from the Policy.
//  2. When a Permit Policy Condition refers to an ignored value, the Policy is dropped.
//  3. When a Scope clause refers to an ignored value, that scope clause is set to match any.
//
// This is the default behavior.
func WithIgnorePermit() Option {
	return Option{ignorePermit: true}
}

type idEvaler struct {
	PolicyID types.PolicyID
	Policy   *ast.Policy
	Evaler   eval.BoolEvaler
}

type idPolicy struct {
	PolicyID types.PolicyID
	Policy   *ast.Policy
}

type batchEvaler struct {
	Variables  []variableItem
	Values     Values
	ignoreBias types.Effect

	policies []idPolicy
	compiled bool
	evalers  []*idEvaler
	env      *eval.Env
	callback Callback
}

type variableItem struct {
	Key    types.String
	Values []types.Value
}

const unknownEntityType = "__cedar::unknown"

func unknownEntity(v types.String) types.EntityUID {
	return types.NewEntityUID(unknownEntityType, v)
}

var errUnboundVariable = fmt.Errorf("unbound variable")
var errUnusedVariable = fmt.Errorf("unused variable")
var errMissingPart = fmt.Errorf("missing part")
var errInvalidPart = fmt.Errorf("invalid part")

// Authorize will run a batch of authorization evaluations.
//
// All the request parts (PARC) must be specified, but you can
// specify Variable or Ignore.  Varibles can be enumerated
// using the Variables.
//
//   - It will error in case of early termination.
//   - It will error in case any of PARC are an incorrect type at eval type.
//   - It will error in case there are unbound variables.
//   - It will error in case there are unused variables.
//
// The result passed to the callback must be used / cloned immediately and not modified.
func Authorize(ctx context.Context, ps *cedar.PolicySet, entityMap types.Entities, request Request, cb Callback, opts ...Option) error {
	be := &batchEvaler{}
	be.ignoreBias = types.Permit
	for _, opt := range opts {
		if opt.ignoreForbid {
			be.ignoreBias = types.Forbid
		}
		if opt.ignorePermit {
			be.ignoreBias = types.Permit
		}
	}
	var found []types.String
	found = findVariables(request.Principal, found)
	found = findVariables(request.Action, found)
	found = findVariables(request.Resource, found)
	found = findVariables(request.Context, found)
	for _, key := range found {
		if _, ok := request.Variables[key]; !ok {
			return fmt.Errorf("%w: %v", errUnboundVariable, key)
		}
	}
	for k := range request.Variables {
		if !slices.Contains(found, k) {
			return fmt.Errorf("%w: %v", errUnusedVariable, k)
		}
	}
	for _, vs := range request.Variables {
		if len(vs) == 0 {
			return nil
		}
	}
	pm := ps.Map()
	be.policies = make([]idPolicy, len(pm))
	i := 0
	for k, p := range pm {
		be.policies[i] = idPolicy{PolicyID: k, Policy: (*ast.Policy)(p.AST())}
		i++
	}
	be.callback = cb
	switch {
	case request.Principal == nil:
		return fmt.Errorf("%w: principal", errMissingPart)
	case request.Action == nil:
		return fmt.Errorf("%w: action", errMissingPart)
	case request.Resource == nil:
		return fmt.Errorf("%w: resource", errMissingPart)
	case request.Context == nil:
		return fmt.Errorf("%w: context", errMissingPart)
	}
	be.env = eval.InitEnv(&eval.Env{
		Entities:  entityMap,
		Principal: request.Principal,
		Action:    request.Action,
		Resource:  request.Resource,
		Context:   request.Context,
	})
	be.Values = Values{}
	for k, v := range request.Variables {
		be.Variables = append(be.Variables, variableItem{Key: k, Values: v})
	}
	slices.SortFunc(be.Variables, func(a, b variableItem) int {
		return len(a.Values) - len(b.Values)
	})

	// resolve ignores if no variables exist
	if len(be.Variables) == 0 {
		doPartial(be)
		fixIgnores(be)
	}

	return doBatch(ctx, be)
}

func doPartial(be *batchEvaler) {
	var np []idPolicy
	for _, p := range be.policies {
		part, keep := eval.PartialPolicy(be.ignoreBias, be.env, p.Policy)
		if !keep {
			continue
		}
		np = append(np, idPolicy{PolicyID: p.PolicyID, Policy: part})
	}
	be.compiled = false
	be.policies = np
	be.evalers = nil
}

func fixIgnores(be *batchEvaler) {
	if eval.IsIgnore(be.env.Principal) {
		be.env.Principal = unknownEntity(consts.Principal)
	}
	if eval.IsIgnore(be.env.Action) {
		be.env.Action = unknownEntity(consts.Action)
	}
	if eval.IsIgnore(be.env.Resource) {
		be.env.Resource = unknownEntity(consts.Resource)
	}
	if eval.IsIgnore(be.env.Context) {
		be.env.Context = types.Record{}
	}
}

func doBatch(ctx context.Context, be *batchEvaler) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// if no variables, authorize
	if len(be.Variables) == 0 {
		return diagnosticAuthzWithCallback(be)
	}

	// save previous state
	prevState := *be

	// else, partial eval what we have so far
	doPartial(be)

	// if no more partial evaluation, fill in ignores with defaults
	if len(be.Variables) == 1 {
		fixIgnores(be)
	}

	// then loop the current variable
	loopEnv := *be.env
	u := be.Variables[0]
	_, chPrincipal := cloneSub(be.env.Principal, u.Key, nil)
	_, chAction := cloneSub(be.env.Action, u.Key, nil)
	_, chResource := cloneSub(be.env.Resource, u.Key, nil)
	_, chContext := cloneSub(be.env.Context, u.Key, nil)
	be.Variables = be.Variables[1:]
	be.Values = maps.Clone(be.Values)
	for _, v := range u.Values {
		*be.env = loopEnv
		be.Values[u.Key] = v
		if chPrincipal {
			be.env.Principal, _ = cloneSub(loopEnv.Principal, u.Key, v)
		}
		if chAction {
			be.env.Action, _ = cloneSub(loopEnv.Action, u.Key, v)
		}
		if chResource {
			be.env.Resource, _ = cloneSub(loopEnv.Resource, u.Key, v)
		}
		if chContext {
			be.env.Context, _ = cloneSub(loopEnv.Context, u.Key, v)
		}
		if err := doBatch(ctx, be); err != nil {
			return err
		}
	}

	// restore previous state
	*be = prevState
	return nil
}

func diagnosticAuthzWithCallback(be *batchEvaler) error {
	var res Result
	var err error
	if res.Request.Principal, err = eval.ValueToEntity(be.env.Principal); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Action, err = eval.ValueToEntity(be.env.Action); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Resource, err = eval.ValueToEntity(be.env.Resource); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	if res.Request.Context, err = eval.ValueToRecord(be.env.Context); err != nil {
		return fmt.Errorf("%w: %w", errInvalidPart, err)
	}
	res.Values = be.Values
	batchCompile(be)
	res.Decision, res.Diagnostic = isAuthorized(be.evalers, be.env)
	be.callback(res)
	return nil
}

func isAuthorized(ps []*idEvaler, env *eval.Env) (types.Decision, types.Diagnostic) {
	var diag types.Diagnostic
	var forbids []types.DiagnosticReason
	var permits []types.DiagnosticReason
	// Don't try to short circuit this.
	// - Even though single forbid means forbid
	// - All policy should be run to collect errors
	// - For permit, all permits must be run to collect annotations
	// - For forbid, forbids must be run to collect annotations
	for _, po := range ps {
		result, err := po.Evaler.Eval(env)
		if err != nil {
			diag.Errors = append(diag.Errors, types.DiagnosticError{PolicyID: po.PolicyID, Position: types.Position(po.Policy.Position), Message: err.Error()})
			continue
		}
		if !result {
			continue
		}
		if po.Policy.Effect == ast.EffectPermit {
			permits = append(permits, types.DiagnosticReason{PolicyID: po.PolicyID, Position: types.Position(po.Policy.Position)})
		} else {
			forbids = append(forbids, types.DiagnosticReason{PolicyID: po.PolicyID, Position: types.Position(po.Policy.Position)})
		}
	}
	if len(forbids) > 0 {
		diag.Reasons = forbids
		return types.Deny, diag
	}
	if len(permits) > 0 {
		diag.Reasons = permits
		return types.Allow, diag
	}
	return types.Deny, diag
}

// func testPrintPolicy(p *ast.Policy) {
// 	pp := (*parser.Policy)(p)
// 	var got bytes.Buffer
// 	pp.MarshalCedar(&got)
// 	fmt.Println(got.String())
// }

func batchCompile(be *batchEvaler) {
	if be.compiled {
		return
	}
	be.evalers = make([]*idEvaler, len(be.policies))
	for i, p := range be.policies {
		be.evalers[i] = &idEvaler{PolicyID: p.PolicyID, Policy: p.Policy, Evaler: eval.Compile(p.Policy)}
	}
	be.compiled = true
}

// cloneSub will return a new value if any of its children have changed
// and signal the change via the boolean
func cloneSub(r types.Value, k types.String, v types.Value) (types.Value, bool) {
	switch t := r.(type) {
	case types.EntityUID:
		if key, ok := eval.ToVariable(t); ok && key == k {
			return v, true
		}
	case types.Record:
		cloned := false
		for kk, vv := range t {
			if vv, delta := cloneSub(vv, k, v); delta {
				if !cloned {
					t = maps.Clone(t) // intentional shallow clone
					cloned = true
				}
				t[kk] = vv
			}
		}
		return t, cloned
	case types.Set:
		cloned := false
		for kk, vv := range t {
			if vv, delta := cloneSub(vv, k, v); delta {
				if !cloned {
					t = slices.Clone(t) // intentional shallow clone
					cloned = true
				}
				t[kk] = vv
			}
		}
		return t, cloned
	}
	return r, false
}

func findVariables(r types.Value, found []types.String) []types.String {
	switch t := r.(type) {
	case types.EntityUID:
		if key, ok := eval.ToVariable(t); ok {
			if !slices.Contains(found, key) {
				found = append(found, key)
			}
		}
	case types.Record:
		for _, vv := range t {
			found = findVariables(vv, found)
		}
	case types.Set:
		for _, vv := range t {
			found = findVariables(vv, found)
		}
	}
	return found
}
