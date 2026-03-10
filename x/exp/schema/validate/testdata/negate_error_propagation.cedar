// Test error propagation in negate operator
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    -(resource.unknownAttr) == 0
};
