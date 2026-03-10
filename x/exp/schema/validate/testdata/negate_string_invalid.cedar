// Test negation of a string value (should fail validation)
permit (
    principal,
    action,
    resource
) when {
    -"hello" > 0
};
