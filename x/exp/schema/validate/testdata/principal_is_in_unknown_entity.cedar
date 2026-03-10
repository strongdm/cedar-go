// Test that principal is User in Unknown::"x" fails validation
// because Unknown is not a known entity type
permit (
    principal is Test::User in Test::Unknown::"x",
    action == Test::Action::"view",
    resource is Test::Document
);
