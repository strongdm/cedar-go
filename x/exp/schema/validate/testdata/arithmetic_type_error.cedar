// Invalid arithmetic: 1 + "two" (type mismatch)
permit (
    principal,
    action,
    resource
) when {
    1 + "two" == 3
};
