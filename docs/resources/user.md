---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "mongodb-users_user Resource - mongodb-users"
subcategory: ""
description: |-
  
---

# mongodb-users_user (Resource)



## Example Usage

```terraform
resource "mongodb-users_user" "user1" {
  user     = "user1"
  db       = "test"
  password = "abc123"
  roles = [
    {
      db   = "test"
      role = "readWrite"
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `db` (String) DB Where the user is registered
- `password` (String, Sensitive) Password of user, cannot be changed once set
- `roles` (Set of Object) Set of roles that the user has (see [below for nested schema](#nestedatt--roles))
- `user` (String) Name of user

### Read-Only

- `id` (String) Placeholder identifier attribute
- `last_updated` (String) Timestamp of the last Terraform update of the order.

<a id="nestedatt--roles"></a>
### Nested Schema for `roles`

Required:

- `db` (String)
- `role` (String)

## Import

Import is supported using the following syntax:

```shell
terraform import mongodb-users_user.user1 test.user1
```