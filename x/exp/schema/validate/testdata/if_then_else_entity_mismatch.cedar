permit(principal, action == Action::"view", resource)
when { (if true then principal else "hello") == principal };
