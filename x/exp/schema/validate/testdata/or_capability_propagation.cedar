// Valid: both OR branches provide the same capability (age)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    (principal has age || principal has age) && principal.age > 18
};

// Invalid: mismatched capabilities - one branch has age, other has name
// The intersection of capabilities is empty, so access to age fails
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    (principal has age || principal has name) && principal.age > 18
};
