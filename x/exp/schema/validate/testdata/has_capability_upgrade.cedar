// has guard followed by access - valid
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal has name && principal has name && principal.name == "admin" };
