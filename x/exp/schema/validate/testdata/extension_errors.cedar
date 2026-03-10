// Wrong argument count for ip() - expects 1 string, not 2
permit (
    principal,
    action == Action::"view",
    resource
) when {
    ip("1.2.3.4", "5.6.7.8").isLoopback()
};

// Wrong argument type for ip() - expects string, not long
forbid (
    principal,
    action == Action::"view",
    resource
) when {
    ip(42).isLoopback()
};
