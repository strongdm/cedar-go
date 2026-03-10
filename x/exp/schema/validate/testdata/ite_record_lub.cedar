permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (if principal has "flag" then {"a": 1} else {"b": 2}) == {"a": 1} };
