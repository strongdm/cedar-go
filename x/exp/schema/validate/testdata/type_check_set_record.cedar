// Test set literal typechecking
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { [1, 2, 3].contains(2) };

// Test record literal typechecking
permit (
    principal is User,
    action == Action::"view",
    resource is Document
)
when { {"a": 1, "b": "hello"}.a == 1 };
