// Test that `has` guard enables safe access to optional attributes
permit(
    principal,
    action == Action::"view",
    resource
)
when {
    principal has age && principal.age > 18
};
