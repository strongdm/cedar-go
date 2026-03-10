// Test that when OR branches provide different capabilities,
// the intersection is empty and access should fail
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    (principal has age || principal has name) && principal.age > 18
};
