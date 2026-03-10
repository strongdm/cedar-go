// Valid: adding two longs
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + 2 > 0
};

// Invalid: adding long and string
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + "two" > 0
};
