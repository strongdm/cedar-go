// Test like operator with string - should validate
permit (
    principal,
    action,
    resource
) when {
    "hello" like "h*"
};

// Test like operator with non-string - should fail validation
forbid (
    principal,
    action,
    resource
) when {
    42 like "h*"
};
