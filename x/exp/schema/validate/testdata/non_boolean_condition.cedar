// Invalid: when condition must be boolean, not Long
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42
};
