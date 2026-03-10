permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { {"a": (Bogus::"x" == Bogus2::"y")} == {"a": true} };
