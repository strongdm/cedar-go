// principal could be User or Admin - is User returns typeBool (line 615)
permit (
    principal,
    action == Action::"view",
    resource is Document
) when { principal is User };
