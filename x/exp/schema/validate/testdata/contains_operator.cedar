// Valid: set.contains() is allowed
permit (
    principal,
    action,
    resource
) when {
    [1, 2].contains(1)
};

// Invalid: contains() on a non-set (should fail validation)
permit (
    principal,
    action,
    resource
) when {
    42.contains(1)
};
