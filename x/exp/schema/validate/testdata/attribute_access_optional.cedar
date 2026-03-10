// Policy 1: Access required attribute (should validate)
permit(principal, action == Test::Action::"view", resource)
when { principal.name == "Alice" };

// Policy 2: Access optional attribute without has guard (should fail in strict mode)
permit(principal, action == Test::Action::"view", resource)
when { principal.age > 18 };
