// dead code if-then-else with bad entity ref in condition
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (if (Bogus::"x" == User::"y") then true else false) };

// dead code record with bad entity ref
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ({"a": Bogus::"x"} == {"a": User::"y"}) };

// dead code set with bad entity ref
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && ([Bogus::"x"] == [User::"y"]) };

// dead code if-then-else with bad entity ref in then branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (if true then (Bogus::"x" == User::"y") else false) };

// dead code if-then-else with bad entity ref in else branch
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { false && (if false then true else (Bogus::"x" == User::"y")) };
