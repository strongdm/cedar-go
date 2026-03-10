// Access a nonexistent attribute on principal - should fail validation
permit (
    principal,
    action == Action::"view",
    resource
) when {
    principal.nonexistent == "test"
};
