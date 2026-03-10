// Test: true && 42 → error (42 is not Bool)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && 42
};
