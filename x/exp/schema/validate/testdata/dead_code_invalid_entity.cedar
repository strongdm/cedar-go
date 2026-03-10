// false && with invalid entity type in dead code - should still fail
permit(principal, action == Action::"view", resource)
when { false && (BadType::"id" == User::"alice") };
