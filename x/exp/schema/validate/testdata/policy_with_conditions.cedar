// Policy with valid when condition
permit(principal is User, action == Action::"view", resource is Document) when { 1 == 1 };

// Policy with invalid when condition (adding Long and String)
permit(principal is User, action == Action::"view", resource is Document) when { 1 + "x" };
