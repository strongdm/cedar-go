// true || expr - dead code on right side of ||
permit(principal, action == Action::"view", resource)
when { true || (User::"alice" == User::"bob") };
