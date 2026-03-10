// Test 1: if false then 1 else "hello" → short-circuit to else (typeString)
permit (
    principal,
    action,
    resource
) when {
    (if false then 1 else "hello") == "hello"
};

// Test 2: if true then 1 else "hello" → short-circuit to then (typeLong)
permit (
    principal,
    action,
    resource
) when {
    (if true then 1 else "hello") == 1
};

// Test 3: if 42 then 1 else 2 → error (condition not bool)
permit (
    principal,
    action,
    resource
) when {
    (if 42 then 1 else 2) == 1
};

// Test 4: if (1==1) then 1 else "hello" → short-circuit to then (constant folding)
permit (
    principal,
    action,
    resource
) when {
    (if (1 == 1) then 1 else "hello") == 1
};

// Test 5: Dead else branch with bad entity ref → error
permit (
    principal,
    action,
    resource
) when {
    (if true then 1 else Unknown::"x") == 1
};

// Test 6: Dead then branch with bad entity ref → error
permit (
    principal,
    action,
    resource
) when {
    (if false then Unknown::"x" else 1) == 1
};

// Test 7: if (1==1) then 1 else (1+"x") → no error (else branch dead, type error not checked)
permit (
    principal,
    action,
    resource
) when {
    (if (1 == 1) then 1 else (1 + "x")) == 1
};

// Test 8: Non-statically-known condition with type error in else → error
permit (
    principal,
    action,
    resource
) when {
    (if principal has name then 1 else (1 + "x")) == 1
};
