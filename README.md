# Terraform Provider for Demo App

A Terraform provider for managing resources in [Demo App](https://github.com/billgrant/demo-app), a universal demo application for infrastructure demonstrations.

## Use Case

Demo App is designed to be stateless — it can crash and restart without losing its purpose. However, during demos you often want to show data that was "provisioned." This provider solves that:

1. Run `terraform apply` to populate demo-app with items and display data
2. If demo-app crashes or restarts, run `terraform apply` again
3. Your demo state is restored from Terraform state

**Terraform becomes the persistence layer for a stateless app.**

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Demo App](https://github.com/billgrant/demo-app) running and accessible
- Go >= 1.23 (only for building from source)

## Installation

### From Terraform Registry (Recommended)

```hcl
terraform {
  required_providers {
    demoapp = {
      source  = "billgrant/demoapp"
      version = "~> 1.0"
    }
  }
}
```

### Building from Source

```bash
git clone https://github.com/billgrant/terraform-provider-demoapp.git
cd terraform-provider-demoapp
go build -o terraform-provider-demoapp
```

## Usage

### Provider Configuration

```hcl
provider "demoapp" {
  endpoint = "http://localhost:8080"
}
```

The endpoint can also be set via the `DEMOAPP_ENDPOINT` environment variable.

### Resources

#### demoapp_item

Manages items in the Demo App inventory.

```hcl
resource "demoapp_item" "example" {
  name        = "Provisioned by Terraform"
  description = "This item was created via Terraform"
}
```

**Arguments:**
- `name` (Required) - The name of the item
- `description` (Optional) - A description of the item

**Attributes:**
- `id` - The unique identifier assigned by Demo App

#### demoapp_display

Manages the display panel content. Posts arbitrary JSON that the Demo App frontend renders.

```hcl
resource "demoapp_display" "status" {
  data = jsonencode({
    provisioned_by = "terraform"
    region         = var.region
    timestamp      = timestamp()
  })
}
```

**Arguments:**
- `data` (Required) - JSON string to display. Use `jsonencode()` to convert HCL to JSON.

**Attributes:**
- `id` - Always "display" (singleton resource)

## Example: Full Demo Setup

```hcl
terraform {
  required_providers {
    demoapp = {
      source  = "billgrant/demoapp"
      version = "~> 1.0"
    }
  }
}

provider "demoapp" {
  endpoint = "http://localhost:8080"
}

# Create inventory items
resource "demoapp_item" "web_server" {
  name        = "Web Server"
  description = "Frontend application server"
}

resource "demoapp_item" "database" {
  name        = "Database"
  description = "PostgreSQL primary"
}

# Display provisioning info
resource "demoapp_display" "info" {
  data = jsonencode({
    environment    = "demo"
    provisioned_by = "terraform"
    components     = [
      demoapp_item.web_server.name,
      demoapp_item.database.name
    ]
  })
}

output "item_ids" {
  value = {
    web_server = demoapp_item.web_server.id
    database   = demoapp_item.database.id
  }
}
```

## Development

### Running Locally

1. Build the provider:
   ```bash
   go build -o terraform-provider-demoapp
   ```

2. Create a dev override in `~/.terraformrc`:
   ```hcl
   provider_installation {
     dev_overrides {
       "registry.terraform.io/billgrant/demoapp" = "/path/to/terraform-provider-demoapp"
     }
     direct {}
   }
   ```

3. Run Demo App:
   ```bash
   cd /path/to/demo-app
   ./demo-app
   ```

4. Test with Terraform:
   ```bash
   cd examples
   terraform plan
   terraform apply
   ```

### Running Tests

```bash
go test ./...
```

## Known Issues

### SQLite Concurrency

Demo App uses SQLite which has limited concurrent write support. When creating multiple resources in parallel, you may see database errors.

**Workaround:** Use `-parallelism=1`:

```bash
terraform apply -parallelism=1
```

This will be fixed in a future Demo App release.

## License

MIT
