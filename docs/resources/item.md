---
page_title: "demoapp_item Resource - Demo App"
subcategory: ""
description: |-
  Manages an item in the Demo App inventory.
---

# demoapp_item (Resource)

Manages an item in the Demo App inventory. Items are generic records with a name and description, useful for demonstrating CRUD operations and database functionality.

## Example Usage

### Basic Item

```terraform
resource "demoapp_item" "example" {
  name        = "My First Item"
  description = "Created by Terraform"
}
```

### Multiple Items

```terraform
resource "demoapp_item" "web_server" {
  name        = "Web Server"
  description = "nginx frontend"
}

resource "demoapp_item" "app_server" {
  name        = "App Server"
  description = "Go backend service"
}

resource "demoapp_item" "database" {
  name        = "Database"
  description = "PostgreSQL primary"
}
```

### Dynamic Items

```terraform
variable "services" {
  default = ["api", "web", "worker"]
}

resource "demoapp_item" "service" {
  for_each = toset(var.services)

  name        = "${each.key}-service"
  description = "Service: ${each.key}"
}
```

## Schema

### Required

- `name` (String) The name of the item.

### Optional

- `description` (String) A description of the item.

### Read-Only

- `id` (String) The unique identifier of the item, assigned by Demo App.

## Import

Items can be imported using their ID:

```shell
terraform import demoapp_item.example 123
```
