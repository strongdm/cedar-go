permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { action has name };
