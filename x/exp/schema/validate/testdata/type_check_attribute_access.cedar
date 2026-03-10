// Test accessing required attribute - should work
permit(
  principal is MyApp::User,
  action == MyApp::Action::"view",
  resource is MyApp::Document
)
when {
  principal.name == "test"
};

// Test accessing optional attribute without has guard - should error
permit(
  principal is MyApp::User,
  action == MyApp::Action::"view",
  resource is MyApp::Document
)
when {
  principal.age > 18
};
