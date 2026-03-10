// Test that accessing required attribute works
permit (
    principal,
    action,
    resource
)
when {
    principal.name == "Alice"
};

// Test that accessing optional attribute without has guard produces error
forbid (
    principal,
    action,
    resource
)
when {
    principal.age > 18
};
