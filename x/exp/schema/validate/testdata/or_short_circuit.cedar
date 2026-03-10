// true || <non-bool> → valid (short-circuit, RHS not type checked)
permit (
    principal,
    action == Action::"view",
    resource
) when {
    true || User::"alice"
};
