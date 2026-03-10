permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if context.isAdmin then Admin::"admin" else User::"user").level > 0
};
