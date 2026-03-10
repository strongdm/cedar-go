// Valid: set.containsAll(set)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.tags.containsAll([1, 2])
};

// Valid: set.containsAny(set)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.tags.containsAny([3, 4])
};

// Invalid: non-set.containsAll(set)
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42.containsAll([1, 2])
};

// Invalid: set.containsAll(non-set)
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.tags.containsAll(42)
};
