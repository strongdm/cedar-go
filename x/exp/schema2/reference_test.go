package schema2_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/cedar-policy/cedar-go/internal/testutil"
	"github.com/cedar-policy/cedar-go/x/exp/schema2"
)

// cedarCLI is the path to the cedar-policy CLI for verification.
// Set CEDAR_CLI env var to override, or it will look in common locations.
func cedarCLI() string {
	if cli := os.Getenv("CEDAR_CLI"); cli != "" {
		return cli
	}
	// Check common locations
	paths := []string{
		"/home/user/cedar-rust/target/release/cedar",
		"/usr/local/bin/cedar",
		"cedar",
	}
	for _, p := range paths {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}
	return ""
}

// verifyWithCedarCLI checks that a schema parses correctly using the reference implementation.
func verifyWithCedarCLI(t *testing.T, schemaContent string) {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available, skipping verification")
	}

	// Write schema to temp file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.cedarschema")
	if err := os.WriteFile(schemaFile, []byte(schemaContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	// Run cedar check-parse
	cmd := exec.Command(cli, "check-parse", "--schema", schemaFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cedar CLI rejected schema:\n%s\nOutput: %s", schemaContent, string(output))
	}
}

// verifyJSONWithCedarCLI checks that a JSON schema parses correctly using the reference implementation.
func verifyJSONWithCedarCLI(t *testing.T, jsonContent string) {
	t.Helper()
	cli := cedarCLI()
	if cli == "" {
		t.Skip("cedar CLI not available, skipping verification")
	}

	// Write schema to temp file
	tmpDir := t.TempDir()
	schemaFile := filepath.Join(tmpDir, "schema.cedarschema.json")
	if err := os.WriteFile(schemaFile, []byte(jsonContent), 0o644); err != nil {
		t.Fatalf("failed to write temp schema: %v", err)
	}

	// Run cedar check-parse --schema-format json
	cmd := exec.Command(cli, "check-parse", "--schema", schemaFile, "--schema-format", "json")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("cedar CLI rejected JSON schema:\n%s\nOutput: %s", jsonContent, string(output))
	}
}

