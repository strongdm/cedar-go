// Test basic type compatibility through control flow

// true and false are subtypes of Bool
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    if true then false else true
};

// Long is subtype of Long
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    (if principal has name then 1 else 2) == 1
};

// String is subtype of String
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    (if principal has name then "a" else "b") == "a"
};

// Set<Long> is subtype of Set<Long>
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    (if principal has name then [1, 2] else [3, 4]).contains(1)
};

// Entity types are compatible
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    principal == (if principal has name then principal else resource)
};
