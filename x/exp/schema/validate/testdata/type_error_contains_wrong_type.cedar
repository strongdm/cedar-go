// Type error: checking if a set of longs contains a string
permit (
    principal,
    action,
    resource
) when {
    [1].contains("test")
};
