permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal.hasTag("key") && principal.getTag("key") == "val" };
