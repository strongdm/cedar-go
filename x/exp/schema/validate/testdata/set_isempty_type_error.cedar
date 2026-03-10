// Test: 42.isEmpty() → error (not a set)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42.isEmpty()
};
