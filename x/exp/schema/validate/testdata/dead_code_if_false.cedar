// if false then dead-then else live-else
permit(principal, action == Action::"view", resource)
when { (if false then (User::"alice" == User::"bob") else true) };
