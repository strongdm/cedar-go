// Policy with unknown entity type in principal constraint
permit (
    principal is Unknown,
    action,
    resource
);
