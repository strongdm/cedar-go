// Policy with valid when condition (equality check)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 == 1
};

// Policy with invalid when condition (type error: Long + String)
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    1 + "x" == 2
};
