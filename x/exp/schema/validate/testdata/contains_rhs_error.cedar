permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { [1, 2, 3].contains(1 + "x") };
