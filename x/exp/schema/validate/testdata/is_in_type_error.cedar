permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { "hello" is User in Document::"doc1" };
