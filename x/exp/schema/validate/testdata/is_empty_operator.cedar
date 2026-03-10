// Valid: isEmpty() on a set
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].isEmpty()
};

// Invalid: isEmpty() on a non-set (Long)
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42.isEmpty()
};