// TestReferenceSchemas tests our parser against schemas from the Rust reference implementation.
func TestReferenceSchemas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
	}{
		// From cedar-policy-cli/sample-data/sandbox_a/schema.cedarschema
		{
			name: "sandbox_a - photo sharing",
			schema: `entity Video in [Account, Album];
entity User in [UserGroup];
entity UserGroup;
entity Administrator;
entity Photo in [Account, Album];
entity Album in [Account];
entity Account;

action listPhotos
  appliesTo { principal: [User], resource: [Album, Photo, Video] };
action view, delete, edit
  appliesTo { principal: [User], resource: [Photo, Video, Album] };
`,
		},

		// From cedar-policy-cli/sample-data/sandbox_b/schema.cedarschema
		{
			name: "sandbox_b - photo sharing with attributes",
			schema: `entity Photo in [Account, Album] {
  account: Account,
  admins: Set<User>,
  private: Bool
};
entity User in [UserGroup] { department: String, jobLevel: Long };
entity AccountGroup;
entity Administrator;
entity UserGroup;
entity Album in [Account] { account: Account, private: Bool };
entity Account in [AccountGroup] { owner?: User };

action view, delete, edit
  appliesTo {
    principal: [User],
    resource: [Photo, Album],
    context: { source_ip: __cedar::ipaddr }
  };
action listPhotos
  appliesTo {
    principal: [User],
    resource: [Album, Photo],
    context: { source_ip: __cedar::ipaddr }
  };
`,
		},

		// From cedar-policy-symcc/tests/data/cedar-examples/github_example/policies.cedarschema
		{
			name: "github example",
			schema: `entity Team, UserGroup in [UserGroup];
entity Issue  = {
  "repo": Repository,
  "reporter": User,
};
entity Org  = {
  "members": UserGroup,
  "owners": UserGroup,
};
entity Repository  = {
  "admins": UserGroup,
  "maintainers": UserGroup,
  "readers": UserGroup,
  "triagers": UserGroup,
  "writers": UserGroup,
};
entity User in [UserGroup, Team];

action push, pull, fork appliesTo {
  principal: [User],
  resource: [Repository]
};
action assign_issue, delete_issue, edit_issue appliesTo {
  principal: [User],
  resource: [Issue]
};
action add_reader, add_writer, add_maintainer, add_admin, add_triager appliesTo {
  principal: [User],
  resource: [Repository]
};
`,
		},

		// From cedar-policy-symcc/tests/data/cedar-examples/tinytodo/policies.cedarschema
		{
			name: "tinytodo",
			schema: `type Task = {
    "id": Long,
    "name": String,
    "state": String,
};

type Tasks = Set<Task>;
entity List in [Application] = {
  "editors": Team,
  "name": String,
  "owner": User,
  "readers": Team,
  "tasks": Tasks,
};
entity Application;
entity User in [Team, Application] = {
  "joblevel": Long,
  "location": String,
};
entity Team in [Team, Application];
action DeleteList, GetList, UpdateList appliesTo {
  principal: [User],
  resource: [List]
};
action CreateList, GetLists appliesTo {
  principal: [User],
  resource: [Application]
};
action CreateTask, UpdateTask, DeleteTask appliesTo {
  principal: [User],
  resource: [List]
};
action EditShare appliesTo {
  principal: [User],
  resource: [List]
};
`,
		},

		// From cedar-policy-symcc/tests/data/cedar-examples/document_cloud/policies.cedarschema
		{
			name: "document_cloud",
			schema: `entity DocumentShare, Drive;
entity Document  = {
  "isPrivate": Bool,
  "manageACL": DocumentShare,
  "modifyACL": DocumentShare,
  "owner": User,
  "publicAccess": String,
  "viewACL": DocumentShare,
};
entity Group in [DocumentShare] = {
  "owner": User,
};
entity Public in [DocumentShare];
entity User in [Group] = {
  "blocked": Set<User>,
  "personalGroup": Group,
};

action DeleteGroup, ModifyGroup appliesTo {
  principal: [User],
  resource: [Group],
  context: {
    "is_authenticated": Bool,
  }
};
action CreateGroup appliesTo {
  principal: [User],
  resource: [Drive],
  context: {
    "is_authenticated": Bool,
  }
};
action ViewDocument appliesTo {
  principal: [User, Public],
  resource: [Document],
  context: {
    "is_authenticated": Bool,
  }
};
action AddToShareACL, DeleteDocument, EditIsPrivate, EditPublicAccess appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "is_authenticated": Bool,
  }
};
action ModifyDocument appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "is_authenticated": Bool,
  }
};
action CreateDocument appliesTo {
  principal: [User],
  resource: [Drive],
  context: {
    "is_authenticated": Bool,
  }
};
`,
		},

		// From cedar-policy-core/src/validator/cedar_schema/testfiles/example.cedarschema
		{
			name: "nested types and namespaces",
			schema: `entity TopLevel = {
  obj: {
    String: String,
    entity: String,
    namespace: String,
    "nested-With-Dash": String,
    nestedStr: String,
    type: String
  }
};

namespace EmptyNs {
}

namespace Ns {
  type Bar = {
    obj: {
      nestedLong: Long,
      nestedObj: {
        nestedStr: String
      }
    },
    setWithAnonymousType: Set<{
      key: String,
      val: String
    }>
  };

  entity Resource = {
    bar: Bar
  };

  entity User;

  action "get" appliesTo {
    principal: [User],
    resource: [Resource],
    context: {}
  };
}
`,
		},

		// From cedar-policy-symcc/tests/data/cedar-examples/tags_n_roles/policies.cedarschema
		{
			name: "tags and roles with action groups",
			schema: `entity Role;
entity User in [Role] {
  allowedTagsForRole: {
    "Role-A"?: {
        production_status?: Set<String>,
        country?: Set<String>,
        stage?: Set<String>,
    },
    "Role-B"?: {
        production_status?: Set<String>,
        country?: Set<String>,
        stage?: Set<String>,
    },
  },
};
entity Workspace {
  tags: {
    production_status?: Set<String>,
    country?: Set<String>,
    stage?: Set<String>,
  }
};

action "Role-A Actions";
action "Role-B Actions";
action UpdateWorkspace in ["Role-A Actions"] appliesTo {
  principal: User,
  resource: Workspace,
};

action DeleteWorkspace in ["Role-A Actions"] appliesTo {
  principal: User,
  resource: Workspace,
};

action ReadWorkspace in ["Role-A Actions", "Role-B Actions"] appliesTo {
  principal: User,
  resource: Workspace,
};
`,
		},

		// From cedar-policy-symcc/tests/data/cedar-examples/hotel_chains/policies.cedarschema
		{
			name: "hotel chains",
			schema: `type PermissionsMap = {
  hotelReservations: Set<Hotel>,
  propertyReservations: Set<Property>,
};
entity User {
  viewPermissions: PermissionsMap,
  memberPermissions: PermissionsMap,
  hotelAdminPermissions: Set<Hotel>,
  propertyAdminPermissions: Set<Property>,
};
entity Property in [Hotel];
entity Hotel in [Hotel];
entity Reservation in [Property];

action viewReservation, updateReservation, grantAccessReservation
  appliesTo {
    principal: User,
    resource: Reservation,
  };

action createReservation, viewProperty, updateProperty, grantAccessProperty
  appliesTo {
    principal: User,
    resource: Property,
  };

action createProperty, createHotel, viewHotel, updateHotel, grantAccessHotel
  appliesTo {
    principal: User,
    resource: Hotel,
  };
`,
		},

		// From cedar-policy-cli/sample-data/tpe_rfc/schema.cedarschema
		{
			name: "tpe_rfc - document access",
			schema: `entity User;

entity Document  = {
  "isPublic": Bool,
  "owner": User
};

action View appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "hasMFA": Bool,
  }
};

action Delete appliesTo {
  principal: [User],
  resource: [Document],
  context: {
    "hasMFA": Bool,
    "srcIP": ipaddr
  }
};
`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Parse with our implementation
			schema, err := schema2.UnmarshalCedar([]byte(tt.schema))
			testutil.OK(t, err)
			testutil.Equals(t, schema != nil, true)

			// Round-trip test
			marshaled := schema.MarshalCedar()
			schema2Parsed, err := schema2.UnmarshalCedar(marshaled)
			testutil.OK(t, err)
			testutil.Equals(t, len(schema2Parsed.Nodes), len(schema.Nodes))

			// Verify our output with the reference implementation
			verifyWithCedarCLI(t, string(marshaled))
		})
	}
}

