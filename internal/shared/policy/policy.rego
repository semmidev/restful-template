package authz

import rego.v1

default allow = false

# Role-permission mapping (RBAC)
role_permissions := {
    "admin": [
        "todo:create", "todo:read", "todo:update", "todo:delete", "todo:list", "todo:stats",
        "auth:delete_account", "auth:switch_role",
        "user:create", "user:read", "user:update", "user:delete", "user:list"
    ],
    "user": [
        "todo:create", "todo:read", "todo:update", "todo:delete", "todo:list", "todo:stats",
        "auth:delete_account", "auth:switch_role"
    ]
}

# Check if the active role has the permission
role_has_permission(role, perm) if {
    role_permissions[role][_] == perm
}

# Allow evaluation
allow if {
    role_has_permission(input.active_role, input.action)
    is_authorized_for_resource
}

is_authorized_for_resource if {
    # If no resource owner is specified, ownership check is not applicable
    not input.resource_owner_id
}

is_authorized_for_resource if {
    # Admin can access any resource
    input.active_role == "admin"
}

is_authorized_for_resource if {
    # Owner can access their own resource
    input.resource_owner_id == input.user_id
}
