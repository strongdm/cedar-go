permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    (if context.flag then Admin::"a" else User::"u") is User
};
