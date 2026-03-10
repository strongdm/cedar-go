// Test that Never type is subtype of all types
// This is unreachable code that would fail type checking
permit (
    principal,
    action,
    resource
) when {
    // This will cause type errors demonstrating Never behavior
    (if false then 1 else "x") > 0
};
