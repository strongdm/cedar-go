// Policy 0: Access required attribute - should pass
permit (
    principal,
    action == Action::"view",
    resource
) when {
    principal.name == "alice"
};

// Policy 1: Access optional attribute without has guard - should fail
permit (
    principal,
    action == Action::"view",
    resource
) when {
    principal.age > 18
};
