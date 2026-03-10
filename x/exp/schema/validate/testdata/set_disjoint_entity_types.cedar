// Test that a set with disjoint entity types causes a type error
permit(
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"report"
)
when {
    [User::"alice", Photo::"sunset"].contains(principal)
};
