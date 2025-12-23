package schema2_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
)

func TestUnmarshalCedar(t *testing.T) {
	t.Parallel()

	t.Run("simple schema", func(t *testing.T) {
		t.Parallel()
		src := []byte(`entity User;`)
		schema, err := schema2.UnmarshalCedar(src)
		testutil.OK(t, err)
		testutil.Equals(t, len(schema.Nodes), 1)
	})

	t.Run("complex schema", func(t *testing.T) {
		t.Parallel()
		src := []byte(`
namespace MyApp {
	type Name = String;

	entity User in Group {
		name: Name,
		email?: String,
	};

	entity Group;

	action view appliesTo {
		principal: User,
		resource: Document,
	};

	entity Document;
}`)
		schema, err := schema2.UnmarshalCedar(src)
		testutil.OK(t, err)
		ns, ok := schema.Nodes[0].(*ast.NamespaceNode)
		testutil.Equals(t, ok, true)
		testutil.Equals(t, ns.Name, types.Path("MyApp"))
	})

	t.Run("invalid schema", func(t *testing.T) {
		t.Parallel()
		src := []byte(`invalid`)
		_, err := schema2.UnmarshalCedar(src)
		testutil.Equals(t, err != nil, true)
	})
}
