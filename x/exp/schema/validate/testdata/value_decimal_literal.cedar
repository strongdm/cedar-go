permit (
    principal,
    action == Action::"view",
    resource
) when {
    decimal("1.23") > decimal("0.5")
};
