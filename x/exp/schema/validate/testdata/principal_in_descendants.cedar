// Test that principal in Group::"admins" is valid because User is a descendant
// of Group, and the "view" action applies to User.
permit (
    principal in Group::"admins",
    action == Action::"view",
    resource is Document
);

// Test that resource in Folder::"root" is valid because Document is a descendant
// of Folder, and the "view" action applies to Document.
permit (
    principal is User,
    action == Action::"view",
    resource in Folder::"root"
);
