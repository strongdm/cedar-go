// Test literal equality evaluation
// 42 == 42 should evaluate to true (short-circuit)
permit (
    principal,
    action,
    resource
)
when {
    42 == 42
};

// 1 == 2 should evaluate to false
forbid (
    principal,
    action,
    resource
)
when {
    1 == 2
};

// 1 != 1 should evaluate to false
forbid (
    principal,
    action,
    resource
)
when {
    1 != 1
};

// 1 != 2 should evaluate to true
permit (
    principal,
    action,
    resource
)
when {
    1 != 2
};

// principal == User::"a" should be typeBool (not literal)
permit (
    principal,
    action,
    resource
)
when {
    principal == User::"a"
};
