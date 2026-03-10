// Test negate with an argument that itself has a type error
// This should produce a compound error: both the argument error and the negate type error
permit (
    principal,
    action,
    resource
) when {
    -(principal + 1)
};
