// Policy with unknown entity type in principal
permit(principal is Unknown, action, resource);

// Policy with unknown action
permit(principal, action == Action::"delete", resource);

// Policy with invalid action application (Group instead of User)
permit(principal is Group, action == Action::"view", resource);

// Valid policy
permit(principal is User, action == Action::"view", resource is Document);
