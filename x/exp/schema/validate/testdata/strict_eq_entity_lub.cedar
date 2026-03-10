// Non-literal cross-kind equality: attr is String, compared to Long
// Rust treats this as disjoint types → typeFalse, so it passes strict mode
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { principal.name == 42 };
