// Valid: string like pattern
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.title like "hello*"
};

// Invalid: integer like pattern (should fail type check)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42 like "h*"
};
