// Test: [1].contains("test") should fail type check (set contains Long, arg is String)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    [1].contains("test")
};