// TestReferenceVerification verifies that specific edge cases match reference implementation behavior.
func TestReferenceVerification(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
	}{
		{
			name:   "empty namespace",
			schema: `namespace EmptyNamespace {}`,
		},
		{
			name:   "multiple entities same line",
			schema: `entity User, Admin, Guest;`,
		},
		{
			name:   "multiple actions same line",
			schema: `action read, write, delete;`,
		},
		{
			name:   "entity with self-reference",
			schema: `entity Group in [Group];`,
		},
		{
			name:   "action quoted name",
			schema: `action "view document";`,
		},
		{
			name:   "action in action group",
			schema: `action "readActions"; action view in ["readActions"];`,
		},
		{
			name:   "record with reserved words as keys",
			schema: `type Config = { entity: String, namespace: String, type: String, action: String };`,
		},
		{
			name:   "deeply nested anonymous types",
			schema: `type Deep = { level1: { level2: { level3: { value: String } } } };`,
		},
		{
			name:   "set of anonymous record",
			schema: `type Items = Set<{ key: String, value: Long }>;`,
		},
		{
			name:   "empty context",
			schema: `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc, context: {} };`,
		},
		{
			name:   "single principal without brackets",
			schema: `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc };`,
		},
		{
			name:   "extension type ipaddr short form",
			schema: `type IP = ipaddr;`,
		},
		{
			name:   "extension type decimal short form",
			schema: `type Price = decimal;`,
		},
		{
			name:   "extension type datetime short form",
			schema: `type Time = datetime;`,
		},
		{
			name:   "extension type duration short form",
			schema: `type Dur = duration;`,
		},
		{
			name:   "entity enum",
			schema: `entity Status enum ["active", "pending", "closed"];`,
		},
		{
			name:   "annotation with value",
			schema: `@doc("User entity description") entity User;`,
		},
		{
			name:   "annotation without value",
			schema: `@deprecated entity LegacyUser;`,
		},
		{
			name:   "multiple annotations",
			schema: `@doc("Deprecated user") @deprecated @internal entity OldUser;`,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// First verify with reference implementation
			verifyWithCedarCLI(t, tt.schema)

			// Parse with our implementation
			schema, err := schema2.UnmarshalCedar([]byte(tt.schema))
			testutil.OK(t, err)
			testutil.Equals(t, schema != nil, true)

			// Round-trip
			marshaled := schema.MarshalCedar()
			_, err = schema2.UnmarshalCedar(marshaled)
			testutil.OK(t, err)
		})
	}
}

