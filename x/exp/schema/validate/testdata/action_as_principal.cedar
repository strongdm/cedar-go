// Action entity used as principal - should fail validation
permit (
    principal == Action::"view",
    action,
    resource
);
