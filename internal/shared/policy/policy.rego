package authz

import rego.v1

default allow = false

# Access permissions from the dynamically loaded permissions package
role_permissions := data.permissions.role_permissions

# Check if any of the user's roles has the permission
role_has_permission(roles, perm) if {
    some role in roles
    role_permissions[role][_] == perm
}

# Allow evaluation
allow if {
    role_has_permission(input.roles, input.action)
    is_authorized_for_resource
}

is_authorized_for_resource if {
    # If no resource owner is specified, ownership check is not applicable
    not input.resource.owner_id
    not input.resource_owner_id
}

is_authorized_for_resource if {
    # Admin can access any resource
    "admin" in input.roles
}

is_authorized_for_resource if {
    # Owner can access their own resource (legacy)
    input.resource_owner_id == input.user_id
}

is_authorized_for_resource if {
    # Owner can access their own resource (new nested)
    input.resource.owner_id == input.user_id
}
