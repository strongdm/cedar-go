// Test case 1: !true (should validate - results in false)
permit (
    principal,
    action,
    resource
) when {
    !true
};

// Test case 2: !false (should validate - results in true)
permit (
    principal,
    action,
    resource
) when {
    !false
};

// Test case 3: !(1 == 1) (should validate - boolean negation)
permit (
    principal,
    action,
    resource
) when {
    !(1 == 1)
};

// Test case 4: !42 (should error - cannot negate a number)
permit (
    principal,
    action,
    resource
) when {
    !42
};
