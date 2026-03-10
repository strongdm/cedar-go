// Test that entity refs are validated even with || short-circuit
// true || Unknown::"x" → error (unknown entity ref)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true || Unknown::"x"
};
