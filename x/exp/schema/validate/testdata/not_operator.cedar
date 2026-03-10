// Test: !true → typeFalse
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !true
};

// Test: !false → typeTrue
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !false
};

// Test: !(1 == 1) → typeFalse (constant folding)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !(1 == 1)
};

// Test: !42 → error (not a boolean)
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    !42
};
