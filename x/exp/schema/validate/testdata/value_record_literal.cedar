permit (
    principal,
    action == Action::"view",
    resource
) when {
    {name: "test"} has name
};
