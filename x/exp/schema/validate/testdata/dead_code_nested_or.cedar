permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ((Bogus::"x" == User::"y") || true) };
