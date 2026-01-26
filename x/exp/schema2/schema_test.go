package schema2_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema2/resolver"
)

var wantCedar = `
@doc("Address information")
type Address = {
  "city": String,
  "country": Country,
  "street": String,
  "zipcode"?: String,
};

type decimal = {
  "decimal": Long,
  "whole": Long,
};

entity Admin;

entity Role enum ["superuser", "operator"];

entity System in [Admin] = {
  "version": String,
};

action audit appliesTo {
  principal: [Admin],
  resource: [MyApp::Document, System],
};

@doc("Doc manager")
namespace MyApp {
  type Metadata = {
    "created": datetime,
    "tags": Set<String>,
  };

  entity Department = {
    "budget": decimal,
  };

  entity Document = {
    "public": Bool,
    "title": String,
  };

  entity Group in [Department] = {
    "metadata": Metadata,
    "name": String,
  };

  entity Status enum ["draft", "published", "archived"];

  @doc("User entity")
  entity User in [Group] = {
    "active": Bool,
    "address": Address,
    "email": String,
    "level": Long,
  };

  @doc("View or edit document")
  action edit appliesTo {
    principal: [User],
    resource: [Document],
    context: {
      "ip": ipaddr,
      "timestamp": datetime,
    }
  };

  action manage appliesTo {
    principal: [User],
    resource: [Document, Group],
  };

  @doc("View or edit document")
  action view appliesTo {
    principal: [User],
    resource: [Document],
    context: {
      "ip": ipaddr,
      "timestamp": datetime,
    }
  };
}
`

var wantJSON = `{
  "": {
    "commonTypes": {
      "Address": {
        "type": "Record",
        "attributes": {
          "city": {
            "type": "EntityOrCommon",
            "name": "String"
          },
          "country": {
            "type": "EntityOrCommon",
            "name": "Country"
          },
          "street": {
            "type": "EntityOrCommon",
            "name": "String"
          },
          "zipcode": {
            "type": "EntityOrCommon",
            "name": "String",
            "required": false
          }
        },
        "annotations": {
          "doc": "Address information"
        }
      },
      "decimal": {
        "type": "Record",
        "attributes": {
          "decimal": {
            "type": "EntityOrCommon",
            "name": "Long"
          },
          "whole": {
            "type": "EntityOrCommon",
            "name": "Long"
          }
        }
      }
    },
    "entityTypes": {
      "Admin": {},
      "Role": {
        "enum": [
          "superuser",
          "operator"
        ]
      },
      "System": {
        "memberOfTypes": [
          "Admin"
        ],
        "shape": {
          "type": "Record",
          "attributes": {
            "version": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      }
    },
    "actions": {
      "audit": {
        "appliesTo": {
          "resourceTypes": [
            "MyApp::Document",
            "System"
          ],
          "principalTypes": [
            "Admin"
          ]
        }
      }
    }
  },
  "MyApp": {
	"annotations": {
		"doc": "Doc manager"
	},
    "commonTypes": {
      "Metadata": {
        "type": "Record",
        "attributes": {
          "created": {
            "type": "EntityOrCommon",
            "name": "datetime"
          },
          "tags": {
            "type": "Set",
            "element": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      }
    },
    "entityTypes": {
      "Department": {
        "shape": {
          "type": "Record",
          "attributes": {
            "budget": {
              "type": "EntityOrCommon",
              "name": "decimal"
            }
          }
        }
      },
      "Document": {
        "shape": {
          "type": "Record",
          "attributes": {
            "public": {
              "type": "EntityOrCommon",
              "name": "Bool"
            },
            "title": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      },
      "Group": {
        "memberOfTypes": [
          "Department"
        ],
        "shape": {
          "type": "Record",
          "attributes": {
            "metadata": {
              "type": "EntityOrCommon",
              "name": "Metadata"
            },
            "name": {
              "type": "EntityOrCommon",
              "name": "String"
            }
          }
        }
      },
      "Status": {
        "enum": [
          "draft",
          "published",
          "archived"
        ]
      },
      "User": {
        "memberOfTypes": [
          "Group"
        ],
        "shape": {
          "type": "Record",
          "attributes": {
            "active": {
              "type": "EntityOrCommon",
              "name": "Bool"
            },
            "address": {
              "type": "EntityOrCommon",
              "name": "Address"
            },
            "email": {
              "type": "EntityOrCommon",
              "name": "String"
            },
            "level": {
              "type": "EntityOrCommon",
              "name": "Long"
            }
          }
        },
        "annotations": {
          "doc": "User entity"
        }
      }
    },
    "actions": {
      "edit": {
        "appliesTo": {
          "resourceTypes": [
            "Document"
          ],
          "principalTypes": [
            "User"
          ],
          "context": {
            "type": "Record",
            "attributes": {
              "ip": {
                "type": "EntityOrCommon",
                "name": "ipaddr"
              },
              "timestamp": {
                "type": "EntityOrCommon",
                "name": "datetime"
              }
            }
          }
        },
        "annotations": {
          "doc": "View or edit document"
        }
      },
      "manage": {
        "appliesTo": {
          "resourceTypes": [
            "Document",
            "Group"
          ],
          "principalTypes": [
            "User"
          ]
        }
      },
      "view": {
        "appliesTo": {
          "resourceTypes": [
            "Document"
          ],
          "principalTypes": [
            "User"
          ],
          "context": {
            "type": "Record",
            "attributes": {
              "ip": {
                "type": "EntityOrCommon",
                "name": "ipaddr"
              },
              "timestamp": {
                "type": "EntityOrCommon",
                "name": "datetime"
              }
            }
          }
        },
        "annotations": {
          "doc": "View or edit document"
        }
      }
    }
  }
}`

