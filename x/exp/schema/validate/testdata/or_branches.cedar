// Test: false || true → should pass (evaluates to true)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { false || true };

// Test: false || false → should pass (evaluates to false)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { false || false };

// Test: (1 == 1) || true → should pass (evaluates to true)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (1 == 1) || true };

// Test: (1 == 1) || false → should pass (bool)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { (1 == 1) || false };

// Test: 42 || true → should fail (left operand is not boolean)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { 42 || true };

// Test: true || User::"alice" → should pass (short-circuit, RHS not evaluated)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { true || User::"alice" };

// Test: false || 42 → should fail (right operand is not boolean)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { false || 42 };
