package schema2_test

import (
	"archive/zip"
	"bytes"
	_ "embed"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

// These Cedar schema files were collected from the Rust reference implementation
// repository found at https://github.com/cedar-policy/cedar

//go:embed testdata.zip
var testdataZip []byte

func mustReadZip(t testing.TB, files map[string]*zip.File, name string) []byte {
	t.Helper()
	file, ok := files[name]
	if !ok {
		t.Fatalf("file %s not found in zip", name)
	}
	rc, err := file.Open()
	testutil.OK(t, err)
	defer func() { _ = rc.Close() }()
	var buf bytes.Buffer
	_, err = buf.ReadFrom(rc)
	testutil.OK(t, err)
	return buf.Bytes()
}

func TestData(t *testing.T) {
	t.Parallel()

	// Open the embedded zip file
	zipReader, err := zip.NewReader(bytes.NewReader(testdataZip), int64(len(testdataZip)))
	testutil.OK(t, err)

	// Create a map of files from the zip
	files := make(map[string]*zip.File)
	for _, file := range zipReader.File {
		files[file.Name] = file
	}

	// Find all .cedarschema files
	var schemas []string
	for name := range files {
		if len(name) > len("testdata/") && name[len(name)-13:] == ".cedarschema" {
			schemas = append(schemas, name)
		}
	}

	for _, name := range schemas {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			cedarBytes := mustReadZip(t, files, name)
			cedarBytes = bytes.ReplaceAll(cedarBytes, []byte("context: {}\n"), nil) // Rust converted JSON never contains the empty context record
			jsonBytes := mustReadZip(t, files, name+".json")
			jsonBytes = bytes.ReplaceAll(jsonBytes, []byte(`"appliesTo":{"resourceTypes":[],"principalTypes":[]}`), nil) // appliesTo is optional

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
				c2, err := jsonSchema.MarshalCedar()
				testutil.OK(t, err)
				var s1, s2 schema2.Schema
				err = s1.UnmarshalCedar(c1)
				testutil.OK(t, err)
				err = s2.UnmarshalCedar(c2)
				testutil.OK(t, err)
				j1, err := s1.MarshalJSON()
				testutil.OK(t, err)
				j2, err := s2.MarshalJSON()
				testutil.OK(t, err)
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
