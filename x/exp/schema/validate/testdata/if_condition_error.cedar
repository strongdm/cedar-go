permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if (Bogus::"x" in User::"y") then true else false };
