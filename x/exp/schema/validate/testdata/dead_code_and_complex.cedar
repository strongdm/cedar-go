// false && complex expressions - dead code branch validation
@id("dead_and_equality")
permit(principal, action == Action::"view", resource)
when { false && (User::"alice" == User::"bob") };

@id("dead_and_arithmetic")
permit(principal, action == Action::"view", resource)
when { false && (1 + 2 > 0) };

@id("dead_and_contains")
permit(principal, action == Action::"view", resource)
when { false && [User::"alice"].contains(principal) };

@id("dead_and_containsAll")
permit(principal, action == Action::"view", resource)
when { false && [User::"alice"].containsAll([User::"bob"]) };

@id("dead_and_containsAny")
permit(principal, action == Action::"view", resource)
when { false && [User::"alice"].containsAny([User::"bob"]) };

@id("dead_and_has")
permit(principal, action == Action::"view", resource)
when { false && ({name: "test"} has name) };

@id("dead_and_not")
permit(principal, action == Action::"view", resource)
when { false && !true };

@id("dead_and_negate")
permit(principal, action == Action::"view", resource)
when { false && (-1 == -1) };

@id("dead_and_access")
permit(principal, action == Action::"view", resource)
when { false && (principal.age > 0) };

@id("dead_and_like")
permit(principal, action == Action::"view", resource)
when { false && ("hello" like "h*") };

@id("dead_and_is")
permit(principal, action == Action::"view", resource)
when { false && (principal is User) };

@id("dead_and_is_in")
permit(principal, action == Action::"view", resource)
when { false && (principal is User in User::"alice") };

@id("dead_and_isEmpty")
permit(principal, action == Action::"view", resource)
when { false && [].isEmpty() };

@id("dead_and_ip")
permit(principal, action == Action::"view", resource)
when { false && ip("1.2.3.4").isIpv4() };

@id("dead_and_if_then_else")
permit(principal, action == Action::"view", resource)
when { false && (if true then true else false) };
