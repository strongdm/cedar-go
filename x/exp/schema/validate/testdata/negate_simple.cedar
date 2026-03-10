// Test negate operator with String (should be type error)
// Just -"hello" by itself
permit (
    principal,
    action,
    resource
)
when {
    -"hello"
};
