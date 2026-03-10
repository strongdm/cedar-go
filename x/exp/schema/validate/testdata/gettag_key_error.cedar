permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal.hasTag("x") && principal.getTag(Bogus::"x") == "z" };
