permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ip("1.2.3.4").isInRange(ip(Bogus::"x")) };
