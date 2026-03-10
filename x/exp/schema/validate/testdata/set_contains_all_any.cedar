// Test containsAll - user must have all required tags
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    principal.tags.containsAll(context.requiredTags)
};

// Test containsAny - user must have at least one optional tag
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    principal.tags.containsAny(context.optionalTags)
};
