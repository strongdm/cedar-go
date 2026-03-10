// Test that variables are correctly typed
// principal should be typed as User
// context should be typed as a record with ip: ipaddr
permit (
    principal,
    action == Action::"view",
    resource
)
when {
    principal has name &&
    context.ip.isIpv4()
};
