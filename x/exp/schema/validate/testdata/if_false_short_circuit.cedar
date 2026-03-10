// if false then 1 else "hello" should type as string (then branch is dead)
// This should validate successfully
permit (
    principal,
    action,
    resource
) when {
    (if false then 1 else "hello") == "hello"
};
