// Test: false && Action::"view" should validate (short-circuit)
// The right side is an entity (not bool), but because left is false,
// it should short-circuit and not type-check the right side.
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    false && Action::"view"
};
