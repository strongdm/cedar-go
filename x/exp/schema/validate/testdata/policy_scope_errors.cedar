// Unknown entity in principal scope
permit (principal == Unknown::"x", action, resource);

// Unknown action in set
permit (principal, action in [Action::"delete"], resource);

// Unknown entity in resource scope
permit (principal, action, resource == Unknown::"x");

// Unknown resource type in is
permit (principal, action, resource is Unknown);

// Unknown entity in resource in scope
permit (principal, action, resource in Unknown::"x");

// Unknown entity in principal in scope
permit (principal in Unknown::"x", action, resource);

// Enum in principal scope - valid enum value
permit (principal == Color::"red", action, resource);

// Enum in principal scope - invalid enum value
permit (principal == Color::"purple", action, resource);

// resource is...in invalid
permit (principal, action, resource is Document in User::"alice");

// resource is...in unknown in-entity
permit (principal, action, resource is Document in Unknown::"x");
