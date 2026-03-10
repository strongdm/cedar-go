// Valid: greater than comparison
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 > 2
};

// Valid: greater than or equal comparison
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 >= 2
};
