permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { {"a": Bogus::"x"} == {"a": User::"y"} };
