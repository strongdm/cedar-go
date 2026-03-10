// Test negation of a long value (should pass)
permit (
    principal,
    action,
    resource
) when {
    -42 > 0
};
