// This policy creates a set with two unrelated entity types (User and Photo)
// In strict mode, this should be an error because the LUB of User and Photo is incompatible
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [User::"alice", Photo::"p1"].contains(principal)
};
