permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal.hasTag("role") && principal.getTag("role") == "admin" };
