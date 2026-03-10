permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { {"key": Bogus::"x"}.key == "hello" };
