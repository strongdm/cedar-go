permit (
    principal,
    action == Action::"view",
    resource
) when {
    [User::"alice", User::"bob"].contains(principal)
};
