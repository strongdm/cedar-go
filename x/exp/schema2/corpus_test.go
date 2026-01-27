package schema2_test

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

//go:embed testdata/*
var testdata embed.FS

func mustRead(t testing.TB, src fs.FS, name string) []byte {
	t.Helper()
	out, err := fs.ReadFile(src, name)
	testutil.OK(t, err)
	return out
}

func TestCorpus(t *testing.T) {
	t.Parallel()

	schemas, err := fs.Glob(testdata, "testdata/*.cedarschema")
	testutil.OK(t, err)
	for _, name := range schemas {
		if name != "testdata/langserver_policies.cedarschema" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cedarBytes := mustRead(t, testdata, name)
			cedarBytes = bytes.ReplaceAll(cedarBytes, []byte("context: {}\n"), nil) // Rust converted JSON never contains the empty context record
			jsonBytes := mustRead(t, testdata, strings.ReplaceAll(name, ".cedarschema", ".json"))
			jsonBytes = bytes.ReplaceAll(jsonBytes, []byte(`"appliesTo":{"resourceTypes":[],"principalTypes":[]}`), nil) // appliesTo is optional
			// for _, ext := range []string{"duration", "datetime", "decimal", "ipaddr"} {
			// 	jsonBytes = bytes.ReplaceAll(jsonBytes,
			// 		[]byte(`{"type":"EntityOrCommon","name":"__cedar::`+ext+`"}}`),
			// 		[]byte(`{"type":"Extension","name":"`+ext+`"}}`),
			// 	)
			// }

			// UnmarshalCedar
			var cedarSchema schema2.Schema
			err := cedarSchema.UnmarshalCedar(cedarBytes)
			testutil.OK(t, err)

			// UnmarshalJSON
			var jsonSchema schema2.Schema
			err = jsonSchema.UnmarshalJSON(jsonBytes)
			testutil.OK(t, err)

			// MarshalCedar
			{
				c1, err := cedarSchema.MarshalCedar()
				testutil.OK(t, err)
				fmt.Println("c1==============", string(c1))
				c2, err := jsonSchema.MarshalCedar()
				testutil.OK(t, err)
				fmt.Println("c2===============", string(c2))
				// round-trip to JSON so we can do a comparison
				var s1, s2 schema2.Schema
				err = s1.UnmarshalCedar(c1)
				testutil.OK(t, err)
				err = s2.UnmarshalCedar(c2)
				testutil.OK(t, err)
				j1, err := s1.MarshalJSON()
				testutil.OK(t, err)
				j2, err := s2.MarshalJSON()
				testutil.OK(t, err)
				fmt.Println("j1===================", string(normalizeJSON(t, j1)))
				fmt.Println("j2===================", string(normalizeJSON(t, j2)))
				stringEquals(t, string(normalizeJSON(t, j1)), string(normalizeJSON(t, jsonBytes)))
				stringEquals(t, string(normalizeJSON(t, j2)), string(normalizeJSON(t, jsonBytes)))
			}

			// MarshalJSON
			{
				j1, err := cedarSchema.MarshalJSON()
				testutil.OK(t, err)
				j2, err := jsonSchema.MarshalJSON()
				testutil.OK(t, err)
				stringEquals(t, string(normalizeJSON(t, j1)), string(normalizeJSON(t, jsonBytes)))
				stringEquals(t, string(normalizeJSON(t, j2)), string(normalizeJSON(t, jsonBytes)))
			}

			// Resolve
			{
				r1, err := cedarSchema.Resolve()
				testutil.OK(t, err)
				r2, err := cedarSchema.Resolve()
				testutil.OK(t, err)
				testutil.Equals(t, r1, r2)
			}
		})
	}
}
