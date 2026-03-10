permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if 1 then true else false };
