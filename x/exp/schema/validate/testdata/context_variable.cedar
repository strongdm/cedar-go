permit(principal, action == Action::"view", resource) when { context.role == "admin" };
