// Test that when OR branches provide the same capability,
// the intersection preserves it and access should succeed
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    (principal has age || principal has age) && principal.age > 18
};
