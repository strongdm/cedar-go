package schema_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
	"github.com/cedar-policy/cedar-go/x/exp/schema/resolver"
)

var wantCedar = `
@doc("Address information")
@personal_information
type Address = {
  @also("town")
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

entity System in [Admin] = {
  "version": String,
};

entity Role enum ["superuser", "operator"];

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

  @doc("User entity")
  entity User in [Group] = {
    "active": Bool,
    "address": Address,
    "email": String,
    "level": Long,
  };

  entity Status enum ["draft", "published", "archived"];

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
            "name": "String",
			"annotations": {
			  "also": "town"
			}
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
          "doc": "Address information",
		  "personal_information": ""
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
// All maps are initialized to match unmarshaling behavior.
var wantAST = &ast.Schema{
	CommonTypes: ast.CommonTypes{
		"Address": ast.CommonType{
			Annotations: ast.Annotations{
				"doc":                  "Address information",
				"personal_information": "",
			},
			Type: ast.RecordType{
				"city": ast.Attribute{
					Type: ast.StringType{},
					Annotations: ast.Annotations{
						"also": "town",
					},
				},
				"country": ast.Attribute{Type: ast.TypeRef("Country")},
				"street":  ast.Attribute{Type: ast.StringType{}},
				"zipcode": ast.Attribute{Type: ast.StringType{}, Optional: true},
			},
		},
		"decimal": ast.CommonType{
			Type: ast.RecordType{
				"decimal": ast.Attribute{Type: ast.LongType{}},
				"whole":   ast.Attribute{Type: ast.LongType{}},
			},
		},
	},
	Entities: ast.Entities{
		"Admin": ast.Entity{},
		"System": ast.Entity{
			MemberOf: []ast.EntityTypeRef{"Admin"},
			Shape: &ast.RecordType{
				"version": ast.Attribute{Type: ast.StringType{}},
			},
		},
	},
	Enums: ast.Enums{
		"Role": ast.Enum{
			Values: []types.String{"superuser", "operator"},
		},
	},
	Actions: ast.Actions{
		"audit": ast.Action{
			AppliesTo: &ast.AppliesTo{
				Principals: []ast.EntityTypeRef{"Admin"},
				Resources:  []ast.EntityTypeRef{"MyApp::Document", "System"},
			},
		},
	},
	Namespaces: ast.Namespaces{
		"MyApp": ast.Namespace{
			Annotations: ast.Annotations{
				"doc": "Doc manager",
			},
			CommonTypes: ast.CommonTypes{
				"Metadata": ast.CommonType{
					Type: ast.RecordType{
						"created": ast.Attribute{Type: ast.TypeRef("datetime")},
						"tags":    ast.Attribute{Type: ast.SetType{Element: ast.StringType{}}},
					},
				},
			},
			Entities: ast.Entities{
				"Department": ast.Entity{
					Shape: &ast.RecordType{
						"budget": ast.Attribute{Type: ast.TypeRef("decimal")},
					},
				},
				"Document": ast.Entity{
					Shape: &ast.RecordType{
						"public": ast.Attribute{Type: ast.BoolType{}},
						"title":  ast.Attribute{Type: ast.StringType{}},
					},
				},
				"Group": ast.Entity{
					MemberOf: []ast.EntityTypeRef{"Department"},
					Shape: &ast.RecordType{
						"metadata": ast.Attribute{Type: ast.TypeRef("Metadata")},
						"name":     ast.Attribute{Type: ast.StringType{}},
					},
				},
				"User": ast.Entity{
					MemberOf: []ast.EntityTypeRef{"Group"},
					Annotations: ast.Annotations{
						"doc": "User entity",
					},
					Shape: &ast.RecordType{
						"active":  ast.Attribute{Type: ast.BoolType{}},
						"address": ast.Attribute{Type: ast.TypeRef("Address")},
						"email":   ast.Attribute{Type: ast.StringType{}},
						"level":   ast.Attribute{Type: ast.LongType{}},
					},
				},
			},
			Enums: ast.Enums{
				"Status": ast.Enum{
					Values: []types.String{"draft", "published", "archived"},
				},
			},
			Actions: ast.Actions{
				"edit": ast.Action{
					Annotations: ast.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document"},
						Context: ast.RecordType{
							"ip":        ast.Attribute{Type: ast.TypeRef("ipaddr")},
							"timestamp": ast.Attribute{Type: ast.TypeRef("datetime")},
						},
					},
				},
				"manage": ast.Action{
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document", "Group"},
					},
				},
				"view": ast.Action{
					Annotations: ast.Annotations{
						"doc": "View or edit document",
					},
					AppliesTo: &ast.AppliesTo{
						Principals: []ast.EntityTypeRef{"User"},
						Resources:  []ast.EntityTypeRef{"Document"},
						Context: ast.RecordType{
							"ip":        ast.Attribute{Type: ast.TypeRef("ipaddr")},
							"timestamp": ast.Attribute{Type: ast.TypeRef("datetime")},
						},
					},
				},
			},
		},
	},
}

// wantResolved is the expected resolved schema structure.
// All type references have been fully qualified.
var wantResolved = &resolver.Schema{
	Namespaces: map[types.Path]resolver.Namespace{
		"MyApp": {
			Name: "MyApp",
			Annotations: resolver.Annotations{
				"doc": "Doc manager",
			},
		},
	},
	Entities: map[types.EntityType]resolver.Entity{
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
			Shape: &resolver.RecordType{
				"version": resolver.Attribute{Type: resolver.StringType{}},
			},
			Tags: nil,
		},
		"MyApp::Department": {
			Name:        "MyApp::Department",
			Annotations: nil,
			MemberOf:    nil,
			Shape: &resolver.RecordType{
				"budget": resolver.Attribute{
					Type: resolver.RecordType{
						"decimal": resolver.Attribute{Type: resolver.LongType{}},
						"whole":   resolver.Attribute{Type: resolver.LongType{}},
					},
				},
			},
			Tags: nil,
		},
		"MyApp::Document": {
			Name:        "MyApp::Document",
			Annotations: nil,
			MemberOf:    nil,
			Shape: &resolver.RecordType{
				"public": resolver.Attribute{Type: resolver.BoolType{}},
				"title":  resolver.Attribute{Type: resolver.StringType{}},
			},
			Tags: nil,
		},
		"MyApp::Group": {
			Name:        "MyApp::Group",
			Annotations: nil,
			MemberOf:    []types.EntityType{"MyApp::Department"},
			Shape: &resolver.RecordType{
				"metadata": resolver.Attribute{
					Type: resolver.RecordType{
						"created": resolver.Attribute{Type: resolver.ExtensionType("datetime")},
						"tags":    resolver.Attribute{Type: resolver.SetType{Element: resolver.StringType{}}},
					},
				},
				"name": resolver.Attribute{Type: resolver.StringType{}},
			},
			Tags: nil,
		},
		"MyApp::User": {
			Name:        "MyApp::User",
			Annotations: resolver.Annotations{"doc": "User entity"},
			MemberOf:    []types.EntityType{"MyApp::Group"},
			Shape: &resolver.RecordType{
				"active": resolver.Attribute{Type: resolver.BoolType{}},
				"address": resolver.Attribute{
					Type: resolver.RecordType{
						// TODO: include annotations of the type?
						"city": resolver.Attribute{
							Type: resolver.StringType{},
							Annotations: resolver.Annotations{
								"also": "town",
							},
						},
						"country": resolver.Attribute{Type: resolver.EntityTypeRef("Country")},
						"street":  resolver.Attribute{Type: resolver.StringType{}},
						"zipcode": resolver.Attribute{Type: resolver.StringType{}, Optional: true},
					},
				},
				"email": resolver.Attribute{Type: resolver.StringType{}},
				"level": resolver.Attribute{Type: resolver.LongType{}},
			},
			Tags: nil,
		},
	},
	Enums: map[types.EntityType]resolver.Enum{
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
	Actions: map[types.EntityUID]resolver.Action{
		types.NewEntityUID("Action", "audit"): {
			Name:        "audit",
			Annotations: nil,
			MemberOf:    nil,
			AppliesTo: &resolver.AppliesTo{
				Principals: []types.EntityType{"Admin"},
				Resources:  []types.EntityType{"MyApp::Document", "System"},
				Context:    nil,
			},
		},
		types.NewEntityUID("MyApp::Action", "edit"): {
			Name:        "edit",
			Annotations: resolver.Annotations{"doc": "View or edit document"},
			MemberOf:    nil,
			AppliesTo: &resolver.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolver.RecordType{
					"ip":        resolver.Attribute{Type: resolver.ExtensionType("ipaddr")},
					"timestamp": resolver.Attribute{Type: resolver.ExtensionType("datetime")},
				},
			},
		},
		types.NewEntityUID("MyApp::Action", "manage"): {
			Name:        "manage",
			Annotations: nil,
			MemberOf:    nil,
			AppliesTo: &resolver.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document", "MyApp::Group"},
				Context:    nil,
			},
		},
		types.NewEntityUID("MyApp::Action", "view"): {
			Name:        "view",
			Annotations: resolver.Annotations{"doc": "View or edit document"},
			MemberOf:    nil,
			AppliesTo: &resolver.AppliesTo{
				Principals: []types.EntityType{"MyApp::User"},
				Resources:  []types.EntityType{"MyApp::Document"},
				Context: resolver.RecordType{
					"ip":        resolver.Attribute{Type: resolver.ExtensionType("ipaddr")},
					"timestamp": resolver.Attribute{Type: resolver.ExtensionType("datetime")},
				},
			},
		},
	},
}

func TestSchema(t *testing.T) {
	t.Parallel()

	t.Run("UnmarshalCedar", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalCedar([]byte(wantCedar))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("UnmarshalJSON", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalJSON([]byte(wantJSON))
		testutil.OK(t, err)
		testutil.Equals(t, s.AST(), wantAST)
	})

	t.Run("MarshalCedar", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		b, err := s.MarshalCedar()
		testutil.OK(t, err)
		stringEquals(t, string(b), wantCedar)
	})

	t.Run("MarshalJSON", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		b, err := s.MarshalJSON()
		testutil.OK(t, err)
		stringEquals(t, string(normalizeJSON(t, b)), string(normalizeJSON(t, []byte(wantJSON))))
	})

	t.Run("Resolve", func(t *testing.T) {
		t.Parallel()
		s := schema.NewSchemaFromAST(wantAST)
		r, err := s.Resolve()
		testutil.OK(t, err)
		testutil.Equals(t, r, wantResolved)
	})

	t.Run("UnmarshalCedarErr", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		const filename = "path/to/my-file-name.cedarschema"
		s.SetFilename(filename)
		err := s.UnmarshalCedar([]byte("LSKJDFN"))
		testutil.Error(t, err)
		testutil.FatalIf(t, !strings.Contains(err.Error(), filename+":1:1"), "expected %q in error: %v", filename, err)
	})

	t.Run("UnmarshalJSONErr", func(t *testing.T) {
		t.Parallel()
		var s schema.Schema
		err := s.UnmarshalJSON([]byte("LSKJDFN"))
		testutil.Error(t, err)
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
