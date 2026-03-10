// if principal has age then false else principal.age > 0
// The else branch should NOT get test capabilities from the condition,
// so accessing principal.age (which is optional) should fail.
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal has age then false else principal.age > 0
};
