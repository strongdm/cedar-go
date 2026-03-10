// Valid: [1].isEmpty() returns bool
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].isEmpty()
};

// Invalid: 42.isEmpty() is a type error
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42.isEmpty()
};
