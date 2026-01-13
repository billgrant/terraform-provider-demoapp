---
page_title: "demoapp_display Resource - Demo App"
subcategory: ""
description: |-
  Manages the display panel content in Demo App.
---

# demoapp_display (Resource)

Manages the display panel content in Demo App. The display panel shows arbitrary JSON data, making it perfect for displaying Terraform outputs, deployment information, or any custom data during demos.

~> **Note:** There is only one display panel in Demo App. Creating multiple `demoapp_display` resources will result in each overwriting the previous one.

## Example Usage

### Basic Display

```terraform
resource "demoapp_display" "main" {
  data = jsonencode({
    message = "Hello from Terraform!"
  })
}
```

### Terraform Outputs as Display

```terraform
resource "demoapp_display" "terraform_info" {
  data = jsonencode({
    workspace      = terraform.workspace
    provisioned_by = "terraform"
    timestamp      = timestamp()
  })
}
```

### Display with Resource References

```terraform
resource "demoapp_item" "web" {
  name = "Web Server"
}

resource "demoapp_item" "db" {
  name = "Database"
}

resource "demoapp_display" "infrastructure" {
  data = jsonencode({
    environment = "demo"
    components = {
      web_server = {
        id   = demoapp_item.web.id
        name = demoapp_item.web.name
      }
      database = {
        id   = demoapp_item.db.id
        name = demoapp_item.db.name
      }
    }
    deployed_at = timestamp()
  })
}
```

### External Data Display

```terraform
# Display outputs from other Terraform resources
resource "demoapp_display" "cloud_info" {
  data = jsonencode({
    cloud_provider = "AWS"
    region         = var.region
    vpc_id         = aws_vpc.main.id
    instance_ids   = aws_instance.app[*].id
  })
}
```

## Schema

### Required

- `data` (String) JSON string to display. Use `jsonencode()` to convert HCL objects to JSON.

### Read-Only

- `id` (String) Always "display" since there is only one display panel.

## Import

The display can be imported using the fixed ID "display":

```shell
terraform import demoapp_display.main display
```