// TestAccessControlSchema tests a comprehensive access control schema similar to real-world usage.
// This is an anonymized version based on common IAM patterns.
func TestAccessControlSchema(t *testing.T) {
	t.Parallel()

	schema := `// Access Control Schema
// Comprehensive example covering common IAM patterns

namespace AccessControl {
	// Common types for reuse
	type Metadata = {
		"createdAt": __cedar::datetime,
		"updatedAt"?: __cedar::datetime,
		"createdBy": User,
	};

	type ConnectionConfig = {
		"host": String,
		"port": Long,
		"protocol": String,
		"timeout"?: __cedar::duration,
	};

	type AuditInfo = {
		"ip": __cedar::ipaddr,
		"userAgent"?: String,
		"sessionId": String,
		"timestamp": __cedar::datetime,
	};

	// Principal hierarchy
	entity User in [Group, Role] = {
		"email": String,
		"displayName": String,
		"department"?: String,
		"manager"?: User,
		"active": Bool,
		"mfaEnabled": Bool,
		"metadata": Metadata,
	};

	entity Group in [Group] = {
		"name": String,
		"description"?: String,
		"metadata": Metadata,
	};

	entity Role in [Role] = {
		"name": String,
		"permissions": Set<String>,
		"metadata": Metadata,
	};

	entity ServiceAccount in [Role] = {
		"name": String,
		"owner": User,
		"expiresAt"?: __cedar::datetime,
		"metadata": Metadata,
	};

	// Resource hierarchy
	entity Organization = {
		"name": String,
		"tier": String,
		"metadata": Metadata,
	};

	entity Project in [Organization] = {
		"name": String,
		"owner": User,
		"members": Set<User>,
		"metadata": Metadata,
	};

	entity Environment in [Project] = {
		"name": String,
		"type": EnvironmentType,
		"locked": Bool,
		"metadata": Metadata,
	};

	entity EnvironmentType enum ["development", "staging", "production"];

	entity Resource in [Environment] = {
		"name": String,
		"resourceType": ResourceType,
		"config": ConnectionConfig,
		"tags": Set<String>,
		"metadata": Metadata,
	};

	entity ResourceType enum ["database", "server", "cluster", "gateway", "secret"];

	// Session context
	type SessionContext = {
		"audit": AuditInfo,
		"mfaVerified": Bool,
		"riskScore"?: Long,
	};

	// Actions organized by resource type
	action "manage-organization";
	action "view-organization" in ["manage-organization"] appliesTo {
		principal: [User, ServiceAccount],
		resource: Organization,
		context: SessionContext,
	};
	action "edit-organization" in ["manage-organization"] appliesTo {
		principal: [User],
		resource: Organization,
		context: SessionContext,
	};

	action "manage-project";
	action "view-project" in ["manage-project"] appliesTo {
		principal: [User, ServiceAccount],
		resource: Project,
		context: SessionContext,
	};
	action "create-project" in ["manage-project"] appliesTo {
		principal: [User],
		resource: Organization,
		context: SessionContext,
	};
	action "delete-project" in ["manage-project"] appliesTo {
		principal: [User],
		resource: Project,
		context: SessionContext,
	};

	action "manage-resource";
	action "connect" in ["manage-resource"] appliesTo {
		principal: [User, ServiceAccount],
		resource: Resource,
		context: SessionContext,
	};
	action "view-resource" in ["manage-resource"] appliesTo {
		principal: [User, ServiceAccount],
		resource: Resource,
		context: SessionContext,
	};
	action "create-resource" in ["manage-resource"] appliesTo {
		principal: [User],
		resource: Environment,
		context: SessionContext,
	};
	action "delete-resource" in ["manage-resource"] appliesTo {
		principal: [User],
		resource: Resource,
		context: SessionContext,
	};
	action "modify-resource" in ["manage-resource"] appliesTo {
		principal: [User],
		resource: Resource,
		context: SessionContext,
	};

	// Admin actions
	action "admin";
	action "manage-users" in ["admin"] appliesTo {
		principal: [User],
		resource: Organization,
		context: SessionContext,
	};
	action "manage-roles" in ["admin"] appliesTo {
		principal: [User],
		resource: Organization,
		context: SessionContext,
	};
	action "view-audit-logs" in ["admin"] appliesTo {
		principal: [User],
		resource: [Organization, Project],
		context: SessionContext,
	};
}
`

	// Parse with our implementation
	parsed, err := schema2.UnmarshalCedar([]byte(schema))
	testutil.OK(t, err)
	testutil.Equals(t, len(parsed.Nodes), 1) // One namespace

	// Round-trip
	marshaled := parsed.MarshalCedar()
	reparsed, err := schema2.UnmarshalCedar(marshaled)
	testutil.OK(t, err)
	testutil.Equals(t, len(reparsed.Nodes), len(parsed.Nodes))

	// Verify with reference implementation
	verifyWithCedarCLI(t, string(marshaled))
}

