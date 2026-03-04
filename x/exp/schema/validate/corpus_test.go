package validate_test

import (
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go"
	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/validate"
)

//go:embed testdata
var testdataFS embed.FS

type corpusTest struct {
	Schema         string `json:"schema"`
	Policies       string `json:"policies"`
	ShouldValidate bool   `json:"shouldValidate"`
	Entities       string `json:"entities"`
	Requests       []struct {
		Desc      string          `json:"description"`
		Principal cedar.EntityUID `json:"principal"`
		Action    cedar.EntityUID `json:"action"`
		Resource  cedar.EntityUID `json:"resource"`
		Context   cedar.Record    `json:"context"`
	} `json:"requests"`
}

type corpusValidationMode struct {
	Strict           bool     `json:"strict"`
	Permissive       bool     `json:"permissive"`
	StrictErrors     []string `json:"strictErrors"`
	PermissiveErrors []string `json:"permissiveErrors"`
}

type corpusValidation struct {
	PolicyValidation corpusValidationMode `json:"policyValidation"`
	EntityValidation corpusValidationMode `json:"entityValidation"`
	RequestValidation []struct {
		Description string   `json:"description"`
		Strict      *bool    `json:"strict"`
		Permissive  *bool    `json:"permissive"`
		Errors      []string `json:"errors"`
	} `json:"requestValidation"`
}

// countErrors recursively counts leaf errors in a (possibly joined) error.
func countErrors(err error) int {
	if err == nil {
		return 0
	}
	if ue, ok := err.(interface{ Unwrap() []error }); ok {
		n := 0
		for _, e := range ue.Unwrap() {
			n += countErrors(e)
		}
		return n
	}
	return 1
}

func TestCorpus(t *testing.T) {
	t.Parallel()

	entries, err := testdataFS.ReadDir("testdata")
	testutil.OK(t, err)

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".json") {
			continue
		}
		if strings.HasSuffix(name, ".entities.json") || strings.HasSuffix(name, ".validation.json") {
			continue
		}

		testName := strings.TrimSuffix(name, ".json")
		t.Run(testName, func(t *testing.T) {
			t.Parallel()

			// Load test manifest
			manifestData, err := testdataFS.ReadFile("testdata/" + name)
			testutil.OK(t, err)
			var tt corpusTest
			testutil.OK(t, json.Unmarshal(manifestData, &tt))

			// Load validation expectations
			validationData, err := testdataFS.ReadFile("testdata/" + testName + ".validation.json")
			testutil.OK(t, err)
			var cv corpusValidation
			testutil.OK(t, json.Unmarshal(validationData, &cv))

			// Load and parse schema
			schemaContent, err := testdataFS.ReadFile(tt.Schema)
			testutil.OK(t, err)
			var s schema.Schema
			s.SetFilename(testName + ".cedarschema")
			testutil.OK(t, s.UnmarshalCedar(schemaContent))
			rs, err := s.Resolve()
			testutil.OK(t, err)

			// Load and parse policies
			policyContent, err := testdataFS.ReadFile(tt.Policies)
			testutil.OK(t, err)
			policySet, err := cedar.NewPolicySetFromBytes(testName+".cedar", policyContent)
			testutil.OK(t, err)

			// Load entities
			entitiesContent, err := testdataFS.ReadFile(tt.Entities)
			testutil.OK(t, err)
			var entities cedar.EntityMap
			testutil.OK(t, json.Unmarshal(entitiesContent, &entities))

			// Validate policies
			t.Run("validate-policy-strict", func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithStrict())
				totalErrors := 0
				for _, p := range policySet.All() {
					if err := v.Policy((*ast.Policy)(p.AST())); err != nil {
						totalErrors += countErrors(err)
					}
				}
				testutil.Equals(t, totalErrors == 0, cv.PolicyValidation.Strict)
				testutil.Equals(t, totalErrors, len(cv.PolicyValidation.StrictErrors))
			})

			t.Run("validate-policy-permissive", func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithPermissive())
				totalErrors := 0
				for _, p := range policySet.All() {
					if err := v.Policy((*ast.Policy)(p.AST())); err != nil {
						totalErrors += countErrors(err)
					}
				}
				testutil.Equals(t, totalErrors == 0, cv.PolicyValidation.Permissive)
				testutil.Equals(t, totalErrors, len(cv.PolicyValidation.PermissiveErrors))
			})

			// Validate entities
			t.Run("validate-entities-strict", func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithStrict())
				err := v.Entities(entities)
				testutil.Equals(t, err == nil, cv.EntityValidation.Strict)
				testutil.Equals(t, countErrors(err), len(cv.EntityValidation.StrictErrors))
			})

			t.Run("validate-entities-permissive", func(t *testing.T) {
				t.Parallel()
				v := validate.New(rs, validate.WithPermissive())
				err := v.Entities(entities)
				testutil.Equals(t, err == nil, cv.EntityValidation.Permissive)
				testutil.Equals(t, countErrors(err), len(cv.EntityValidation.PermissiveErrors))
			})

			// Validate requests
			for i, reqVal := range cv.RequestValidation {
				if reqVal.Strict != nil {
					t.Run(fmt.Sprintf("validate-request-strict/%s", reqVal.Description), func(t *testing.T) {
						t.Parallel()
						v := validate.New(rs, validate.WithStrict())
						req := cedar.Request{
							Principal: tt.Requests[i].Principal,
							Action:    tt.Requests[i].Action,
							Resource:  tt.Requests[i].Resource,
							Context:   tt.Requests[i].Context,
						}
						err := v.Request(req)
						testutil.Equals(t, err == nil, *reqVal.Strict)
						testutil.Equals(t, countErrors(err), len(reqVal.Errors))
					})
				}
				if reqVal.Permissive != nil {
					t.Run(fmt.Sprintf("validate-request-permissive/%s", reqVal.Description), func(t *testing.T) {
						t.Parallel()
						v := validate.New(rs, validate.WithPermissive())
						req := cedar.Request{
							Principal: tt.Requests[i].Principal,
							Action:    tt.Requests[i].Action,
							Resource:  tt.Requests[i].Resource,
							Context:   tt.Requests[i].Context,
						}
						err := v.Request(req)
						testutil.Equals(t, err == nil, *reqVal.Permissive)
						testutil.Equals(t, countErrors(err), len(reqVal.Errors))
					})
				}
			}
		})
	}
}
