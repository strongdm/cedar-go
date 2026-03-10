// hasTag on entity with tags - should succeed
permit (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  resource.hasTag("category")
};

// hasTag on entity without tags - should give typeFalse
forbid (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  principal.hasTag("x")
};

// hasTag on non-entity (Long) - should error
forbid (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  42.hasTag("x")
};
