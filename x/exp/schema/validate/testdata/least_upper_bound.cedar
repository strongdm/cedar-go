// Test least upper bound computation in if-then-else
// The if condition evaluates to bool, then/else branches should compute LUB
permit (
    principal,
    action,
    resource
) when {
    if true then true else false
};

// This should fail - branches have incompatible types (Long vs String)
forbid (
    principal,
    action,
    resource
) when {
    if true then 1 else "hello"
};
