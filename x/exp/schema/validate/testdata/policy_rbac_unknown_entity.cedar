// Policy with unknown entity type in principal scope
permit(
  principal is Unknown,
  action,
  resource
);
