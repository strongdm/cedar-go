package types

import (
	"testing"
)

func TestTypes(t *testing.T) {
	t.Parallel()

	user := EntityType{ID: "User"}
	johnny := EntityUID{
		Type: user,
		ID:   "johnny",
	}
	action := EntityType{ID: "Action"}
	seed := EntityType{ID: "Seed"}
	genus := EntityType{ID: "Genus"}
	classification := EntityType{ID: "Classification"}

	_ = PolicySet{
		Policies: []Policy{
			{
				Effect:    PermitEffect,
				Principal: EqScopeConstraint{johnny},
				Action: InEntitiesScopeConstraint([]EntityUID{
					{
						Type: action,
						ID:   "sow",
					},
					{
						Type: action,
						ID:   "cast",
					},
				}),
				Resource: IsInScopeConstraint{seed, EntityUID{genus, "Malus"}},
				Conditions: []Condition{
					WhenCondition{Boolean(true)},
					UnlessCondition{Boolean(false)},
				},
			},
			{
				Effect:    ForbidEffect,
				Principal: EqScopeConstraint{johnny},
				Resource:  InEntityScopeConstraint{classification, "Poisonous"},
			},
			{
				Effect: ForbidEffect,
				Conditions: []Condition{
					WhenCondition{
						Contains{
							LHS: GetAttribute{
								LHS:  ResourceVariable,
								Name: "tags",
							},
							Arg: String("private"),
						},
					},
					UnlessCondition{
						In{
							LHS: ResourceVariable,
							RHS: GetAttribute{
								LHS:  PrincipalVariable,
								Name: "account",
							},
						},
					},
				},
			},
		},
	}
}
