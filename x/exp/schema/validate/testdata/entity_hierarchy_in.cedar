// principal in with entity hierarchy - User can be in Org
permit (
    principal in Org::"acme",
    action == Action::"view",
    resource is Document
) when { true };

// principal is User in Team - tests isIn and getEntityTypesIn
permit (
    principal is User in Team::"devs",
    action == Action::"view",
    resource is Document
) when { true };
