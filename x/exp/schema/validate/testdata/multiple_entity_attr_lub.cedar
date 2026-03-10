permit(principal, action == Action::"view", resource) when {
    (if true then User::"a" else Admin::"b").name == "test"
};
