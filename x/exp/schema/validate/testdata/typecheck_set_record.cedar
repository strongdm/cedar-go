// Test set type-checking [1, 2, 3] → typeSet{Long}
permit(
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    [1, 2, 3].contains(1)
};

// Test record type-checking {"a": 1, "b": "hello"}
permit(
    principal is User,
    action == Action::"view",
    resource is Document
)
when {
    {"a": 1, "b": "hello"}.a == 1
};
