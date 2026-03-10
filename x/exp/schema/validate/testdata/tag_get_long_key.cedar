// Using a Long literal as a tag key in getTag (should error for both type and unsafe access)
permit (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  resource.getTag(42) == "value"
};
