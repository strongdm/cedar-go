// Test: true && 42 → error (right operand is not bool)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && 42
};
