permit(
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    (if true then 1 else 2) == 1
};
