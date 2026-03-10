// principal is User, check if principal in Org::"x" - should pass (User in Team in Org)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal in Org::"acme" };

// Test 'in' with impossible hierarchy - User cannot be in Document
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal in Document::"doc1" };
