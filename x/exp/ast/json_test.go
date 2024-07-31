package ast_test

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/ast"
)

func TestUnmarshalJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    *ast.Policy
		wantErr bool
	}{
		/*
		   @key("value")
		   permit (
		       principal == User::"12UA45",
		       action == Action::"view",
		       resource in Folder::"abc"
		   ) when {
		       context.tls_version == "1.3"
		   };
		*/
		{"exampleFromDocs", `{
	"annotations": {
		"key": "value"
	},
    "effect": "permit",
    "principal": {
        "op": "==",
        "entity": {
			"type": "User",
			"id": "12UA45"
		}
    },
    "action": {
        "op": "==",
        "entity": {
			"type": "Action",
			"id": "view"
		}
    },
    "resource": {
        "op": "in",
        "entity": {
			"type": "Folder",
			"id": "abc"
		}
    },
    "conditions": [
        {
            "kind": "when",
            "body": {
                "==": {
                    "left": {
                        ".": {
                            "left": {
                                "Var": "context"
                            },
                            "attr": "tls_version"
                        }
                    },
                    "right": {
                        "Value": "1.3"
                    }
                }
            }
        }
    ]
}`,
			ast.Permit().
				Annotate("key", "value").
				PrincipalEq(types.NewEntityUID("User", "12UA45")).
				ActionEq(types.NewEntityUID("Action", "view")).
				ResourceIn(types.NewEntityUID("Folder", "abc")).
				When(
					ast.Context().Access("tls_version").Equals(ast.String("1.3")),
				),
			false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var p ast.Policy
			fixedInput := strings.ReplaceAll(tt.input, "\t", "    ")
			err := json.Unmarshal([]byte(fixedInput), &p)
			if (err != nil) != tt.wantErr {
				t.Errorf("error got: %v want: %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(&p, tt.want) {
				t.Errorf("policy mismatch: got: %+v want: %+v", p, *tt.want)
			}

			b, err := json.MarshalIndent(&p, "", "    ")
			testutil.OK(t, err)
			testutil.Equals(t, string(b), fixedInput)
		})
	}

}