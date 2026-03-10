// Test valid logical AND operation - should pass
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { true && false };

// Test invalid logical AND - right side is not boolean, should fail
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { true && 42 };
