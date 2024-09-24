package types_test

import (
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestString(t *testing.T) {
	t.Parallel()

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		hello := types.String("hello")
		hello2 := types.String("hello")
		goodbye := types.String("goodbye")
		testutil.FatalIf(t, !hello.Equal(hello), "%v not Equal to %v", hello, hello)
		testutil.FatalIf(t, !hello.Equal(hello2), "%v not Equal to %v", hello, hello2)
		testutil.FatalIf(t, hello.Equal(goodbye), "%v Equal to %v", hello, goodbye)
	})

	t.Run("string", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, types.String("hello").String(), `hello`)
		testutil.Equals(t, types.String("hello\ngoodbye").String(), "hello\ngoodbye")
	})

	t.Run("Hash", func(t *testing.T) {
		t.Parallel()

		testutil.Equals(t, types.String("foo").Hash(), types.String("foo").Hash())
		testutil.Equals(t, types.String("bar").Hash(), types.String("bar").Hash())

		// This isn't necessarily true for all values of types.String, but we want to ensure we aren't just returning the
		// same Hash value for types.String.Hash() for every instance.
		testutil.FatalIf(t, types.String("foo").Hash() == types.String("bar").Hash(), "unexpected Hash collision")
	})
}
