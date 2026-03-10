permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if (1 == 2) then true else (1 + "x" > 0) };
