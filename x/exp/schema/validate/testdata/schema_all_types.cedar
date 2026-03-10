permit(principal, action == Action::"read", resource) when {
    resource.name == "x" && resource.count > 0 && resource.active
};
