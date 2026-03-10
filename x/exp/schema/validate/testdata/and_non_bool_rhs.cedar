// Test that && with non-bool right hand side produces type error
permit (
  principal,
  action == Action::"view",
  resource
) when {
  true && User::"alice"
};
