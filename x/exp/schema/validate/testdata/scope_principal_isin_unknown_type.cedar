// This should fail: Unknown type in principal is...in
permit (
    principal is Unknown in Group::"admins",
    action,
    resource
);
