// Policy with invalid action application - Group is not a valid principal for "view"
permit(
  principal is Group,
  action == Action::"view",
  resource
);
