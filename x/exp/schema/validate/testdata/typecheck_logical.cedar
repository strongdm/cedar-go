// Valid logical AND with bool operands
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && false
};

// Invalid logical AND with non-bool operand
@id("invalid_and")
forbid (
    principal is User,
    action == Action::"view",
    resource is Document
) when {
    true && 42
};
