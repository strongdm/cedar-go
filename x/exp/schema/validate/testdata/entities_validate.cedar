// Simple policy to validate entities
permit (
    principal is User,
    action == Action::"view",
    resource is User
);
