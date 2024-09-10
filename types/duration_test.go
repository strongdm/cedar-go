package types_test

import (
	"fmt"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
)

func TestDuration(t *testing.T) {
	t.Parallel()
	{
		tests := []struct{ in, out string }{
			{"1h", "1h"},
			{"60m", "1h"},
			{"3600s", "1h"},
			{"3600000ms", "1h"},
			{"24h", "1d"},
			{"36h", "1d12h"},
			{"1d12h", "1d12h"},
			{"1d11h60m", "1d12h"},
			{"1d11h59m60s", "1d12h"},
			{"1d11h59m59s1000ms", "1d12h"},
			{"60s60000ms", "2m"},
			{"62m", "1h2m"},
			{"2m3600s", "1h2m"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.out), func(t *testing.T) {
				t.Parallel()
				d, err := types.ParseDuration(tt.in)
				testutil.OK(t, err)
				testutil.Equals(t, d.String(), tt.out)
			})
		}
	}

	{
		tests := []struct{ in, errStr string }{
			{"", "error parsing duration value: string too short"},
			{"-", "error parsing duration value: string too short"},
			{"h", "error parsing duration value: string too short"},
			{"3", "error parsing duration value: string too short"},
			{"-m", "error parsing duration value: unit found without quantity"},
			{"-1t", "error parsing duration value: unexpected character 't'"},
			{"-1h1h", "error parsing duration value: unexpected unit 'h'"},
			{"-3h3", "error parsing duration value: expected unit"},
			{"3h-1m", "error parsing duration value: unexpected character '-'"},
			{"3h1m   ", "error parsing duration value: unexpected character ' '"},
			{"3600ms30ms", "error parsing duration value: invalid duration"},
			{"36ms30h", "error parsing duration value: invalid duration"},
			{"999999999999999999999ms", "error parsing duration value: overflow"},
		}
		for ti, tt := range tests {
			tt := tt
			t.Run(fmt.Sprintf("%d_%s->%s", ti, tt.in, tt.errStr), func(t *testing.T) {
				t.Parallel()
				_, err := types.ParseDuration(tt.in)
				testutil.ErrorIs(t, err, types.ErrDuration)
				testutil.Equals(t, err.Error(), tt.errStr)
			})
		}
	}

	t.Run("Equal", func(t *testing.T) {
		t.Parallel()
		one := types.UnsafeDuration(1)
		one2 := types.UnsafeDuration(1)
		zero := types.UnsafeDuration(0)
		f := types.Boolean(false)
		testutil.FatalIf(t, !one.Equal(one), "%v not Equal to %v", one, one)
		testutil.FatalIf(t, !one.Equal(one2), "%v not Equal to %v", one, one2)
		testutil.FatalIf(t, one.Equal(zero), "%v Equal to %v", one, zero)
		testutil.FatalIf(t, zero.Equal(one), "%v Equal to %v", zero, one)
		testutil.FatalIf(t, zero.Equal(f), "%v Equal to %v", zero, f)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		testutil.Equals(t, string(types.UnsafeDuration(42).MarshalCedar()), `duration("42ms")`)
	})

}