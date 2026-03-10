// Policy with reference to unknown entity type
permit (
    principal,
    action,
    resource
) when {
    UnknownEntity::"x" == UnknownEntity::"x"
};
