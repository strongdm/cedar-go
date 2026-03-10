permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    context.access_time > datetime("2024-01-01T00:00:00Z") &&
    context.timeout > duration("1h")
};
