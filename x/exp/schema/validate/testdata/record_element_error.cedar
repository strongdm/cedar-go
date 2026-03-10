permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { {"a": 1 + "x"} == {"a": 1} };
