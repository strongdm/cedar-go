// principal is Photo in Group::"admins" should fail
// Photo can never be "in" Group (unrelated types)
permit (
    principal is Photo in Group::"admins",
    action == Action::"view",
    resource
);
