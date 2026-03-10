// Policy that uses getTag with boolean tag value
permit(principal, action == App::Action::"view", resource)
when {
    resource.getTag("flag") == true
};
