// Valid: ip("127.0.0.1") creates an IP address
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("127.0.0.1").isLoopback()
};

// Valid: isMulticast method on IP address
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    ip("224.0.0.1").isMulticast()
};
