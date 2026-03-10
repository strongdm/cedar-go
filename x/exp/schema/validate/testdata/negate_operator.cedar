// Test negate operator with valid Long
permit (
    principal,
    action,
    resource
) when {
    -42 == -42
};

// Test negate operator on invalid type (String) - should fail validation
permit (
    principal,
    action,
    resource
) when {
    -"hello"
};
