// Policy with unknown entity type - should fail
permit (
    principal is Unknown,
    action,
    resource
);

// Policy with unknown action - should fail
permit (
    principal,
    action == Action::"delete",
    resource
);

// Policy with invalid action application (Group not allowed for "view") - should fail
permit (
    principal is Group,
    action == Action::"view",
    resource
);

// Valid simple policy - should pass
permit (
    principal is User,
    action == Action::"view",
    resource is Document
);
