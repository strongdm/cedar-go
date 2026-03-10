// Access shared attribute 'name' on a multi-type principal
permit (
    principal,
    action == Action::"view",
    resource is Document
) when { principal has name && principal.name == "admin" };
