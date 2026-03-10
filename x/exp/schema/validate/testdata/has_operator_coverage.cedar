// Test various has operator scenarios

// Required attribute - should always be true
permit (
    principal,
    action,
    resource
) when {
    principal has name
};

// Optional attribute - should be bool
permit (
    principal,
    action,
    resource
) when {
    principal has age
};

// Context with required attribute
permit (
    principal,
    action,
    resource
) when {
    context has ip
};

// Context with optional attribute
permit (
    principal,
    action,
    resource
) when {
    context has requestTime
};

// Non-existent attribute on closed record
permit (
    principal,
    action,
    resource
) when {
    principal has unknownAttr
};
