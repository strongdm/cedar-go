// Test that accessing optional attribute without has guard fails
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    principal.age > 18
};
