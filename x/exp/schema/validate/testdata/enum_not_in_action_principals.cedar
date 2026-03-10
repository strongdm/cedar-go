// Test that enum entities cannot be used in principal scope
// if the action doesn't allow them
permit (
    principal == Color::"red",
    action,
    resource
);
