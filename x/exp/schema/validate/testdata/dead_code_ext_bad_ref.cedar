// true || (expr with invalid entity ref in extension call context)
// The or short-circuits to true, but validateEntityRefs is called on RHS
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { true || (Bogus::"x" == User::"y") };
