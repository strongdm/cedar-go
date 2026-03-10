permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1").isLoopback()
};
