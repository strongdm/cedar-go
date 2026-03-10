// Test that comparing entity with string is a type error
permit(
    principal,
    action == Action::"view",
    resource
)
when {
    principal == "foo"
};
