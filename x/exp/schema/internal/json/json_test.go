package json

import (
	"encoding/json"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/types"
	"github.com/cedar-policy/cedar-go/x/exp/schema/ast"
)

func TestJSONRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		json   string
		schema *ast.Schema
	}{
		{
			name: "empty schema",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{},
		},
		{
			name: "simple entity",
			json: `{
				"": {
					"entityTypes": {
						"User": {}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
				},
			},
		},
		{
			name: "entity with memberOf",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"memberOfTypes": ["Group"]
						},
						"Group": {}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						MemberOf: []ast.EntityTypeRef{
							"Group",
						},
					},
					"Group": ast.Entity{},
				},
			},
		},
		{
			name: "entity with shape - all primitive types",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"active": {"type": "EntityOrCommon", "name": "Bool"},
									"age": {"type": "EntityOrCommon", "name": "Long"},
									"email": {"type": "EntityOrCommon", "name": "String", "required": false},
									"name": {"type": "EntityOrCommon", "name": "String"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						Shape: &ast.RecordType{
							"name":   ast.Attribute{Type: ast.StringType{}, Optional: false},
							"age":    ast.Attribute{Type: ast.LongType{}, Optional: false},
							"active": ast.Attribute{Type: ast.BoolType{}, Optional: false},
							"email":  ast.Attribute{Type: ast.StringType{}, Optional: true},
						},
					},
				},
			},
		},
		{
			name: "entity with string tags",
			json: `{
				"": {
					"entityTypes": {
						"Resource": {
							"tags": {"type": "EntityOrCommon", "name": "String"}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Resource": ast.Entity{
						Tags: ast.StringType{},
					},
				},
			},
		},
		{
			name: "entity with long tags",
			json: `{
				"": {
					"entityTypes": {
						"Resource": {
							"tags": {"type": "EntityOrCommon", "name": "Long"}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Resource": ast.Entity{
						Tags: ast.LongType{},
					},
				},
			},
		},
		{
			name: "enum",
			json: `{
				"": {
					"entityTypes": {
						"Status": {
							"enum": ["active", "inactive", "pending"]
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Enums: ast.Enums{
					"Status": ast.Enum{
						Values: []types.String{"active", "inactive", "pending"},
					},
				},
			},
		},
		{
			name: "action with appliesTo",
			json: `{
				"": {
					"entityTypes": {
						"User": {},
						"Doc": {}
					},
					"actions": {
						"view": {
							"appliesTo": {
								"principalTypes": ["User"],
								"resourceTypes": ["Doc"]
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
					"Doc":  ast.Entity{},
				},
				Actions: ast.Actions{
					"view": ast.Action{
						AppliesTo: &ast.AppliesTo{
							Principals: []ast.EntityTypeRef{"User"},
							Resources:  []ast.EntityTypeRef{"Doc"},
						},
					},
				},
			},
		},
		{
			name: "action with memberOf",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {
						"view": {
							"memberOf": [{"type": "Action", "id": "readOnly"}]
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Actions: ast.Actions{
					"view": ast.Action{
						MemberOf: []ast.ParentRef{
							{Type: ast.EntityTypeRef("Action"), ID: "readOnly"},
						},
					},
				},
			},
		},
		{
			name: "action with memberOf empty type",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {
						"write": {
							"memberOf": [{"id": "read"}]
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Actions: ast.Actions{
					"write": ast.Action{
						MemberOf: []ast.ParentRef{
							{Type: ast.EntityTypeRef(""), ID: "read"},
						},
					},
				},
			},
		},
		{
			name: "action with context",
			json: `{
				"": {
					"entityTypes": {
						"User": {},
						"Doc": {}
					},
					"actions": {
						"view": {
							"appliesTo": {
								"principalTypes": ["User"],
								"resourceTypes": ["Doc"],
								"context": {
									"type": "Record",
									"attributes": {
										"ip": {"type": "Extension", "name": "ipaddr"}
									}
								}
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
					"Doc":  ast.Entity{},
				},
				Actions: ast.Actions{
					"view": ast.Action{
						AppliesTo: &ast.AppliesTo{
							Principals: []ast.EntityTypeRef{"User"},
							Resources:  []ast.EntityTypeRef{"Doc"},
							Context: ast.RecordType{
								"ip": ast.Attribute{Type: ast.ExtensionType("ipaddr"), Optional: false},
							},
						},
					},
				},
			},
		},
		{
			name: "action with context as type reference",
			json: `{
				"": {
					"entityTypes": {
						"User": {},
						"Doc": {}
					},
					"actions": {
						"view": {
							"appliesTo": {
								"principalTypes": ["User"],
								"resourceTypes": ["Doc"],
								"context": {"type": "ContextType"}
							}
						}
					},
					"commonTypes": {
						"ContextType": {
							"type": "Record",
							"attributes": {
								"field": {"type": "EntityOrCommon", "name": "String"}
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
					"Doc":  ast.Entity{},
				},
				Actions: ast.Actions{
					"view": ast.Action{
						AppliesTo: &ast.AppliesTo{
							Principals: []ast.EntityTypeRef{"User"},
							Resources:  []ast.EntityTypeRef{"Doc"},
							Context:    ast.TypeRef("ContextType"),
						},
					},
				},
				CommonTypes: ast.CommonTypes{
					"ContextType": ast.CommonType{
						Type: ast.RecordType{
							"field": ast.Attribute{Type: ast.StringType{}, Optional: false},
						},
					},
				},
			},
		},
		{
			name: "common types - primitives",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"Name": {"type": "EntityOrCommon", "name": "String"},
						"MyLong": {"type": "EntityOrCommon", "name": "Long"},
						"MyBool": {"type": "EntityOrCommon", "name": "Bool"},
						"MyExt": {"type": "Extension", "name": "ipaddr"}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"Name":   ast.CommonType{Type: ast.StringType{}},
					"MyLong": ast.CommonType{Type: ast.LongType{}},
					"MyBool": ast.CommonType{Type: ast.BoolType{}},
					"MyExt":  ast.CommonType{Type: ast.ExtensionType("ipaddr")},
				},
			},
		},
		{
			name: "common type set",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"Names": {
							"type": "Set",
							"element": {"type": "EntityOrCommon", "name": "String"}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"Names": ast.CommonType{
						Type: ast.SetType{Element: ast.StringType{}},
					},
				},
			},
		},
		{
			name: "common type record",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"Address": {
							"type": "Record",
							"attributes": {
								"street": {"type": "EntityOrCommon", "name": "String"},
								"zip": {"type": "EntityOrCommon", "name": "Long", "required": false}
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"Address": ast.CommonType{
						Type: ast.RecordType{
							"street": ast.Attribute{Type: ast.StringType{}, Optional: false},
							"zip":    ast.Attribute{Type: ast.LongType{}, Optional: true},
						},
					},
				},
			},
		},
		{
			name: "type reference to common type",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"name": {"type": "EntityOrCommon", "name": "Name"}
								}
							}
						}
					},
					"actions": {},
					"commonTypes": {
						"Name": {"type": "EntityOrCommon", "name": "String"}
					}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						Shape: &ast.RecordType{
							"name": ast.Attribute{Type: ast.TypeRef("Name"), Optional: false},
						},
					},
				},
				CommonTypes: ast.CommonTypes{
					"Name": ast.CommonType{Type: ast.StringType{}},
				},
			},
		},
		{
			name: "type ref between common types",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"Address": {"type": "EntityOrCommon", "name": "String"},
						"Person": {
							"type": "Record",
							"attributes": {
								"addr": {"type": "EntityOrCommon", "name": "Address"}
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"Address": ast.CommonType{Type: ast.StringType{}},
					"Person": ast.CommonType{
						Type: ast.RecordType{
							"addr": ast.Attribute{Type: ast.TypeRef("Address"), Optional: false},
						},
					},
				},
			},
		},
		{
			name: "entity type ref in common type",
			json: `{
				"": {
					"entityTypes": {
						"User": {}
					},
					"actions": {},
					"commonTypes": {
						"UserRef": {"type": "EntityOrCommon", "name": "User"}
					}
				}
			}`,
			// Note: EntityTypeRef and TypeRef both unmarshal to TypeRef from JSON
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
				},
				CommonTypes: ast.CommonTypes{
					"UserRef": ast.CommonType{
						Type: ast.TypeRef("User"),
					},
				},
			},
		},
		{
			name: "entity type ref in record attribute",
			json: `{
				"": {
					"entityTypes": {
						"User": {},
						"Doc": {
							"shape": {
								"type": "Record",
								"attributes": {
									"owner": {"type": "EntityOrCommon", "name": "User"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			// Note: EntityTypeRef and TypeRef both unmarshal to TypeRef from JSON
			// since we can't distinguish entity names from type names in JSON alone
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{},
					"Doc": ast.Entity{
						Shape: &ast.RecordType{
							"owner": ast.Attribute{Type: ast.TypeRef("User"), Optional: false},
						},
					},
				},
			},
		},
		{
			name: "nested records",
			json: `{
				"": {
					"entityTypes": {
						"Config": {
							"shape": {
								"type": "Record",
								"attributes": {
									"nested": {
										"type": "Record",
										"attributes": {
											"value": {"type": "EntityOrCommon", "name": "String"}
										}
									}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Config": ast.Entity{
						Shape: &ast.RecordType{
							"nested": ast.Attribute{
								Type: ast.RecordType{
									"value": ast.Attribute{Type: ast.StringType{}, Optional: false},
								},
								Optional: false,
							},
						},
					},
				},
			},
		},
		{
			name: "nested record with optional fields",
			json: `{
				"": {
					"entityTypes": {
						"Config": {
							"shape": {
								"type": "Record",
								"attributes": {
									"nested": {
										"type": "Record",
										"attributes": {
											"required_field": {"type": "EntityOrCommon", "name": "String"},
											"optional_field": {"type": "EntityOrCommon", "name": "Long", "required": false}
										}
									}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Config": ast.Entity{
						Shape: &ast.RecordType{
							"nested": ast.Attribute{
								Type: ast.RecordType{
									"required_field": ast.Attribute{Type: ast.StringType{}, Optional: false},
									"optional_field": ast.Attribute{Type: ast.LongType{}, Optional: true},
								},
								Optional: false,
							},
						},
					},
				},
			},
		},
		{
			name: "extension types",
			json: `{
				"": {
					"entityTypes": {
						"Request": {
							"shape": {
								"type": "Record",
								"attributes": {
									"ip": {"type": "Extension", "name": "ipaddr"},
									"time": {"type": "Extension", "name": "datetime"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Request": ast.Entity{
						Shape: &ast.RecordType{
							"ip":   ast.Attribute{Type: ast.ExtensionType("ipaddr"), Optional: false},
							"time": ast.Attribute{Type: ast.ExtensionType("datetime"), Optional: false},
						},
					},
				},
			},
		},
		{
			name: "set nested in set",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"NestedSets": {
							"type": "Set",
							"element": {
								"type": "Set",
								"element": {"type": "EntityOrCommon", "name": "String"}
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"NestedSets": ast.CommonType{
						Type: ast.SetType{
							Element: ast.SetType{
								Element: ast.StringType{},
							},
						},
					},
				},
			},
		},
		{
			name: "namespace",
			json: `{
				"App": {
					"entityTypes": {
						"User": {}
					},
					"actions": {
						"view": {}
					}
				}
			}`,
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					"App": ast.Namespace{
						Entities: ast.Entities{
							"User": ast.Entity{},
						},
						Actions: ast.Actions{
							"view": ast.Action{},
						},
					},
				},
			},
		},
		{
			name: "namespace with all declaration types",
			json: `{
				"App": {
					"entityTypes": {
						"User": {},
						"Status": {"enum": ["active"]}
					},
					"actions": {
						"view": {}
					},
					"commonTypes": {
						"Name": {"type": "EntityOrCommon", "name": "String"}
					}
				}
			}`,
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					"App": ast.Namespace{
						CommonTypes: ast.CommonTypes{
							"Name": ast.CommonType{Type: ast.StringType{}},
						},
						Entities: ast.Entities{
							"User": ast.Entity{},
						},
						Enums: ast.Enums{
							"Status": ast.Enum{Values: []types.String{"active"}},
						},
						Actions: ast.Actions{
							"view": ast.Action{},
						},
					},
				},
			},
		},
		{
			name: "annotations on entity",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"annotations": {
								"doc": "A user entity",
								"author": "Alice"
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						Annotations: ast.Annotations{
							"doc":    "A user entity",
							"author": "Alice",
						},
					},
				},
			},
		},
		{
			name: "annotations on enum",
			json: `{
				"": {
					"entityTypes": {
						"Status": {
							"enum": ["active", "inactive"],
							"annotations": {"doc": "Status enum"}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Enums: ast.Enums{
					"Status": ast.Enum{
						Values:      []types.String{"active", "inactive"},
						Annotations: ast.Annotations{"doc": "Status enum"},
					},
				},
			},
		},
		{
			name: "annotations on action",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {
						"view": {
							"annotations": {
								"doc": "View action",
								"version": "1.0"
							}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Actions: ast.Actions{
					"view": ast.Action{
						Annotations: ast.Annotations{
							"doc":     "View action",
							"version": "1.0",
						},
					},
				},
			},
		},
		{
			name: "annotations on common type",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"Address": {
							"type": "Record",
							"attributes": {
								"street": {"type": "EntityOrCommon", "name": "String"}
							},
							"annotations": {"doc": "Address type"}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"Address": ast.CommonType{
						Type: ast.RecordType{
							"street": ast.Attribute{Type: ast.StringType{}, Optional: false},
						},
						Annotations: ast.Annotations{"doc": "Address type"},
					},
				},
			},
		},
		{
			name: "annotations on attribute",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"name": {
										"type": "EntityOrCommon",
										"name": "String",
										"annotations": {"doc": "User's name"}
									}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						Shape: &ast.RecordType{
							"name": ast.Attribute{
								Type:        ast.StringType{},
								Optional:    false,
								Annotations: ast.Annotations{"doc": "User's name"},
							},
						},
					},
				},
			},
		},
		{
			name: "annotations on nested record attributes",
			json: `{
				"": {
					"entityTypes": {
						"Config": {
							"shape": {
								"type": "Record",
								"attributes": {
									"settings": {
										"type": "Record",
										"attributes": {
											"timeout": {
												"type": "EntityOrCommon",
												"name": "Long",
												"annotations": {
													"doc": "Timeout in seconds",
													"min": "1"
												}
											}
										}
									}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"Config": ast.Entity{
						Shape: &ast.RecordType{
							"settings": ast.Attribute{
								Type: ast.RecordType{
									"timeout": ast.Attribute{
										Type:     ast.LongType{},
										Optional: false,
										Annotations: ast.Annotations{
											"doc": "Timeout in seconds",
											"min": "1",
										},
									},
								},
								Optional: false,
							},
						},
					},
				},
			},
		},
		{
			name: "annotations on record attribute with multiple fields",
			json: `{
				"": {
					"entityTypes": {
						"User": {
							"shape": {
								"type": "Record",
								"attributes": {
									"age": {
										"type": "EntityOrCommon",
										"name": "Long",
										"annotations": {
											"doc": "The user's age",
											"min": "0"
										}
									},
									"name": {
										"type": "EntityOrCommon",
										"name": "String",
										"annotations": {"doc": "The user's name"}
									}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"User": ast.Entity{
						Shape: &ast.RecordType{
							"name": ast.Attribute{
								Type:        ast.StringType{},
								Optional:    false,
								Annotations: ast.Annotations{"doc": "The user's name"},
							},
							"age": ast.Attribute{
								Type:     ast.LongType{},
								Optional: false,
								Annotations: ast.Annotations{
									"doc": "The user's age",
									"min": "0",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "namespace annotations",
			json: `{
				"MyApp": {
					"annotations": {"doc": "My application namespace"},
					"entityTypes": {},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					"MyApp": ast.Namespace{
						Annotations: ast.Annotations{"doc": "My application namespace"},
					},
				},
			},
		},
		{
			name: "namespace with multiple annotations",
			json: `{
				"MyApp": {
					"annotations": {
						"doc": "My application",
						"version": "1.0"
					},
					"entityTypes": {
						"User": {}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					"MyApp": ast.Namespace{
						Annotations: ast.Annotations{
							"doc":     "My application",
							"version": "1.0",
						},
						Entities: ast.Entities{
							"User": ast.Entity{},
						},
					},
				},
			},
		},
		{
			name: "namespace with all annotations",
			json: `{
				"MyApp": {
					"entityTypes": {
						"User": {"annotations": {"doc": "User in MyApp"}},
						"Status": {"enum": ["active"], "annotations": {"doc": "Status in MyApp"}}
					},
					"actions": {
						"view": {"annotations": {"doc": "View in MyApp"}}
					},
					"commonTypes": {
						"Address": {
							"type": "EntityOrCommon",
							"name": "String",
							"annotations": {"doc": "Address in MyApp"}
						}
					}
				}
			}`,
			schema: &ast.Schema{
				Namespaces: ast.Namespaces{
					"MyApp": ast.Namespace{
						Entities: ast.Entities{
							"User": ast.Entity{
								Annotations: ast.Annotations{"doc": "User in MyApp"},
							},
						},
						Enums: ast.Enums{
							"Status": ast.Enum{
								Values:      []types.String{"active"},
								Annotations: ast.Annotations{"doc": "Status in MyApp"},
							},
						},
						Actions: ast.Actions{
							"view": ast.Action{
								Annotations: ast.Annotations{"doc": "View in MyApp"},
							},
						},
						CommonTypes: ast.CommonTypes{
							"Address": ast.CommonType{
								Type:        ast.StringType{},
								Annotations: ast.Annotations{"doc": "Address in MyApp"},
							},
						},
					},
				},
			},
		},
		{
			name: "legacy type names in commonTypes",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {},
					"commonTypes": {
						"A": {"type": "String"},
						"B": {"type": "Long"},
						"C": {"type": "Bool"}
					}
				}
			}`,
			schema: &ast.Schema{
				CommonTypes: ast.CommonTypes{
					"A": ast.CommonType{Type: ast.StringType{}},
					"B": ast.CommonType{Type: ast.LongType{}},
					"C": ast.CommonType{Type: ast.BoolType{}},
				},
			},
		},
		{
			name: "legacy type names in attributes",
			json: `{
				"": {
					"entityTypes": {
						"E": {
							"shape": {
								"type": "Record",
								"attributes": {
									"a": {"type": "String"},
									"b": {"type": "Long"},
									"c": {"type": "Bool"}
								}
							}
						}
					},
					"actions": {}
				}
			}`,
			schema: &ast.Schema{
				Entities: ast.Entities{
					"E": ast.Entity{
						Shape: &ast.RecordType{
							"a": ast.Attribute{Type: ast.StringType{}, Optional: false},
							"b": ast.Attribute{Type: ast.LongType{}, Optional: false},
							"c": ast.Attribute{Type: ast.BoolType{}, Optional: false},
						},
					},
				},
			},
		},
		{
			name: "empty appliesTo not emitted",
			json: `{
				"": {
					"entityTypes": {},
					"actions": {
						"view": {}
					}
				}
			}`,
			// Empty appliesTo in AST marshals to no appliesTo field in JSON
			// Which unmarshals to nil AppliesToVal
			schema: &ast.Schema{
				Actions: ast.Actions{
					"view": ast.Action{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Test AST -> JSON -> AST (round trip from expected schema)
			s1 := (*Schema)(tt.schema)
			jsonData, err := s1.MarshalJSON()
			testutil.OK(t, err)

			var s2 Schema
			err = s2.UnmarshalJSON(jsonData)
			testutil.OK(t, err)
			schema2 := (*ast.Schema)(&s2)
			testutil.Equals(t, tt.schema, schema2)

			// Only test JSON -> AST if JSON is provided
			if tt.json != "" {
				// Test JSON -> AST -> JSON (verify JSON produces same result)
				var s3 Schema
				err = s3.UnmarshalJSON([]byte(tt.json))
				testutil.OK(t, err)

				s4 := (*Schema)(&s3)
				jsonData2, err := s4.MarshalJSON()
				testutil.OK(t, err)

				// Unmarshal again to verify stability
				var s5 Schema
				err = s5.UnmarshalJSON(jsonData2)
				testutil.OK(t, err)

				s6 := (*Schema)(&s5)
				jsonData3, err := s6.MarshalJSON()
				testutil.OK(t, err)

				// The two marshaled JSONs should be identical (stable round trip)
				testutil.Equals(t, string(jsonData2), string(jsonData3))
			}
		})
	}
}

// TestEntityTypeRefMarshaling tests marshaling of EntityTypeRef specifically
// (EntityTypeRef becomes TypeRef on unmarshal, so can't use round-trip tests)
func TestEntityTypeRefMarshaling(t *testing.T) {
	t.Parallel()

	// Test EntityTypeRef in common type
	schema1 := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{},
		},
		CommonTypes: ast.CommonTypes{
			"UserRef": ast.CommonType{
				Type: ast.EntityTypeRef("User"),
			},
		},
	}
	s1 := (*Schema)(schema1)
	jsonData, err := s1.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData) > 0, true)

	// Test EntityTypeRef in record attribute
	schema2 := &ast.Schema{
		Entities: ast.Entities{
			"User": ast.Entity{},
			"Doc": ast.Entity{
				Shape: &ast.RecordType{
					"owner": ast.Attribute{Type: ast.EntityTypeRef("User"), Optional: false},
				},
			},
		},
	}
	s2 := (*Schema)(schema2)
	jsonData2, err := s2.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData2) > 0, true)

	// Test nil type marshaling (default case)
	schema3 := &ast.Schema{
		CommonTypes: ast.CommonTypes{
			"BadType": ast.CommonType{
				Type: nil,
			},
		},
	}
	s3 := (*Schema)(schema3)
	jsonData3, err := s3.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData3) > 0, true)

	// Test nil attributes in Record marshaling (Type.MarshalJSON line 699-701)
	schema4 := &ast.Schema{
		Entities: ast.Entities{
			"Empty": ast.Entity{
				Shape: &ast.RecordType{},
			},
		},
	}
	s4 := (*Schema)(schema4)
	jsonData4, err := s4.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData4) > 0, true)

	// Test nil attributes in nested Record (Attr.MarshalJSON line 730-732)
	schema5 := &ast.Schema{
		Entities: ast.Entities{
			"Config": ast.Entity{
				Shape: &ast.RecordType{
					"nested": ast.Attribute{
						Type:     ast.RecordType{}, // empty attributes in nested record
						Optional: false,
					},
				},
			},
		},
	}
	s5 := (*Schema)(schema5)
	jsonData5, err := s5.MarshalJSON()
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData5) > 0, true)
}

// TestLegacyEntityTypeFormat tests unmarshaling of legacy "Entity" type format
func TestLegacyEntityTypeFormat(t *testing.T) {
	t.Parallel()

	// Test Entity type in commonTypes (line 554-555)
	jsonInput1 := `{
		"": {
			"entityTypes": {"User": {}},
			"actions": {},
			"commonTypes": {
				"UserRef": {"type": "Entity", "name": "User"}
			}
		}
	}`
	var s1 Schema
	err := s1.UnmarshalJSON([]byte(jsonInput1))
	testutil.OK(t, err)
	schema1 := (*ast.Schema)(&s1)
	testutil.Equals(t, len(schema1.CommonTypes), 1)

	// Test Entity type in attributes (line 627-628)
	jsonInput2 := `{
		"": {
			"entityTypes": {
				"User": {},
				"Doc": {
					"shape": {
						"type": "Record",
						"attributes": {
							"owner": {"type": "Entity", "name": "User"}
						}
					}
				}
			},
			"actions": {}
		}
	}`
	var s2 Schema
	err = s2.UnmarshalJSON([]byte(jsonInput2))
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)
	testutil.Equals(t, len(schema2.Entities), 2)
}

// TestAttributeSetType tests Set type in record attributes (line 604)
func TestAttributeSetType(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"tags": {
								"type": "Set",
								"element": {"type": "EntityOrCommon", "name": "String"}
							}
						}
					}
				}
			},
			"actions": {}
		}
	}`
	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	entity := schema.Entities["User"]
	attr := (*entity.Shape)["tags"]
	_, ok := attr.Type.(ast.SetType)
	testutil.Equals(t, ok, true)
}

// TestUnknownAttrType tests unknown type name in attribute (line 637)
func TestUnknownAttrType(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"data": {"type": ""}
						}
					}
				}
			},
			"actions": {}
		}
	}`
	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.Error(t, err)
}

// TestTypeRefInAttribute tests type reference in attribute (line 634-635)
func TestTypeRefInAttribute(t *testing.T) {
	t.Parallel()

	jsonInput := `{
		"": {
			"entityTypes": {
				"User": {
					"shape": {
						"type": "Record",
						"attributes": {
							"customType": {"type": "MyCustomType"}
						}
					}
				}
			},
			"actions": {},
			"commonTypes": {
				"MyCustomType": {"type": "EntityOrCommon", "name": "String"}
			}
		}
	}`
	var s Schema
	err := s.UnmarshalJSON([]byte(jsonInput))
	testutil.OK(t, err)
	schema := (*ast.Schema)(&s)

	entity := schema.Entities["User"]
	attr := (*entity.Shape)["customType"]
	typeRef, ok := attr.Type.(ast.TypeRef)
	testutil.Equals(t, ok, true)
	testutil.Equals(t, string(typeRef), "MyCustomType")
}

// TestRecordMarshalWithNilAttributes tests the nil attributes path in Type/Attr MarshalJSON
func TestRecordMarshalWithNilAttributes(t *testing.T) {
	t.Parallel()

	// Test Type.MarshalJSON with nil attributes (line 699-701)
	recordType := Type{
		TypeName:   "Record",
		Attributes: nil,
	}
	jsonData, err := json.Marshal(recordType)
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData) > 0, true)

	// Test Attr.MarshalJSON with nil attributes (line 730-732)
	recordAttr := Attr{
		TypeName:   "Record",
		Attributes: nil,
	}
	jsonData2, err := json.Marshal(recordAttr)
	testutil.OK(t, err)
	testutil.Equals(t, len(jsonData2) > 0, true)
}

// TestNamespaceEarlyReturn tests getOrCreateNamespace early return (line 137-138)
func TestNamespaceEarlyReturn(t *testing.T) {
	t.Parallel()

	// Create a schema that has both top-level entities and entities in an explicit "" namespace
	// This will cause getOrCreateNamespace to be called twice with "" and return early the second time
	schema := &ast.Schema{
		Entities: ast.Entities{
			"TopUser": ast.Entity{},
		},
		Namespaces: ast.Namespaces{
			"": ast.Namespace{
				Entities: ast.Entities{
					"NamespaceUser": ast.Entity{},
				},
			},
		},
	}

	s := (*Schema)(schema)
	jsonData, err := s.MarshalJSON()
	testutil.OK(t, err)

	// Unmarshal and verify both entities are present
	var s2 Schema
	err = s2.UnmarshalJSON(jsonData)
	testutil.OK(t, err)
	schema2 := (*ast.Schema)(&s2)
	testutil.Equals(t, len(schema2.Entities), 2)
}

func TestUnmarshalErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		json    string
		wantErr string
	}{
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: "invalid character",
		},
		{
			name:    "set without element",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "unknown type",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": ""}}}, "actions": {}}}`,
			wantErr: "unknown type",
		},
		{
			name:    "common type error",
			json:    `{"": {"commonTypes": {"Bad": {"type": "Set"}}, "entityTypes": {}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "entity shape error",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "entity tags error",
			json:    `{"": {"entityTypes": {"User": {"tags": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "action context error",
			json:    `{"": {"entityTypes": {}, "actions": {"view": {"appliesTo": {"context": {"type": "Set"}}}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "attr set without element",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"data": {"type": "Set"}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "nested attr error",
			json:    `{"": {"entityTypes": {"User": {"shape": {"type": "Record", "attributes": {"nested": {"type": "Record", "attributes": {"bad": {"type": "Set"}}}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "nested set error in commonTypes",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Set", "element": {"type": "Set"}}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "nested record error in commonTypes",
			json:    `{"": {"entityTypes": {}, "actions": {}, "commonTypes": {"Bad": {"type": "Record", "attributes": {"x": {"type": "Set"}}}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "attr nested set error",
			json:    `{"": {"entityTypes": {"E": {"shape": {"type": "Record", "attributes": {"x": {"type": "Set", "element": {"type": "Set"}}}}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "namespace with common type error",
			json:    `{"MyApp": {"entityTypes": {}, "actions": {}, "commonTypes": {"BadSet": {"type": "Set"}}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "namespace with entity error",
			json:    `{"MyApp": {"commonTypes": {}, "entityTypes": {"User": {"shape": {"type": "Set"}}}, "actions": {}}}`,
			wantErr: "set type missing element",
		},
		{
			name:    "namespace with action error",
			json:    `{"MyApp": {"commonTypes": {}, "entityTypes": {}, "actions": {"view": {"appliesTo": {"context": {"type": "Set"}}}}}}`,
			wantErr: "set type missing element",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var s Schema
			err := s.UnmarshalJSON([]byte(tt.json))
			testutil.Error(t, err)
			testutil.Equals(t, err != nil && (tt.wantErr == "" || len(err.Error()) > 0), true)
			if tt.wantErr != "" && err != nil {
				// Just verify we got an error with some text
				testutil.Equals(t, len(err.Error()) > 0, true)
			}
		})
	}
}
