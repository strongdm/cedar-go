// Test case: unknown action "delete" should fail
permit (
    principal,
    action == Action::"delete",
    resource
);

// Test case: Group as principal when view requires User should fail
permit (
    principal is Group,
    action == Action::"view",
    resource
);

// Test case: valid simple policy
permit (
    principal is User,
    action == Action::"view",
    resource is Document
);
