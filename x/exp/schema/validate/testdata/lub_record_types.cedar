permit(principal, action == Action::"view", resource)
when { (if true then {a: 1, b: "x"} else {a: 2, c: true}) has a };
