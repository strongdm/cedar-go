permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal is User in Bogus::"y" };