// wantAST is the expected AST structure for the test schema.
// All attributes are in alphabetical order to match the deterministic marshaling.
// All slices are initialized as empty slices (not nil) to match unmarshaling behavior.
var wantAST = &ast.Schema{
	Nodes: []ast.IsNode{
		ast.CommonTypeNode{
			Name: "Address",
			Annotations: []ast.Annotation{
				{Key: "doc", Value: "Address information"},
			},
			Type: ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "city", Type: ast.StringType{}},
					{Key: "country", Type: ast.TypeRef{Name: "Country"}},
					{Key: "street", Type: ast.StringType{}},
					{Key: "zipcode", Type: ast.StringType{}, Optional: true},
				},
			},
		},
		ast.CommonTypeNode{
			Name: "decimal",
			Type: ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "decimal", Type: ast.LongType{}},
					{Key: "whole", Type: ast.LongType{}},
				},
			},
		},
		// Top-level entities (alphabetically)
		ast.EntityNode{
			Name: "Admin",
		},
		ast.EnumNode{
			Name:   "Role",
			Values: []types.String{"superuser", "operator"},
		},
		ast.EntityNode{
			Name:        "System",
			MemberOfVal: []ast.EntityTypeRef{{Name: "Admin"}},
			ShapeVal: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "version", Type: ast.StringType{}},
				},
			},
		},
		ast.ActionNode{
			Name: "audit",
			AppliesToVal: &ast.AppliesTo{
				PrincipalTypes: []ast.EntityTypeRef{{Name: "Admin"}},
				ResourceTypes:  []ast.EntityTypeRef{{Name: "MyApp::Document"}, {Name: "System"}},
			},
		},
		// MyApp namespace
		ast.NamespaceNode{
			Name: "MyApp",
			Annotations: []ast.Annotation{
				{Key: "doc", Value: "Doc manager"},
			},
			Declarations: []ast.IsDeclaration{
				// Common types (alphabetically)
				ast.CommonTypeNode{
					Name: "Metadata",
					Type: ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "created", Type: ast.TypeRef{Name: "datetime"}},
							{Key: "tags", Type: ast.SetType{Element: ast.StringType{}}},
						},
					},
				},
				// Entities (alphabetically)
				ast.EntityNode{
					Name: "Department",
					ShapeVal: &ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "budget", Type: ast.TypeRef{Name: "decimal"}},
						},
					},
				},
				ast.EntityNode{
					Name: "Document",
					ShapeVal: &ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "public", Type: ast.BoolType{}},
							{Key: "title", Type: ast.StringType{}},
						},
					},
				},
				ast.EntityNode{
					Name:        "Group",
					MemberOfVal: []ast.EntityTypeRef{{Name: "Department"}},
					ShapeVal: &ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "metadata", Type: ast.TypeRef{Name: "Metadata"}},
							{Key: "name", Type: ast.StringType{}},
						},
					},
				},
				ast.EnumNode{
					Name:   "Status",
					Values: []types.String{"draft", "published", "archived"},
				},
				ast.EntityNode{
					Name:        "User",
					MemberOfVal: []ast.EntityTypeRef{{Name: "Group"}},
					Annotations: []ast.Annotation{
						{Key: "doc", Value: "User entity"},
					},
					ShapeVal: &ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "active", Type: ast.BoolType{}},
							{Key: "address", Type: ast.TypeRef{Name: "Address"}},
							{Key: "email", Type: ast.StringType{}},
							{Key: "level", Type: ast.LongType{}},
						},
					},
				},
				// Actions (alphabetically)
				ast.ActionNode{
					Name: "edit",
					Annotations: []ast.Annotation{
						{Key: "doc", Value: "View or edit document"},
					},
					AppliesToVal: &ast.AppliesTo{
						PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
						ResourceTypes:  []ast.EntityTypeRef{{Name: "Document"}},
						Context: ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "ip", Type: ast.TypeRef{Name: "ipaddr"}},
								{Key: "timestamp", Type: ast.TypeRef{Name: "datetime"}},
							},
						},
					},
				},
				ast.ActionNode{
					Name: "manage",
					AppliesToVal: &ast.AppliesTo{
						PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
						ResourceTypes:  []ast.EntityTypeRef{{Name: "Document"}, {Name: "Group"}},
					},
				},
				ast.ActionNode{
					Name: "view",
					Annotations: []ast.Annotation{
						{Key: "doc", Value: "View or edit document"},
					},
					AppliesToVal: &ast.AppliesTo{
						PrincipalTypes: []ast.EntityTypeRef{{Name: "User"}},
						ResourceTypes:  []ast.EntityTypeRef{{Name: "Document"}},
						Context: ast.RecordType{
							Pairs: []ast.Pair{
								{Key: "ip", Type: ast.TypeRef{Name: "ipaddr"}},
								{Key: "timestamp", Type: ast.TypeRef{Name: "datetime"}},
							},
						},
					},
				},
			},
		},
	},
}

