// Test if-then-else with long values
permit (
    principal,
    action == Action::"view",
    resource
) when {
    (if true then 1 else 2) == 1
};
