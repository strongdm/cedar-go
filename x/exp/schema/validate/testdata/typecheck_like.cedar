// Valid: string like pattern
permit(principal, action == App::Action::"view", resource)
when { "hello" like "h*" };

// Invalid: long like pattern
forbid(principal, action == App::Action::"view", resource)
when { 42 like "h*" };
