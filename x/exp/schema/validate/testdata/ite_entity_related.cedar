permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (if principal has "flag" then User::"alice" else Group::"admins") == principal };
