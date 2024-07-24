package sugar

import (
	"github.com/cedar-policy/cedar-go/types"
)

type Annotate map[string]string

func (a Annotate) Policy(s *scope, conditions ...types.Condition) types.Policy {
	return types.Policy{
		// Annotations: // TODO
		Effect:     s.effect,
		Principal:  s.principal,
		Action:     s.action,
		Resource:   s.resource,
		Conditions: conditions,
	}
}

type scope struct {
	effect    types.Effect
	principal types.PrincipalConstraint
	action    types.ActionConstraint
	resource  types.ResourceConstraint
}

func entityFromPath(path []string) types.EntityUID {
	return types.EntityUID{
		Type: typeFromPath(path[:len(path)-1]),
		ID:   path[len(path)-1],
	}
}

func typeFromPath(path []string) types.EntityType {
	if len(path) == 1 {
		return types.EntityType{ID: path[0]}
	}
	return types.EntityType{
		ID:         path[0],
		Namespaces: path[1 : len(path)-1],
	}
}
func EntityUID(path ...string) types.EntityUID {
	return entityFromPath(path)
}

func (s *scope) Principal() *scope {
	return s
}
func (s *scope) Action() *scope {
	return s
}
func (s *scope) Resource() *scope {
	return s
}

func (s *scope) PrincipalEq(path ...string) *scope {
	s.principal = types.EqScopeConstraint{
		Entity: entityFromPath(path),
	}
	return s
}

func (s *scope) ActionIn(ents ...types.EntityUID) *scope {
	var c types.InEntitiesScopeConstraint
	for _, e := range ents {
		c = append(c, e)
	}
	s.action = c
	return s
}

func (s *scope) ResourceIn(path ...string) *scope {
	s.resource = types.InEntityScopeConstraint(entityFromPath(path))
	return s
}
func (s *scope) ResourceIsIn(typ []string, ent ...string) *scope {
	s.resource = types.IsInScopeConstraint{
		EntityType: typeFromPath(typ),
		EntityUID:  entityFromPath(ent),
	}
	return s
}

func Permit() *scope {
	return &scope{
		effect: types.PermitEffect,
	}
}

func Forbid() *scope {
	return &scope{
		effect: types.ForbidEffect,
	}
}

func When(x expr) types.Condition {
	return types.WhenCondition{
		Expr: x.Expr,
	}
}

func Unless(x expr) types.Condition {
	return types.UnlessCondition{
		Expr: x.Expr,
	}
}

type expr struct {
	Expr types.Expr
}

func (e expr) Index(k string) expr {
	return expr{
		Expr: types.GetAttribute{
			LHS:  e.Expr,
			Name: k,
		},
	}
}

func (e expr) Contains(v expr) expr {
	return expr{
		Expr: types.Contains{
			LHS: e.Expr,
			Arg: v.Expr,
		},
	}
}

func (e expr) In(v expr) expr {
	return expr{
		Expr: types.In{
			LHS: e.Expr,
			RHS: v.Expr,
		},
	}
}

func String(v string) expr {
	return expr{
		Expr: types.String(v),
	}
}

func True() expr {
	return expr{
		Expr: types.Boolean(true),
	}
}

func False() expr {
	return expr{
		Expr: types.Boolean(false),
	}
}
func Principal() expr {
	return expr{
		Expr: types.PrincipalVariable,
	}
}
func Action() expr {
	return expr{
		Expr: types.ActionVariable,
	}
}
func Resource() expr {
	return expr{
		Expr: types.ResourceVariable,
	}
}
func Context() expr {
	return expr{
		Expr: types.ContextVariable,
	}
}
