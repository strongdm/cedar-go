// Test negate operation type checking
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    -42 == -42
};

// Invalid: negating a string should fail
forbid (
    principal,
    action == Action::"view",
    resource
)
when {
    -"hello" == 0
};
