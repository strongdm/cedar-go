// Test that principal variable access to required attributes is allowed
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.name == "alice"
};
