package types_test

import (
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

type colliderValue struct {
	Value   types.Value
	HashVal uint64
}

func (c colliderValue) String() string           { return "" }
func (c colliderValue) MarshalCedar() []byte     { return nil }
func (c colliderValue) Equal(v types.Value) bool { return v.Equal(c.Value) }
func (c colliderValue) Hash() uint64             { return c.HashVal }

func TestSet(t *testing.T) {
	t.Parallel()

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.Set{}.String(), "[]")
		testutil.Equals(
			t,
			types.NewSet([]types.Value{types.Boolean(true), types.Long(1)}).String(),
			"[true, 1]")
	})

	// This test is intended to show the NewSet makes a copy of the Values in the input slice
	t.Run("immutable", func(t *testing.T) {
		t.Parallel()

		slice := []types.Value{types.Long(42)}
		p := &slice[0]

		set := types.NewSet(slice)

		*p = types.Long(1337)

		testutil.Equals(t, set.Len(), 1)

		var got types.Long
		set.Iterate(func(v types.Value) bool {
			var ok bool
			got, ok = v.(types.Long)
			testutil.FatalIf(t, !ok, "incorrect type for set element")
			return true
		})

		testutil.Equals(t, got, types.Long(42))
	})

	t.Run("Hash", func(t *testing.T) {
		t.Parallel()

		t.Run("order independent", func(t *testing.T) {
			t.Parallel()
			s1 := types.NewSet([]types.Value{types.Long(42), types.Long(1337)})
			s2 := types.NewSet([]types.Value{types.Long(1337), types.Long(42)})
			testutil.Equals(t, s1.Hash(), s2.Hash())
		})

		t.Run("order independent with collisions", func(t *testing.T) {
			t.Parallel()

			v1 := colliderValue{Value: types.String("foo"), HashVal: 1337}
			v2 := colliderValue{Value: types.String("bar"), HashVal: 1337}
			v3 := colliderValue{Value: types.String("baz"), HashVal: 1337}

			permutations := []types.Set{
				types.NewSet([]types.Value{v1, v2, v3}),
				types.NewSet([]types.Value{v1, v3, v2}),
				types.NewSet([]types.Value{v2, v1, v3}),
				types.NewSet([]types.Value{v2, v3, v1}),
				types.NewSet([]types.Value{v3, v1, v2}),
				types.NewSet([]types.Value{v3, v2, v1}),
			}
			expected := permutations[0].Hash()
			for _, p := range permutations {
				testutil.Equals(t, p.Hash(), expected)
			}
		})

		t.Run("order independent with interleaving collisions", func(t *testing.T) {
			t.Parallel()

			v1 := colliderValue{Value: types.String("foo"), HashVal: 1337}
			v2 := colliderValue{Value: types.String("bar"), HashVal: 1338}
			v3 := colliderValue{Value: types.String("baz"), HashVal: 1337}

			permutations := []types.Set{
				types.NewSet([]types.Value{v1, v2, v3}),
				types.NewSet([]types.Value{v1, v3, v2}),
				types.NewSet([]types.Value{v2, v1, v3}),
				types.NewSet([]types.Value{v2, v3, v1}),
				types.NewSet([]types.Value{v3, v1, v2}),
				types.NewSet([]types.Value{v3, v2, v1}),
			}
			expected := permutations[0].Hash()
			for _, p := range permutations {
				testutil.Equals(t, p.Hash(), expected)
			}
		})

		t.Run("duplicates unimportant", func(t *testing.T) {
			t.Parallel()
			s1 := types.NewSet([]types.Value{types.Long(42), types.Long(1337)})
			s2 := types.NewSet([]types.Value{types.Long(42), types.Long(1337), types.Long(1337)})
			testutil.Equals(t, s1.Hash(), s2.Hash())
		})

		t.Run("empty set", func(t *testing.T) {
			t.Parallel()
			m1 := types.Set{}
			m2 := types.NewSet([]types.Value{})
			m3 := types.NewSet(nil)
			testutil.Equals(t, m1.Hash(), m2.Hash())
			testutil.Equals(t, m2.Hash(), m3.Hash())
		})

		// These tests don't necessarily hold for all values of Set, but we want to ensure we are considering
		// different aspects of the Set, which these particular tests demonstrate.

		t.Run("extra element", func(t *testing.T) {
			t.Parallel()
			s1 := types.NewSet([]types.Value{types.Long(42), types.Long(1337)})
			s2 := types.NewSet([]types.Value{types.Long(42), types.Long(1337), types.Long(1)})
			testutil.FatalIf(t, s1.Hash() == s2.Hash(), "unexpected Hash collision")
		})

		t.Run("disjoint", func(t *testing.T) {
			t.Parallel()
			s1 := types.NewSet([]types.Value{types.Long(42), types.Long(1337)})
			s2 := types.NewSet([]types.Value{types.Long(0), types.String("hi")})
			testutil.FatalIf(t, s1.Hash() == s2.Hash(), "unexpected Hash collision")
		})
	})

	t.Run("collisions", func(t *testing.T) {
		t.Parallel()

		v1 := colliderValue{Value: types.String("foo"), HashVal: 1337}
		v2 := colliderValue{Value: types.String("bar"), HashVal: 1337}
		v3 := colliderValue{Value: types.String("baz"), HashVal: 1338}
		v4 := colliderValue{Value: types.String("baz"), HashVal: 1337}

		set := types.NewSet([]types.Value{v1, v2, v3, v4})

		testutil.Equals(t, set.Len(), 3)

		var vals []types.Value
		set.Iterate(func(v types.Value) bool {
			vals = append(vals, v)
			return true
		})

		testutil.Equals(t, slices.ContainsFunc(vals, func(v types.Value) bool { return v.Equal(v1) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v types.Value) bool { return v.Equal(v2) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v types.Value) bool { return v.Equal(v3) }), true)
		testutil.Equals(t, slices.ContainsFunc(vals, func(v types.Value) bool { return v.Equal(v4) }), true)
	})
}
