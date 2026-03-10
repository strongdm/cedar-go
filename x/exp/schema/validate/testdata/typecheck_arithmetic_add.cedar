// Valid: 1 + 2
permit (
    principal,
    action,
    resource
) when {
    1 + 2 == 3
};

// Invalid: 1 + "two"
permit (
    principal,
    action,
    resource
) when {
    1 + "two" == 0
};
