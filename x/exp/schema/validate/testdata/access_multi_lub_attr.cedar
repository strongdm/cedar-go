permit (
    principal is A,
    action == Action::"do",
    resource is C
) when { (if principal has "flag" then A::"a" else B::"b").x == 1 };
