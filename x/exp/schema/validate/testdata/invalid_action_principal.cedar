// Group used as principal when action only applies to User
permit (
    principal is Group,
    action == Action::"view",
    resource
);
