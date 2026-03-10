permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { "hello".getTag("x") == "z" };
