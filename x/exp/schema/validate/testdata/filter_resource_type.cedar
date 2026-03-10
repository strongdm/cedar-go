permit (
    principal is User,
    action == Action::"viewDoc",
    resource is Document
) when { true };
