// Negate with both a type error in the argument AND a type mismatch
// This tests the error.Join path in typeOfNegate
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    -(1 + "bad") > 0
};
