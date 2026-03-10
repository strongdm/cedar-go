// Test type checking for literal values

// Entity UID value of known type
permit (
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"doc1"
);

// IPAddr extension value
permit (
    principal,
    action == Action::"view",
    resource
) when {
    ip("127.0.0.1").isLoopback()
};

// Datetime extension value
permit (
    principal,
    action == Action::"view",
    resource
) when {
    datetime("2024-01-01") < datetime("2025-01-01")
};

// Record value with has operator
permit (
    principal,
    action == Action::"view",
    resource
) when {
    {"x": 1} has "x"
};

