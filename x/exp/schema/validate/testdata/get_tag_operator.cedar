// Test getTag without hasTag guard - should error
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.getTag("category") == "public"
};

// Test getTag on entity that doesn't support tags - should error
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    principal.getTag("x") == "y"
};

// Test getTag on non-entity (Long) - should error
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    42.getTag("x") == "y"
};

// Test getTag with non-string key - should error
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    resource.hasTag("category") && resource.getTag(42) == "public"
};
