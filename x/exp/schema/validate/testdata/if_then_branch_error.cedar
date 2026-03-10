permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if (1 == 1) then (1 + "x" > 0) else true };
