package sets_test

import (
	"fmt"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/sets"
	"github.com/cedar-policy/cedar-go/internal/testutil"
)

type dummyItem struct {
	Value   string `json:"value"`
	HashVal uint64 `json:"hashVal"`
}

func (d dummyItem) Hash() uint64 {
	return d.HashVal
}

func (d dummyItem) Equal(o dummyItem) bool {
	return d.Value == o.Value
}

type jsonErrItem struct{}

func (jsonErrItem) Hash() uint64 { return 0 }

func (jsonErrItem) Equal(jsonErrItem) bool { return true }

func (jsonErrItem) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("marshal error")
}

func (jsonErrItem) UnmarshalJSON([]byte) error {
	return fmt.Errorf("unmarshal error")
}

func TestSet(t *testing.T) {
	t.Parallel()

	item1 := dummyItem{"one", 1}
	item2 := dummyItem{"two", 2}
	item3 := dummyItem{"three", 3}
	item4 := dummyItem{"four", 1} // Hash collision with item1

	t.Run("empty set", func(t *testing.T) {
		empty := sets.ImmutableHashSet[dummyItem]{}
		testutil.Equals(t, empty.Len(), 0)

		testutil.Equals(t, empty.Slice(), nil)

		empty.Iterate(func(dummyItem) bool {
			testutil.FatalIf(t, true, "unexpected iteration")
			return false
		})
		testutil.Equals(t, empty.Contains(dummyItem{"foo", 1}), false)

		out, err := empty.MarshalJSON()
		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[]")

		var emptyUnmarshaled sets.ImmutableHashSet[dummyItem]
		err = empty.UnmarshalJSON([]byte("[]"))
		testutil.OK(t, err)
		testutil.Equals(t, empty.Equal(emptyUnmarshaled), true)

		empty2 := sets.NewImmutableHashSet([]dummyItem{})
		testutil.Equals(t, empty.Equal(empty2), true)
		testutil.Equals(t, empty2.Equal(empty), true)
	})

	t.Run("len", func(t *testing.T) {
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1}).Len(), 1)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item2}).Len(), 2)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item2, item3}).Len(), 3)
	})

	t.Run("duplicates removed", func(t *testing.T) {
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item1, item1}).Len(), 1)
	})

	t.Run("collisions on initialization", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]dummyItem{item1, item4})
		testutil.Equals(t, s.Contains(item1), true)
		testutil.Equals(t, s.Contains(item4), true)
	})

	t.Run("iterate full", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]dummyItem{item1, item2, item3})

		var items []dummyItem
		s.Iterate(func(i dummyItem) bool {
			items = append(items, i)
			return true
		})

		testutil.Equals(t, len(items), 3)
		testutil.Equals(t, slices.Contains(items, item1), true)
		testutil.Equals(t, slices.Contains(items, item2), true)
		testutil.Equals(t, slices.Contains(items, item3), true)
	})

	t.Run("iterate partial", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]dummyItem{item1, item2, item3})

		var items []dummyItem
		s.Iterate(func(i dummyItem) bool {
			items = append(items, i)
			return false
		})

		testutil.Equals(t, len(items), 1)
		testutil.Equals(
			t, slices.Contains(items, item1) || slices.Contains(items, item2) || slices.Contains(items, item3), true)
	})

	t.Run("contains", func(t *testing.T) {
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1}).Contains(item1), true)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1}).Contains(item2), false)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item2}).Contains(item1), true)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item2}).Contains(item2), true)
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1, item2}).Contains(item3), false)
	})

	t.Run("contains collision", func(t *testing.T) {
		testutil.Equals(t, sets.NewImmutableHashSet([]dummyItem{item1}).Contains(item4), false)
	})

	t.Run("slice", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]dummyItem{item1, item2, item3})
		items := s.Slice()
		testutil.Equals(t, len(items), 3)
		testutil.Equals(t, slices.Contains(items, item1), true)
		testutil.Equals(t, slices.Contains(items, item2), true)
		testutil.Equals(t, slices.Contains(items, item3), true)
	})

	t.Run("equal", func(t *testing.T) {
		s1 := sets.NewImmutableHashSet([]dummyItem{item1})
		s123 := sets.NewImmutableHashSet([]dummyItem{item1, item2, item3})
		s321 := sets.NewImmutableHashSet([]dummyItem{item3, item2, item1})
		testutil.Equals(t, s1.Equal(s123), false)
		testutil.Equals(t, s123.Equal(s1), false)

		s12 := sets.NewImmutableHashSet([]dummyItem{item1, item2})
		s13 := sets.NewImmutableHashSet([]dummyItem{item1, item3})
		testutil.Equals(t, s123.Equal(s321), true)
		testutil.Equals(t, s12.Equal(s13), false)

		s4 := sets.NewImmutableHashSet([]dummyItem{item4}) // collision with item1
		testutil.Equals(t, s1.Equal(s4), false)
	})

	t.Run("marshal json", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]dummyItem{item1, item2, item3})

		out, err := s.MarshalJSON()

		testutil.OK(t, err)
		testutil.Equals(t, string(out), `[{"value":"one","hashVal":1},{"value":"two","hashVal":2},{"value":"three","hashVal":3}]`)
	})

	t.Run("marshal json error", func(t *testing.T) {
		s := sets.NewImmutableHashSet([]jsonErrItem{{}})

		out, err := s.MarshalJSON()

		testutil.Error(t, err)
		testutil.Equals(t, out, nil)
	})

	t.Run("unmarshal json", func(t *testing.T) {

		json := `[{"value":"one","hashVal":1},{"value":"two","hashVal":2},{"value":"three","hashVal":3}]`
		var s sets.ImmutableHashSet[dummyItem]
		err := s.UnmarshalJSON([]byte(json))

		testutil.OK(t, err)
		testutil.Equals(t, s, sets.NewImmutableHashSet([]dummyItem{item1, item2, item3}))
	})

	t.Run("unmarshal json duplicates", func(t *testing.T) {

		json := `[{"value":"one","hashVal":1},{"value":"one","hashVal":1}]`
		var s sets.ImmutableHashSet[dummyItem]
		err := s.UnmarshalJSON([]byte(json))

		testutil.OK(t, err)
		testutil.Equals(t, s, sets.NewImmutableHashSet([]dummyItem{item1}))
	})

	t.Run("unmarshal json error", func(t *testing.T) {
		var s sets.ImmutableHashSet[jsonErrItem]
		err := s.UnmarshalJSON([]byte("[{}]"))

		testutil.Error(t, err)
	})
}
