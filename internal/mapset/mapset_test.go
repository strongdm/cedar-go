package mapset

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func hashSetMustNotContain[T comparable](t *testing.T, s *MapSet[T], item T) {
	testutil.FatalIf(t, s.Contains(item), "set %v unexpectedly contained item %v", s, 1)
}

func TestHashSet(t *testing.T) {
	t.Run("empty set contains nothing", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.Len(), 0)
		hashSetMustNotContain(t, s, 1)

		s = New[int](10)
		testutil.Equals(t, s.Len(), 0)
		hashSetMustNotContain(t, s, 1)
	})

	t.Run("add => contains", func(t *testing.T) {
		s := New[int]()
		s.Add(1)
		testutil.Equals(t, s.Contains(1), true)
	})

	t.Run("add twice", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.Add(1), true)
		testutil.Equals(t, s.Add(1), false)
	})

	t.Run("add slice", func(t *testing.T) {
		s := New[int]()
		s.AddSlice([]int{1, 2})
		testutil.Equals(t, s.Contains(1), true)
		testutil.Equals(t, s.Contains(2), true)
		hashSetMustNotContain(t, s, 3)
	})

	t.Run("add same slice", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.AddSlice([]int{1, 2}), true)
		testutil.Equals(t, s.AddSlice([]int{1, 2}), false)
	})

	t.Run("add disjoint slices", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.AddSlice([]int{1, 2}), true)
		testutil.Equals(t, s.AddSlice([]int{3, 4}), true)
		testutil.Equals(t, s.AddSlice([]int{1, 2, 3, 4}), false)
	})

	t.Run("add overlapping slices", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.AddSlice([]int{1, 2}), true)
		testutil.Equals(t, s.AddSlice([]int{2, 3}), true)
		testutil.Equals(t, s.AddSlice([]int{1, 3}), false)
	})

	t.Run("remove nonexistent", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.Remove(1), false)
	})

	t.Run("remove existing", func(t *testing.T) {
		s := New[int]()
		s.Add(1)
		testutil.Equals(t, s.Remove(1), true)
	})

	t.Run("remove => !contains", func(t *testing.T) {
		s := New[int]()
		s.Add(1)
		s.Remove(1)
		testutil.FatalIf(t, s.Contains(1), "set unexpectedly contained item")
	})

	t.Run("remove slice", func(t *testing.T) {
		s := New[int]()
		s.AddSlice([]int{1, 2, 3})
		s.RemoveSlice([]int{1, 2})
		hashSetMustNotContain(t, s, 1)
		hashSetMustNotContain(t, s, 2)
		testutil.Equals(t, s.Contains(3), true)
	})

	t.Run("remove non-existent slice", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.RemoveSlice([]int{1, 2}), false)
	})

	t.Run("remove overlapping slice", func(t *testing.T) {
		s := New[int]()
		s.Add(1)
		testutil.Equals(t, s.RemoveSlice([]int{1, 2}), true)
		testutil.Equals(t, s.RemoveSlice([]int{1, 2}), false)
	})

	t.Run("new from slice", func(t *testing.T) {
		s := FromSlice([]int{1, 2, 2, 3})
		testutil.Equals(t, s.Len(), 3)
		testutil.Equals(t, s.Contains(1), true)
		testutil.Equals(t, s.Contains(2), true)
		testutil.Equals(t, s.Contains(3), true)
	})

	t.Run("slice", func(t *testing.T) {
		s := New[int]()
		testutil.Equals(t, s.Slice(), []int{})

		s = New[int](10)
		testutil.Equals(t, s.Slice(), []int{})

		inSlice := []int{1, 2, 3}
		s = FromSlice(inSlice)
		outSlice := s.Slice()
		slices.Sort(outSlice)
		testutil.Equals(t, inSlice, outSlice)
	})

	t.Run("equal", func(t *testing.T) {
		s1 := FromSlice([]int{1, 2, 3})
		testutil.Equals(t, s1.Equal(s1), true)

		s2 := FromSlice([]int{1, 2, 3})
		testutil.Equals(t, s1.Equal(s2), true)

		s2.Add(4)
		testutil.Equals(t, s1.Equal(s2), false)

		s2.Remove(3)
		testutil.Equals(t, s1.Equal(s2), false)

		s1.Add(4)
		s1.Remove(3)
		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("iterate", func(t *testing.T) {
		s1 := FromSlice([]int{1, 2, 3})

		s2 := New[int]()
		s1.Iterate(func(item int) bool {
			s2.Add(item)
			return true
		})

		testutil.Equals(t, s1.Equal(s2), true)
	})

	t.Run("iterate break early", func(t *testing.T) {
		s1 := FromSlice([]int{1, 2, 3})

		i := 0
		var items []int
		s1.Iterate(func(item int) bool {
			if i == 2 {
				return false
			}
			items = append(items, item)
			i++
			return true
		})

		// Because iteration order is non-deterministic, all we can say is that the right number of items ended up in
		// the set and that the items were in the original set.
		testutil.Equals(t, len(items), 2)
		testutil.Equals(t, s1.Contains(items[0]), true)
		testutil.Equals(t, s1.Contains(items[1]), true)
	})

	t.Run("intersection with overlap", func(t *testing.T) {
		s1 := FromSlice([]int{1, 2, 3})
		s2 := FromSlice([]int{2, 3, 4})

		s3 := s1.Intersection(s2)
		testutil.Equals(t, s3, FromSlice([]int{2, 3}))

		s4 := s1.Intersection(s2)
		testutil.Equals(t, s4, FromSlice([]int{2, 3}))
	})

	t.Run("intersection disjoint", func(t *testing.T) {
		s1 := FromSlice([]int{1, 2})
		s2 := FromSlice([]int{3, 4})

		s3 := s1.Intersection(s2)
		testutil.Equals(t, s3.Len(), 0)

		s4 := s1.Intersection(s2)
		testutil.Equals(t, s4.Len(), 0)
	})

	t.Run("encode nil set", func(t *testing.T) {
		s := New[int]()

		out, err := json.Marshal(s)

		testutil.OK(t, err)
		testutil.Equals(t, string(out), "[]")
	})

	t.Run("encode json", func(t *testing.T) {
		s := FromSlice([]int{1, 2, 3})

		out, err := json.Marshal(s)

		correctOutputs := []string{
			"[1,2,3]",
			"[1,3,2]",
			"[2,1,3]",
			"[2,3,1]",
			"[3,1,2]",
			"[3,2,1]",
		}

		testutil.OK(t, err)
		testutil.FatalIf(t, !slices.Contains(correctOutputs, string(out)), "%v is not a valid output", string(out))
	})

	t.Run("decode json", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte("[2,3,1,2]"))
		testutil.OK(t, err)
		testutil.Equals(t, &s1, FromSlice([]int{1, 2, 3}))
	})

	t.Run("decode json empty", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte("[]"))
		testutil.OK(t, err)
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode mixed types in array", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte(`[2,3,1,"2"]`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type int")
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("decode wrong type", func(t *testing.T) {
		var s1 MapSet[int]
		err := s1.UnmarshalJSON([]byte(`"1,2,3"`))
		testutil.Error(t, err)
		testutil.Equals(t, err.Error(), "json: cannot unmarshal string into Go value of type []int")
		testutil.Equals(t, s1.Len(), 0)
	})

	t.Run("panic if too many args", func(t *testing.T) {
		t.Parallel()

		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("code did not panic as expected")
			}
		}()

		New[int](0, 1)
	})

	// The zero value MapSet is usable, but care must be taken to ensure that it is not mutated when passed by value
	// because those mutations may or may not be reflected in the caller's version of the MapSet.
	t.Run("zero value", func(t *testing.T) {
		s := MapSet[int]{}
		hashSetMustNotContain(t, &s, 1)
		testutil.Equals(t, s.Slice(), nil)

		addByValue := func(m MapSet[int], val int) {
			m.Add(val)
		}

		// Calling addByValue when s is still the zero value results in no mutation
		addByValue(s, 1)
		testutil.Equals(t, s.Len(), 0)

		// However, calling addByValue after the internal map in s has been initialized results in mutation
		s.Add(0)
		testutil.Equals(t, s.Len(), 1)
		addByValue(s, 1)
		testutil.Equals(t, s.Len(), 2)
	})

}