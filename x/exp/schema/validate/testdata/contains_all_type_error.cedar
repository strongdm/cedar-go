// Policy with type error: containsAll comparing Set<Long> with Set<String>
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    [1, 2, 3].containsAll(["a", "b"])
};
