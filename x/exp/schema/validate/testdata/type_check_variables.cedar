// Test basic variable access - principal, context
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    principal.name == "alice" &&
    context.ip == ip("127.0.0.1")
};
