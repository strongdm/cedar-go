// Empty policy - we're just validating entities
permit (
    principal is User,
    action == Action::"view",
    resource is User
);
