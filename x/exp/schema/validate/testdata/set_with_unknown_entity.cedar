// Policy with a set containing an unknown entity type
permit(principal, action, resource)
when { principal in [Unknown::"x"] };
