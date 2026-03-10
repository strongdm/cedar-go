// Valid: principal in entity
permit (
    principal,
    action,
    resource
) when {
    principal in Group::"admins"
};

// Valid: principal in set of entities
permit (
    principal,
    action,
    resource
) when {
    principal in [Group::"a", Group::"b"]
};

// Invalid: non-entity in entity
permit (
    principal,
    action,
    resource
) when {
    42 in Group::"admins"
};

// Invalid: entity in non-entity/non-set
permit (
    principal,
    action,
    resource
) when {
    principal in "admins"
};

// Invalid: entity in set of non-entities
permit (
    principal,
    action,
    resource
) when {
    principal in [42]
};
