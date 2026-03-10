// Record with unknown entity type should fail validation
permit (
    principal == User::"alice",
    action == Action::"view",
    resource == Document::"doc1"
)
when {
    {"x": Unknown::"x"} == {"x": Unknown::"x"}
};
