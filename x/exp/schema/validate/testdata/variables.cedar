// Test that principal and context variables are properly typed
permit(principal, action, resource)
when { principal.name == "alice" && context.ip.isLoopback() };
