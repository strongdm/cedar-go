permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { Bogus::"x" is User in Bogus2::"y" };
