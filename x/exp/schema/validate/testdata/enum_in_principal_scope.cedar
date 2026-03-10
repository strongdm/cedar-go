// Test that enum types can be used in principal scope
permit (
    principal is Test::Color,
    action == Test::Action::"view",
    resource
);
