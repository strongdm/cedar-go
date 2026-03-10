permit(principal, action == Action::"view", resource)
when { principal has name && principal has nickname };