// wantResolved is the expected resolved schema structure.
// All type references have been fully qualified.
var wantResolved = &resolver.ResolvedSchema{
	Namespaces: map[types.Path]resolver.ResolvedNamespace{
		"MyApp": {
			Name: "MyApp",
			Annotations: []ast.Annotation{
				{Key: "doc", Value: "Doc manager"},
			},
		},
	},
	Entities: map[types.EntityType]resolver.ResolvedEntity{
		"Admin": {
			Name:        "Admin",
			Annotations: nil,
			MemberOf:    nil,
			Shape:       nil,
			Tags:        nil,
		},
		"System": {
			Name:        "System",
			Annotations: nil,
			MemberOf:    []types.EntityType{"Admin"},
			Shape: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "version", Type: ast.StringType{}},
				},
			},
			Tags: nil,
		},
		"MyApp::Department": {
			Name:        "MyApp::Department",
			Annotations: nil,
			MemberOf:    nil,
			Shape: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "budget", Type: ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "decimal", Type: ast.LongType{}},
							{Key: "whole", Type: ast.LongType{}},
						},
					}},
				},
			},
			Tags: nil,
		},
		"MyApp::Document": {
			Name:        "MyApp::Document",
			Annotations: nil,
			MemberOf:    nil,
			Shape: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "public", Type: ast.BoolType{}},
					{Key: "title", Type: ast.StringType{}},
				},
			},
			Tags: nil,
		},
		"MyApp::Group": {
			Name:        "MyApp::Group",
			Annotations: nil,
			MemberOf:    []types.EntityType{"MyApp::Department"},
			Shape: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "metadata", Type: ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "created", Type: ast.ExtensionType{Name: "datetime"}},
							{Key: "tags", Type: ast.SetType{Element: ast.StringType{}}},
						},
					}},
					{Key: "name", Type: ast.StringType{}},
				},
			},
			Tags: nil,
		},
		"MyApp::User": {
			Name:        "MyApp::User",
			Annotations: []ast.Annotation{{Key: "doc", Value: "User entity"}},
			MemberOf:    []types.EntityType{"MyApp::Group"},
			Shape: &ast.RecordType{
				Pairs: []ast.Pair{
					{Key: "active", Type: ast.BoolType{}},
					{Key: "address", Type: ast.RecordType{
						Pairs: []ast.Pair{
							{Key: "city", Type: ast.StringType{}},
							{Key: "country", Type: ast.EntityTypeRef{Name: types.EntityType("Country")}},
							{Key: "street", Type: ast.StringType{}},
							{Key: "zipcode", Type: ast.StringType{}, Optional: true},
						},
					}},
					{Key: "email", Type: ast.StringType{}},
					{Key: "level", Type: ast.LongType{}},
				},
			},
			Tags: nil,
		},
	},
	Enums: map[types.EntityType]resolver.ResolvedEnum{
		"Role": {
			Name:        "Role",
			Annotations: nil,
			Values:      []types.String{"superuser", "operator"},
		},
		"MyApp::Status": {
			Name:        "MyApp::Status",
			Annotations: nil,
			Values:      []types.String{"draft", "published", "archived"},
		},
	},
	Actions: map[types.EntityUID]resolver.ResolvedAction{
		types.NewEntityUID("Action", "audit"): {
			Name:        "audit",
			Annotations: nil,
			MemberOf:    nil,
			AppliesTo: &resolver.ResolvedAppliesTo{
				PrincipalTypes: []types.EntityType{"Admin"},
				ResourceTypes:  []types.EntityType{"MyApp::Document", "System"},
				Context:        ast.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "edit"): {
			Name:        "edit",
			Annotations: []ast.Annotation{{Key: "doc", Value: "View or edit document"}},
			MemberOf:    nil,
			AppliesTo: &resolver.ResolvedAppliesTo{
				PrincipalTypes: []types.EntityType{"MyApp::User"},
				ResourceTypes:  []types.EntityType{"MyApp::Document"},
				Context: ast.RecordType{
					Pairs: []ast.Pair{
						{Key: "ip", Type: ast.ExtensionType{Name: "ipaddr"}},
						{Key: "timestamp", Type: ast.ExtensionType{Name: "datetime"}},
					},
				},
			},
		},
		types.NewEntityUID("MyApp::Action", "manage"): {
			Name:        "manage",
			Annotations: nil,
			MemberOf:    nil,
			AppliesTo: &resolver.ResolvedAppliesTo{
				PrincipalTypes: []types.EntityType{"MyApp::User"},
				ResourceTypes:  []types.EntityType{"MyApp::Document", "MyApp::Group"},
				Context:        ast.RecordType{},
			},
		},
		types.NewEntityUID("MyApp::Action", "view"): {
			Name:        "view",
			Annotations: []ast.Annotation{{Key: "doc", Value: "View or edit document"}},
			MemberOf:    nil,
			AppliesTo: &resolver.ResolvedAppliesTo{
				PrincipalTypes: []types.EntityType{"MyApp::User"},
				ResourceTypes:  []types.EntityType{"MyApp::Document"},
				Context: ast.RecordType{
					Pairs: []ast.Pair{
						{Key: "ip", Type: ast.ExtensionType{Name: "ipaddr"}},
						{Key: "timestamp", Type: ast.ExtensionType{Name: "datetime"}},
					},
				},
			},
		},
	},
}

