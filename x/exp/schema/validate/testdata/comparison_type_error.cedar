// Valid: long comparison
permit(principal, action == App::Action::"view", resource)
when { 1 < 2 };

// Invalid: string on right side of less than
forbid(principal, action == App::Action::"view", resource)
when { 1 < "not a long" };

// Invalid: string on left side of less than
forbid(principal, action == App::Action::"view", resource)
when { "not a long" < 1 };

// Invalid: both sides wrong type
forbid(principal, action == App::Action::"view", resource)
when { "hello" > "world" };
