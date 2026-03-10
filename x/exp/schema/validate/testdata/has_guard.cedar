// Tests that capability propagation works with has guard
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    principal has age && principal.age > 18
};
