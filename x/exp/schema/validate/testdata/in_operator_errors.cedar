// Test: 42 in Group::"admins" → error (left must be entity)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42 in Group::"admins"
};

// Test: principal in "admins" → error (right must be entity or set of entities)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in "admins"
};

// Test: principal in [42] → error (set of non-entities)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal in [42]
};