// TestMultiNamespaceSchema tests schemas with multiple namespaces.
func TestMultiNamespaceSchema(t *testing.T) {
	t.Parallel()

	schema := `// Multi-namespace schema demonstrating cross-namespace references

namespace Core {
	type Timestamp = __cedar::datetime;
	type IPAddress = __cedar::ipaddr;

	entity BaseUser = {
		"id": String,
		"createdAt": Timestamp,
	};
}

namespace Auth {
	entity User in [Core::BaseUser] = {
		"email": String,
		"role": String,
	};

	entity Session = {
		"user": User,
		"expiresAt": Core::Timestamp,
		"ip": Core::IPAddress,
	};

	action login appliesTo {
		principal: User,
		resource: Session,
	};
}

namespace Resources {
	entity Document = {
		"owner": Auth::User,
		"title": String,
	};

	action view, edit, delete appliesTo {
		principal: Auth::User,
		resource: Document,
	};
}
`

	parsed, err := schema2.UnmarshalCedar([]byte(schema))
	testutil.OK(t, err)
	testutil.Equals(t, len(parsed.Nodes), 3) // Three namespaces

	// Round-trip
	marshaled := parsed.MarshalCedar()
	reparsed, err := schema2.UnmarshalCedar(marshaled)
	testutil.OK(t, err)
	testutil.Equals(t, len(reparsed.Nodes), 3)

	// Verify with reference
	verifyWithCedarCLI(t, string(marshaled))
}

// TestEdgeCasesFromReference tests edge cases discovered in reference implementation tests.
func TestEdgeCasesFromReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		schema string
		valid  bool
	}{
		{
			name:   "entity with equals and shape",
			schema: `entity User = { name: String };`,
			valid:  true,
		},
		{
			name:   "entity with in and equals",
			schema: `entity Group; entity User in [Group] = { name: String };`,
			valid:  true,
		},
		{
			name:   "optional in nested record",
			schema: `entity User { config: { setting?: Bool } };`,
			valid:  true,
		},
		{
			name:   "multiple optional attributes",
			schema: `entity User { a?: String, b?: Long, c?: Bool };`,
			valid:  true,
		},
		{
			name:   "set of set",
			schema: `type Matrix = Set<Set<Long>>;`,
			valid:  true,
		},
		{
			name:   "attribute key with dash",
			schema: `type Config = { "my-key": String };`,
			valid:  true,
		},
		{
			name:   "attribute key with spaces",
			schema: `type Config = { "my key here": String };`,
			valid:  true,
		},
		{
			name:   "action name with spaces",
			schema: `action "do something";`,
			valid:  true,
		},
		{
			name:   "unicode in string attribute key",
			schema: `type Config = { "キー": String };`,
			valid:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.schema))
			if tt.valid {
				testutil.OK(t, err)
				testutil.Equals(t, schema != nil, true)

				// Verify with reference
				verifyWithCedarCLI(t, tt.schema)
			} else {
				testutil.Error(t, err)
			}
		})
	}
}

// TestMarshalMatchesReference ensures our marshal output is valid according to reference.
func TestMarshalMatchesReference(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		schema   string
		contains []string // Strings that must be present in marshaled output
	}{
		{
			name:   "entity with optional",
			schema: `entity User { name?: String };`,
			contains: []string{
				"entity User",
				"name?",
				"String",
			},
		},
		{
			name:   "action with context",
			schema: `entity User; entity Doc; action view appliesTo { principal: User, resource: Doc, context: { ip: __cedar::ipaddr } };`,
			contains: []string{
				"action view",
				"principal",
				"resource",
				"context",
				"ipaddr",
			},
		},
		{
			name:   "namespace with declarations",
			schema: `namespace App { entity User; action view; }`,
			contains: []string{
				"namespace App",
				"entity User",
				"action view",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			schema, err := schema2.UnmarshalCedar([]byte(tt.schema))
			testutil.OK(t, err)

			marshaled := string(schema.MarshalCedar())

			for _, substr := range tt.contains {
				if !strings.Contains(marshaled, substr) {
					t.Errorf("marshaled output missing %q:\n%s", substr, marshaled)
				}
			}

			// Verify marshaled output with reference
			verifyWithCedarCLI(t, marshaled)
		})
	}
}
