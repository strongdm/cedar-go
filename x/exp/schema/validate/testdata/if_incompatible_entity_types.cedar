// Test: if-then-else with incompatible entity types in branches should fail
// principal has name ? User::"a" : Photo::"b"
// The LUB of User and Photo cannot be computed (they are unrelated entity types)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    if principal has name
    then User::"a"
    else Photo::"b"
};
