// Test Never <: Long: if false then 1 else 2 should have type Long
permit(principal, action, resource) 
when { (if false then 1 else 2) > 0 };

// Test Never <: Bool: if false then false else false should have type Bool
permit(principal, action, resource) 
when { if false then false else false };

// Test True <: Bool: true is subtype of Bool
permit(principal, action, resource) 
when { true };

// Test False <: Bool: false is subtype of Bool
permit(principal, action, resource) 
when { false };

// Test Bool <: Bool: (true || false) is Bool subtype of Bool
permit(principal, action, resource) 
when { (true || false) };

// Test Long <: Long: 1 is subtype of Long
permit(principal, action, resource) 
when { 1 > 0 };

// Test String <: String: "test" is subtype of String
permit(principal, action, resource) 
when { "test" == "test" };

// Test Set <: Set: [1] is subtype of Set<Long>
permit(principal, action, resource) 
when { [1].contains(2) };

// Test Entity <: Entity: User::"alice" is subtype of User
permit(principal, action, resource) 
when { principal == User::"alice" };
