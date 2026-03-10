permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (if principal has "flag" then principal else action).hasTag("x") };
