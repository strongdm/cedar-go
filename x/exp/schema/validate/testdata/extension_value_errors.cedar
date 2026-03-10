// Policy with invalid IP
permit(principal, action, resource)
when { ip("not-an-ip") == ip("1.2.3.4") };

// Policy with invalid decimal
permit(principal, action, resource)
when { decimal("not-a-decimal") == decimal("1.23") };

// Policy with invalid datetime
permit(principal, action, resource)
when { datetime("not-a-datetime") == datetime("2024-01-01T00:00:00Z") };

// Policy with invalid duration
permit(principal, action, resource)
when { duration("not-a-duration") == duration("1h") };

// Policy with valid IP
permit(principal, action, resource)
when { ip("1.2.3.4") == ip("1.2.3.5") };

// Policy with valid decimal
permit(principal, action, resource)
when { decimal("1.23") == decimal("4.56") };

// Policy with valid datetime
permit(principal, action, resource)
when { datetime("2024-01-01T00:00:00Z") == datetime("2024-01-02T00:00:00Z") };

// Policy with valid duration
permit(principal, action, resource)
when { duration("1h") == duration("2h") };
