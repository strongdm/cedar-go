// false || <non-bool> → error (no short-circuit, RHS must be bool)
permit (
    principal,
    action == Action::"view",
    resource
) when {
    false || User::"alice"
};
