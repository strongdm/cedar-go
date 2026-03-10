permit(
    principal == User::"alice",
    action == Action::"view",
    resource
) when {
    principal == resource
};
