// Test that if-then-else with same capability in both branches preserves it
// Both branches check `principal has age`, so `principal.age` is safe to access
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    (if 1 == 1 then (principal has age && true) else (principal has age && true)) && principal.age > 0
};
