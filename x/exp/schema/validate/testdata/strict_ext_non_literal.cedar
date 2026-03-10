permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal has addr && ip(principal.addr).isIpv4() };
