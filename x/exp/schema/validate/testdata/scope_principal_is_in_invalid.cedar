// Principal is Photo in Group::"admins" - Photo can never be "in" Group
permit (
    principal is Photo in Group::"admins",
    action == Action::"view",
    resource
);
