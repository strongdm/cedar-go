// All unconstrained scopes - should validate
permit (
    principal,
    action,
    resource
);

// Specific entities in scope - should validate
permit (
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"doc1"
);

// Action in set - should validate
permit (
    principal is User,
    action in [Action::"view", Action::"edit"],
    resource is Document
);

// Action in hierarchy - should validate
permit (
    principal is User,
    action in Action::"view",
    resource is Document
);
