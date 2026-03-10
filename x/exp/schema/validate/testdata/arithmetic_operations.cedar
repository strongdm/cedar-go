// Valid arithmetic: 1 + 2
permit (
    principal,
    action,
    resource
) when {
    1 + 2 == 3
};

// Invalid arithmetic: 1 + "two" should fail type checking
forbid (
    principal,
    action,
    resource
) when {
    1 + "two" == 0
};
