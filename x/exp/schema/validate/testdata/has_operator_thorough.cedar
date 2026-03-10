// Test has on required attribute
permit(principal, action, resource)
when { principal has name };

// Test has on optional attribute  
permit(principal, action, resource)
when { principal has age };

// Test has on non-existent attribute
permit(principal, action, resource)
when { principal has missing };

// Test has on resource
permit(principal, action, resource)
when { resource has name };

// Test has on resource non-existent
permit(principal, action, resource)
when { resource has missing };
