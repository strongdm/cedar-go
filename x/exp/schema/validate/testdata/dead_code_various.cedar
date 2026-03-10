// false && (complex dead expression with entity refs)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (principal in User::"admin") };

// true || (dead branch with entity ref in set)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { true || (principal in [User::"alice", User::"bob"]) };

// if true then ... else (dead branch with record containing entity ref)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if true then true else (User::"x" == User::"y") };

// if false then (dead branch) else ...
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { if false then (User::"x" == User::"y") else true };

// false && (has check in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && principal has name };

// false && (like check in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ("test" like "*") };

// false && (is check in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (principal is User) };

// false && (negation in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (-(1) > 0) };

// false && (not in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && !true };

// false && (isEmpty in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && [].isEmpty() };

// false && (comparison in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (1 < 2) };

// false && (arithmetic in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (1 + 2 > 0) };

// false && (contains in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && [1].contains(1) };

// false && (containsAll in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && [1].containsAll([1]) };

// false && (containsAny in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && [1].containsAny([1]) };

// false && (if-then-else in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (if true then true else false) };

// false && (record in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ({"a": 1} == {"a": 1}) };

// false && (set in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ([1, 2] == [1, 2]) };

// false && (extension in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ip("1.2.3.4").isIpv4() };

// false && (isIn in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (principal is User in User::"admin") };

// false && (hasTag in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && principal.hasTag("x") };

// false && (getTag in dead branch)  
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (principal.getTag("x") == "y") };

// false && (notEquals in dead branch)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (1 != 2) };
