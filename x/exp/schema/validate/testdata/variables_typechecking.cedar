// Test that principal variable is correctly typed as User
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal == principal
};

// Test that context variable is correctly typed with ip attribute
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.ip.isLoopback()
};

// Test that resource variable is correctly typed as Document
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource == resource
};

// Test that action variable is correctly typed
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    action == Action::"view"
};
