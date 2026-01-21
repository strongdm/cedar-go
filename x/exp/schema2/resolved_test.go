package schema2

import (
	"testing"
)

// TestResolvedSchemaComprehensive demonstrates the full resolution capabilities:
// - Global types that reference other global types
// - Namespaced types that reference types within their namespace
// - Entities, enums, and actions with complex relationships
// - Verification that resolved output is fully qualified and deterministic
func TestResolvedSchemaComprehensive(t *testing.T) {
	t.Parallel()

	input := `
// Global type definitions that reference each other
type Coordinate = {
	x: Long,
	y: Long,
};

type Location = {
	position: Coordinate,
	name: String,
};

// Global entity that uses global types
entity Place {
	location: Location,
};

// Enum at global level
entity GlobalStatus enum ["active", "inactive", "pending"];

namespace OrgA {
	// Type within namespace
	type Office = {
		building: String,
		floor: Long,
	};

	// Type that references another type in the same namespace
	type Employee = {
		name: String,
		office: Office,
	};

	// Entity that uses namespace-local type
	entity User {
		info: Employee,
	};

	// Enum within namespace
	entity Role enum ["admin", "member", "guest"];

	// Entity hierarchy
	entity Group;
	entity Team in [Group];

	// Actions with appliesTo
	action View appliesTo {
		principal: [User, Team],
		resource: Place,
		context: {
			reason: String,
		},
	};

	action Edit appliesTo {
		principal: User,
		resource: Place,
	};
}

namespace OrgB {
	entity Member;

	// Action that references entities from another namespace and global entities
	action Access appliesTo {
		principal: [Member, OrgA::User],
		resource: Place,
	};
}

// Top-level action
action GlobalView appliesTo {
	principal: [OrgA::User, OrgB::Member],
	resource: Place,
};
`

	// Expected output after resolution - all types are fully resolved and inlined
	// Common types are NOT declared separately - they're inlined where used
	// Namespaces ARE present with unqualified names within them
	// Top-level entities/enums/actions come first, then namespaces (sorted)
	expected := `entity Place = {"location": {
  "position": {
    "x": Long,
    "y": Long,
  },
  "name": String,
}};

entity GlobalStatus enum ["active", "inactive", "pending"];

action GlobalView appliesTo {
  principal: [OrgA::User, OrgB::Member],
  resource: Place,
  context: {},
};

namespace OrgA {
  entity Group;

  entity Team in OrgA::Group;

  entity User = {"info": {
    "name": String,
    "office": {
      "building": String,
      "floor": Long,
    },
  }};

  entity Role enum ["admin", "member", "guest"];

  action Edit appliesTo {
    principal: OrgA::User,
    resource: Place,
    context: {},
  };

  action View appliesTo {
    principal: [OrgA::User, OrgA::Team],
    resource: Place,
    context: {
      "reason": String,
    },
  };
}

namespace OrgB {
  entity Member;

  action Access appliesTo {
    principal: [OrgB::Member, OrgA::User],
    resource: Place,
    context: {},
  };
}
`

	// Parse the input schema
	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() failed: %v", err)
	}

	// Resolve the schema
	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Convert to Schema and marshal to Cedar format
	// Use MarshalCedar directly
	marshaled, err := resolved.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() failed: %v", err)
	}

	// Verify exact output match
	if string(marshaled) != expected {
		t.Errorf("MarshalCedar() output mismatch\n\nGot:\n%s\n\nExpected:\n%s", string(marshaled), expected)
	}
}

// TestResolvedSchemaGlobalEntityReference validates that global entities
// are correctly referenced (without namespace qualification) from within namespaces.
func TestResolvedSchemaGlobalEntityReference(t *testing.T) {
	t.Parallel()

	input := `
entity GlobalEntity;

namespace MyNamespace {
	entity LocalEntity;

	action MyAction appliesTo {
		principal: LocalEntity,
		resource: GlobalEntity,
	};
}
`

	expected := `entity GlobalEntity;

namespace MyNamespace {
  entity LocalEntity;

  action MyAction appliesTo {
    principal: MyNamespace::LocalEntity,
    resource: GlobalEntity,
    context: {},
  };
}
`

	var s Schema
	if err := s.UnmarshalCedar([]byte(input)); err != nil {
		t.Fatalf("UnmarshalCedar() failed: %v", err)
	}

	resolved, err := s.Resolve()
	if err != nil {
		t.Fatalf("Resolve() failed: %v", err)
	}

	// Use MarshalCedar directly
	marshaled, err := resolved.MarshalCedar()
	if err != nil {
		t.Fatalf("MarshalCedar() failed: %v", err)
	}

	// Verify exact output match
	if string(marshaled) != expected {
		t.Errorf("MarshalCedar() output mismatch\n\nGot:\n%s\n\nExpected:\n%s", string(marshaled), expected)
	}
}
