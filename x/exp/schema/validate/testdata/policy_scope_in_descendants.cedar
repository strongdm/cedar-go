// principal in Group::"admins" should be valid because User is a descendant
// of Group, and the "view" action applies to User.
permit (
    principal in Group::"admins",
    action == Action::"view",
    resource is Document
);

// resource in Folder::"root" should be valid because Document is a descendant
// of Folder, and the "view" action applies to Document.
permit (
    principal is User,
    action == Action::"view",
    resource in Folder::"root"
);
