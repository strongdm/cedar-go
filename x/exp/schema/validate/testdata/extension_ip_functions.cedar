// Test ip() extension function returns ipaddr type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1") == context.ip
};

// Test isLoopback() method on ipaddr returns bool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1").isLoopback()
};
