// Tests that variables (principal, action, resource, context) have correct types
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    principal.name == "alice" &&
    context.ip.isLoopback()
};
