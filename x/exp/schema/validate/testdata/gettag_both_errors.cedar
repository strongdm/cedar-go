permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal.hasTag("key") && principal.getTag(1 + "bad") == "val" };
