// Set with incompatible element types (Long and String) should fail
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    [1, "x"].contains(1)
};
