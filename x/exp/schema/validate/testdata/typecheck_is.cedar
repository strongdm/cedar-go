// Test principal is User - should always be true
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { principal is User };

// Test principal is Photo - should always be false
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { principal is Photo };
