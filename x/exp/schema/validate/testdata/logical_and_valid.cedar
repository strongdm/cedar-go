// Test: true && false → typeBool (should validate)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && false
};
