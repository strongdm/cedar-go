// Valid arithmetic: 1 + 2
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + 2 == 3
};

// Invalid arithmetic: 1 + "two"
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + "two" == 3
};
