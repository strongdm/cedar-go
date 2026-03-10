// Policy with invalid IP address literal
permit(principal, action, resource)
when { ip("not-a-valid-ip") == ip("192.168.1.1") };
