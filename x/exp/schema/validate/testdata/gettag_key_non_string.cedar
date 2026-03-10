permit(principal, action == Action::"view", resource) when { principal.hasTag("x") && principal.getTag(1) == "y" };
