// Test entity in unrelated type hierarchy
// User in Photo::"b" should validate (permissively) even though they're unrelated
permit (
    principal,
    action == Action::"view",
    resource
) when {
    User::"a" in Photo::"b"
};
