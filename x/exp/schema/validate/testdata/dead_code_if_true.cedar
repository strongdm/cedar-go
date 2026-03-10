// if true then live-then else dead-else
permit(principal, action == Action::"view", resource)
when { (if true then true else (User::"alice" == User::"bob")) };
