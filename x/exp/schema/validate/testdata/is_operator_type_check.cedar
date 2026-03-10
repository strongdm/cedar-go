// principal is User should typecheck as true (principal type is User)
permit(principal is User, action == Action::"view", resource);
// principal is Photo should typecheck as false (principal type is User, not Photo)
permit(principal is Photo, action == Action::"view", resource);
