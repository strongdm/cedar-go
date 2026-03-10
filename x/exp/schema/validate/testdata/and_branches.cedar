// true && true → passes (typeTrue)
@id("true_and_true")
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    true && true
};

// true && false → passes (typeFalse)
@id("true_and_false")
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    true && false
};

// false && <non-bool> → passes due to short-circuit (typeFalse)
@id("false_and_entity")
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    false && principal
};

// (1==1) && false → passes (typeFalse)
@id("comparison_and_false")
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    (1 == 1) && false
};

// (1==1) && (1==1) → passes (typeTrue from literal equality)
@id("comparison_and_comparison")
permit (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    (1 == 1) && (1 == 1)
};

// 42 && true → error (left not bool)
@id("non_bool_and_true")
forbid (
    principal is User,
    action == Action::"view",
    resource is User
) when {
    42 && true
};
