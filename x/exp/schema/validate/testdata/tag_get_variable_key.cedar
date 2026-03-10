// Using a variable as a tag key in getTag (should error for unsafe access)
permit (
  principal is User,
  action == Action::"view",
  resource is Document
) when {
  resource.getTag(context.tagName) == "value"
};
