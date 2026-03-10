// Test comparison operators typecheck to bool

// Greater than
permit (
    principal,
    action,
    resource
)
when {
    1 > 2
};

// Greater than or equal
permit (
    principal,
    action,
    resource
)
when {
    1 >= 2
};

// Less than
permit (
    principal,
    action,
    resource
)
when {
    1 < 2
};

// Less than or equal
permit (
    principal,
    action,
    resource
)
when {
    1 <= 2
};
