package types

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
)

func zeroValue() Value {
	return nil
}

func mustDecimalValue(v string) Decimal {
	r, _ := ParseDecimal(v)
	return r
}

func mustDatetimeValue(v string) Datetime {
	r, _ := ParseDatetime(v)
	return r
}

func mustDurationValue(v string) Duration {
	r, _ := ParseDuration(v)
	return r
}

func mustIPValue(v string) IPAddr {
	r, _ := ParseIPAddr(v)
	return r
}

func AssertValue(t *testing.T, got, want Value) {
	t.Helper()
	testutil.FatalIf(
		t,
		!((got == zeroValue() && want == zeroValue()) ||
			(got != zeroValue() && want != zeroValue() && got.Equal(want))),
		"got %v want %v", got, want)
}

func TestJSON_Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want Value
		err  error
	}{
		{"impliedEntity", `{ "type": "User", "id": "alice" }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"explicitEntity", `{ "__entity": { "type": "User", "id": "alice" } }`, EntityUID{Type: "User", ID: "alice"}, nil},
		{"impliedLongEntity", `{ "type": "User::External", "id": "alice" }`, EntityUID{Type: "User::External", ID: "alice"}, nil},
		{"explicitLongEntity", `{ "__entity": { "type": "User::External", "id": "alice" } }`, EntityUID{Type: "User::External", ID: "alice"}, nil},
		{"invalidJSON", `!@#$`, zeroValue(), errJSONDecode},
		{"numericOverflow", "12341234123412341234", zeroValue(), errJSONLongOutOfRange},
		{"unsupportedNull", "null", zeroValue(), errJSONUnsupportedType},
		{"explicitIP", `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`, mustIPValue("222.222.222.7"), nil},
		{"explicitSubnet", `{ "__extn": { "fn": "ip", "arg": "192.168.0.0/16" } }`, mustIPValue("192.168.0.0/16"), nil},
		{"explicitDecimal", `{ "__extn": { "fn": "decimal", "arg": "33.57" } }`, mustDecimalValue("33.57"), nil},
		{"explicitDatetime", `{ "__extn": { "fn": "datetime", "arg": "1970-01-01T00:00:01Z" } }`, mustDatetimeValue("1970-01-01T00:00:01Z"), nil},
		{"explicitDuration", `{ "__extn": { "fn": "duration", "arg": "1d12h30m30s500ms" } }`, mustDurationValue("1d12h30m30s500ms"), nil},
		{"invalidExtension", `{ "__extn": { "fn": "asdf", "arg": "blah" } }`, zeroValue(), errJSONInvalidExtn},
		{"badIP", `{ "__extn": { "fn": "ip", "arg": "bad" } }`, zeroValue(), ErrIP},
		{"badDecimal", `{ "__extn": { "fn": "decimal", "arg": "bad" } }`, zeroValue(), ErrDecimal},
		{"badDatetime", `{ "__extn": { "fn": "datetime", "arg": "bad" } }`, zeroValue(), ErrDatetime},
		{"badDuration", `{ "__extn": { "fn": "duration", "arg": "bad" } }`, zeroValue(), ErrDuration},
		{"set", `[42]`, NewSet([]Value{Long(42)}), nil},
		{"record", `{"a":"b"}`, NewRecord(RecordMap{"a": String("b")}), nil},
		{"bool", `false`, Boolean(false), nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got Value
			ptr := &got
			err := UnmarshalJSON([]byte(tt.in), ptr)
			testutil.ErrorIs(t, err, tt.err)
			AssertValue(t, got, tt.want)
			if tt.err != nil {
				return
			}

			// Now assert that when we Marshal/Unmarshal that value, we still
			// have what we started with
			gotJSON, err := json.Marshal(ptr)
			testutil.OK(t, err)
			var gotRetry Value
			ptr = &gotRetry
			err = UnmarshalJSON(gotJSON, ptr)
			testutil.OK(t, err)
			testutil.Equals(t, gotRetry, tt.want)
		})
	}
}

func TestTypedJSONUnmarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		f         func(b []byte) (Value, error)
		in        string
		wantValue Value
		wantErr   error
	}{
		{
			name: "string",
			f: func(b []byte) (Value, error) {
				var res String
				err := json.Unmarshal(b, &res)
				return res, err
			},
			in:        `"hello"`,
			wantValue: String("hello"),
			wantErr:   nil,
		},
		{
			name: "ip",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "ip", "arg": "222.222.222.7" } }`,
			wantValue: mustIPValue("222.222.222.7"),
			wantErr:   nil,
		},
		{
			name: "ip/implicit",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"222.222.222.7"`,
			wantValue: mustIPValue("222.222.222.7"),
			wantErr:   nil,
		},
		{
			name: "ip/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: IPAddr{},
			wantErr:   errJSONDecode,
		},
		{
			name: "ip/badArg",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "ip", "arg": "bad" } }`,
			wantValue: IPAddr{},
			wantErr:   ErrIP,
		},
		{
			name: "ip/badJSON",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: IPAddr{},
			wantErr:   errJSONDecode,
		},
		{
			name: "ip/badFn",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "222.222.222.7" } }`,
			wantValue: IPAddr{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name: "ip/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res IPAddr
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: IPAddr{},
			wantErr:   errJSONExtNotFound,
		},

		{
			name: "decimal",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "decimal", "arg": "1234.5678" } }`,
			wantValue: mustDecimalValue("1234.5678"),
			wantErr:   nil,
		},
		{
			name: "decimal/implicit",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"1234.5678"`,
			wantValue: mustDecimalValue("1234.5678"),
			wantErr:   nil,
		},
		{
			name: "decimal/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: Decimal{},
			wantErr:   errJSONDecode,
		},
		{
			name: "decimal/badArg",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "decimal", "arg": "bad" } }`,
			wantValue: Decimal{},
			wantErr:   ErrDecimal,
		},
		{
			name: "decimal/badJSON",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: Decimal{},
			wantErr:   errJSONDecode,
		},
		{
			name: "decimal/badFn",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "1234.5678" } }`,
			wantValue: Decimal{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name: "decimal/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res Decimal
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: Decimal{},
			wantErr:   errJSONExtNotFound,
		},

		{
			name: "datetime",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "datetime", "arg": "1970-01-01T00:00:01Z" } }`,
			wantValue: mustDatetimeValue("1970-01-01T00:00:01Z"),
			wantErr:   nil,
		},
		{
			name: "datetime/implicit",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"1970-01-01T00:00:01Z"`,
			wantValue: mustDatetimeValue("1970-01-01T00:00:01Z"),
			wantErr:   nil,
		},
		{
			name: "datetime/direct",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"datetime(\"1970-01-01T00:00:01Z\")"`,
			wantValue: mustDatetimeValue("1970-01-01T00:00:01Z"),
			wantErr:   nil,
		},
		{
			name: "datetime/direct/JSON",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "datetime", "arg": "1970-01-01T00:00:01Z" }`,
			wantValue: mustDatetimeValue("1970-01-01T00:00:01Z"),
			wantErr:   nil,
		},
		{
			name: "datetime/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: Datetime{},
			wantErr:   errJSONDecode,
		},
		{
			name: "datetime/badArg",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "datetime", "arg": "bad" } }`,
			wantValue: Datetime{},
			wantErr:   ErrDatetime,
		},
		{
			name: "datetime/badJSON",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: Datetime{},
			wantErr:   errJSONDecode,
		},
		{
			name: "datetime/badFn",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "1970-01-01T00:00:01Z" } }`,
			wantValue: Datetime{},
			wantErr:   errJSONExtFnMatch,
		},
		{
			name: "datetime/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: Datetime{},
			wantErr:   errJSONExtNotFound,
		},

		{
			name: "datetime/direct/badJSON",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "datetime", "arg": "1970-01-01`,
			wantValue: Datetime{},
			wantErr:   errJSONDecode,
		},
		{
			name: "datetime/direct/badFn",
			f: func(b []byte) (Value, error) {
				var res Datetime
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "bad", "arg": "1970-01-01" }`,
			wantValue: Datetime{},
			wantErr:   errJSONExtFnMatch,
		},

		{
			name: "duration",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "duration", "arg": "1ms" } }`,
			wantValue: mustDurationValue("1ms"),
			wantErr:   nil,
		},
		{
			name: "duration/implicit",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"1ms"`,
			wantValue: mustDurationValue("1ms"),
			wantErr:   nil,
		},
		{
			name: "duration/direct",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"duration(\"1ms\")"`,
			wantValue: mustDurationValue("1ms"),
			wantErr:   nil,
		},
		{
			name: "duration/direct/JSON",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "duration", "arg": "1ms" }`,
			wantValue: mustDurationValue("1ms"),
			wantErr:   nil,
		},

		{
			name: "duration/implicit/badJSON",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `"bad`,
			wantValue: Duration{},
			wantErr:   errJSONDecode,
		},
		{
			name: "duration/badArg",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "duration", "arg": "bad" } }`,
			wantValue: Duration{},
			wantErr:   ErrDuration,
		},
		{
			name: "duration/badJSON",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `bad`,
			wantValue: Duration{},
			wantErr:   errJSONDecode,
		},
		{
			name: "duration/badFn",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "__extn": { "fn": "bad", "arg": "10ms" } }`,
			wantValue: Duration{},
			wantErr:   errJSONExtFnMatch,
		},

		{
			name: "duration/direct/badJSON",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "duration", "arg": "4h`,
			wantValue: Duration{},
			wantErr:   errJSONDecode,
		},
		{
			name: "duration/direct/badFn",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ "fn": "bad", "arg": "4h" }`,
			wantValue: Duration{},
			wantErr:   errJSONExtFnMatch,
		},

		{
			name: "duration/ExtNotFound",
			f: func(b []byte) (Value, error) {
				var res Duration
				err := (&res).UnmarshalJSON(b)
				return res, err
			},
			in:        `{ }`,
			wantValue: Duration{},
			wantErr:   errJSONExtNotFound,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotValue, gotErr := tt.f([]byte(tt.in))
			testutil.Equals(t, gotValue, tt.wantValue)
			testutil.ErrorIs(t, gotErr, tt.wantErr)
		})
	}
}

func TestJSONMarshal(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   Value
		out  string
	}{
		{
			"record",
			NewRecord(RecordMap{
				"ak": String("av"),
				"ck": String("cv"),
				"bk": String("bv"),
			}),
			`{"ak":"av","bk":"bv","ck":"cv"}`,
		},
		{
			"recordWithExt",
			NewRecord(RecordMap{
				"ip": mustIPValue("222.222.222.7"),
			}),
			`{"ip":{"__extn":{"fn":"ip","arg":"222.222.222.7"}}}`,
		},
		{
			"set",
			NewSet([]Value{
				String("av"),
				String("cv"),
				String("bv"),
			}),
			`["cv","bv","av"]`,
		},
		{
			"entity",
			EntityUID{"User", "alice"},
			`{"__entity":{"type":"User","id":"alice"}}`,
		},
		{
			"ip",
			mustIPValue("222.222.222.7"),
			`{"__extn":{"fn":"ip","arg":"222.222.222.7"}}`,
		},
		{
			"decimal",
			mustDecimalValue("33.57"),
			`{"__extn":{"fn":"decimal","arg":"33.57"}}`,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			out, err := json.Marshal(tt.in)
			testutil.OK(t, err)
			testutil.Equals(t, string(out), tt.out)
		})
	}
}

func TestJSONSet(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var s Set
		err := json.Unmarshal([]byte(`[{"__extn":{"fn":"err"}}]`), &s)
		testutil.Error(t, err)
	})
}

func TestJSONRecord(t *testing.T) {
	t.Parallel()
	t.Run("UnmarshalErr", func(t *testing.T) {
		t.Parallel()
		var r Record
		err := json.Unmarshal([]byte(`{"key":{"__extn":{"fn":"err"}}}`), &r)
		testutil.Error(t, err)
	})
	t.Run("MarshalKeyErrImpossible", func(t *testing.T) {
		t.Parallel()
		k := []byte{0xde, 0x01}
		m := RecordMap{String(k): Boolean(false)}
		r := NewRecord(m)
		v, err := json.Marshal(r)
		// this demonstrates that invalid keys will still result in json
		testutil.Equals(t, string(v), `{"\ufffd\u0001":false}`)
		testutil.OK(t, err)
	})
}
