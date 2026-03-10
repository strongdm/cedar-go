// typeBool || true → typeTrue (line 254)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (principal == principal) || true };

// typeBool || false → lt (line 258)
permit (
    principal is User,
    action == Action::"view",
    resource is Document
) when { (principal == principal) || false };
