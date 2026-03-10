permit(principal, action == Action::"view", resource) when {
    (if principal.active then true else true)
};
