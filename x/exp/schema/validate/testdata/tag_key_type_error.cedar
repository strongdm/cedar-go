// Using a Long literal as a tag key (should error)
permit (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  resource.hasTag(42)
};
