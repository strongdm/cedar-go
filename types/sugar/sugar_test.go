package sugar_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/types/sugar"
)

/*

permit (
	principal = User::"johnny",
	action in [Action::"sow", Action::"cast"],
	resource is Seed in Genus::"Malus"
) when {
	true
} unless {
 	false
};

forbid {
	principal = User::"johnny",
	action,
	resource in Classification::"Poisonous"
};

forbid {
	principal,
	action,
	resource
} when {
	resource.tags.contains("private")
} unless {
	resource in principal.account
};

*/

func TestSugar(t *testing.T) {
	t.Parallel()
	_ = sugar.Annotate{
		"example": "one",
	}.Policy(
		sugar.Permit().
			PrincipalEq("User", "johnny").
			ActionIn(sugar.EntityUID("Action", "sow"), sugar.EntityUID("Action", "cast")).
			ResourceIsIn([]string{"Seed"}, "Genus", "Malus"),
		sugar.When(sugar.True()),
		sugar.Unless(sugar.False()),
	)

	_ = sugar.Annotate{
		"example": "two",
	}.Policy(
		sugar.Forbid().
			PrincipalEq("User", "johnny").
			Action().
			ResourceIn("Classification", "Poisonous"),
	)

	_ = sugar.Annotate{
		"example": "three",
	}.Policy(
		sugar.Forbid().
			Principal().
			Action().
			Resource(),
		sugar.When(
			sugar.Resource().Index("tags").Contains(sugar.String("private")),
		),
		sugar.Unless(
			sugar.Resource().In(sugar.Principal().Index("account")),
		),
	)

}
