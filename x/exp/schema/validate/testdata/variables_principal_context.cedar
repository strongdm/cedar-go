// Policy that uses principal variable
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { principal == principal };

// Policy that uses context variable
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { context.ip == context.ip };
