permit (
    principal is User,
    action == Action::"view",
    resource is Document in Folder::"root"
);
