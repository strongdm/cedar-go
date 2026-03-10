// Test value type checking

// Entity UID of known type
permit (
    principal,
    action,
    resource
) when {
    User::"alice" == User::"alice"
};

// Entity UID of action type
permit (
    principal,
    action,
    resource
) when {
    Action::"view" == Action::"view"
};

// Entity UID of enum type
permit (
    principal,
    action,
    resource
) when {
    Color::"red" == Color::"red"
};

// Entity UID of unknown type → error
permit (
    principal,
    action,
    resource
) when {
    Unknown::"x" == Unknown::"x"
};

// Set value
permit (
    principal,
    action,
    resource
) when {
    [] == []
};

// Record value
permit (
    principal,
    action,
    resource
) when {
    {"x": 1} == {"x": 1}
};

// IPAddr value
permit (
    principal,
    action,
    resource
) when {
    ip("127.0.0.1") == ip("127.0.0.1")
};

// Decimal value
permit (
    principal,
    action,
    resource
) when {
    decimal("1.0") == decimal("1.0")
};
