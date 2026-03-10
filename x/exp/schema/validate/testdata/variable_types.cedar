// Test that variables principal, action, resource, and context have the correct types
// and that unknown variables result in typeNever
permit (
  principal,
  action == Action::"view",
  resource
) when {
  // These should type-check: action is Action, resource is Document
  action == Action::"view" &&
  resource is Document
};
