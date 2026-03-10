permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if (principal == principal) then (1 + "x" > 0) else true };
