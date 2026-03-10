permit (
    principal,
    action == Action::"view",
    resource
) when {
    ip("192.168.1.1").isIpv4()
};
