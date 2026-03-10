// Policy referencing unknown entity type
permit(
    principal,
    action == Action::"view",
    resource
)
when {
    principal == Unknown::"x"
};
