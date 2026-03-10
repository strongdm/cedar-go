// Test that capability tracking properly handles prior capabilities
// The else branch should get the capability from the outer && (principal has age)
// but not the test capability (principal has name)
permit(principal, action, resource)
when {
  principal has age && 
  (if principal has name then false else principal.age > 0)
};
