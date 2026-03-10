permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (Bogus::"x" in User::"y") && true };
