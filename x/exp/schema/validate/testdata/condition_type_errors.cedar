// Valid: boolean condition
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true
};

// Invalid: non-boolean condition (Long)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42
};

// Invalid: type error in condition (Long + String)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + "x"
};