func TestSchema(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalCedar", func(t *testing.T) {
		t.Parallel()
		var s schema2.Schema
		err := s.UnmarshalCedar([]byte(wantCedar))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()
		var s schema2.Schema
		err := s.UnmarshalJSON([]byte(wantJSON))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		s := schema2.NewSchemaFromAST(wantAST)
		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		stringEquals(t, string(b), wantCedar)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		s := schema2.NewSchemaFromAST(wantAST)
		b, err := s.MarshalJSON()
		testutil.OK(t, err)
		stringEquals(t, string(normalizeJSON(t, b)), string(normalizeJSON(t, []byte(wantJSON))))
	})

	t.Run("Resolve", func(t *testing.T) {
		t.Parallel()
		s := schema2.NewSchemaFromAST(wantAST)
		r, err := s.Resolve()
		testutil.OK(t, err)
		testutil.Equals(t, r, wantResolved)
	})
}

func stringEquals(t *testing.T, got, want string) {
	t.Helper()
	testutil.Equals(t, strings.TrimSpace(got), strings.TrimSpace(want))
}

func normalizeJSON(t *testing.T, in []byte) []byte {
	t.Helper()
	var out any
	err := json.Unmarshal(in, &out)
	testutil.OK(t, err)
	b, err := json.MarshalIndent(out, "", "  ")
	testutil.OK(t, err)
	return b
}
