// Exercises many node types in dead code via false &&
permit(principal is User, action == Action::"view", resource)
when {
  false && (
    User::"a" == User::"b" &&
    User::"a" != User::"b" &&
    1 < 2 &&
    1 <= 2 &&
    1 > 0 &&
    1 >= 0 &&
    1 + 1 == 2 &&
    2 - 1 == 1 &&
    2 * 1 == 2 &&
    User::"a" in User::"b" &&
    [User::"a"].contains(User::"a") &&
    [User::"a"].containsAll([User::"b"]) &&
    [User::"a"].containsAny([User::"b"]) &&
    User::"a" has age &&
    User::"a".age > 0 &&
    -1 < 0 &&
    !false &&
    "hello" like "h*" &&
    User::"a" is User &&
    User::"a" is User in Group::"g" &&
    [].isEmpty() &&
    ip("1.2.3.4").isIpv4() &&
    (if true then true else false) &&
    {name: "test"}.name == "test" &&
    User::"a".hasTag("x") &&
    (User::"a".hasTag("x") && User::"a".getTag("x") == "v")
  )
};
