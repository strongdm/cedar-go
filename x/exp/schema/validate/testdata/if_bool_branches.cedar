permit(principal, action == Action::"view", resource) when {
    (if principal.active then false else false)
};

permit(principal, action == Action::"view", resource) when {
    (if principal.active then principal.premium else principal.active)
};
