// Access 'email' which exists on User but not Admin
// With multi-type principal (ScopeTypeAll), in User env this should type-check,
// in Admin env it should return typeNever (attr doesn't exist on all types)
permit (
    principal,
    action == Action::"view",
    resource is Document
) when { principal has email && principal.email == "x" };
