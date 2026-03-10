// Test extension function ip() returns ipaddr type
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1") == context.ip
};

// Test extension function isLoopback() on ipaddr returns bool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1").isLoopback()
};

// Test chaining extension functions
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.ip.isLoopback()
};
