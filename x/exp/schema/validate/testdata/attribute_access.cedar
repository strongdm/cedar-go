// Test 1: Access required attribute - should validate
permit (
    principal,
    action == Action::"view",
    resource
) when {
    principal.name == "alice"
};

// Test 2: Access optional attribute without has guard - should fail
forbid (
    principal,
    action == Action::"view",
    resource
) when {
    principal.age > 18
};
