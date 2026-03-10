// Test: !(principal has age) → typeBool (valid)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    !(principal has age)
};
