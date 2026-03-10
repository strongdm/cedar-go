// Photo can never be "in" Group because Photo and Group are unrelated types
permit (
    principal is Photo in Group::"admins",
    action == Action::"view",
    resource
);
