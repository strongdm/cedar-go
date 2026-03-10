// Test: negate an expression that itself has a type error
// This tests the error joining path in typeOfNegate
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    // principal.unknownAttr produces a type error (attribute doesn't exist)
    // AND we negate it, which produces another error (not a Long)
    -principal.unknownAttr
};
