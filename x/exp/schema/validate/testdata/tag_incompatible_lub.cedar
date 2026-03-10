permit(principal is TypeA, action == Action::"do", resource)
when {
    (if context.chooser then TypeA::"a" else TypeB::"b").getTag("foo") == "test"
};
