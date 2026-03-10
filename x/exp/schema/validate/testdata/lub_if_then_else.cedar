// Test LUB of true|false (should produce Bool)
permit (
    principal,
    action,
    resource
) when {
    if true then true else false
};
