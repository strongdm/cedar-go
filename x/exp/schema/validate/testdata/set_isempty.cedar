// Test: [1].isEmpty() → typeBool (should validate)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].isEmpty()
};
