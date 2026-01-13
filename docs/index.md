---
page_title: "Demo App Provider"
description: |-
  The Demo App provider manages resources in Demo App, a universal demo application for infrastructure demonstrations.
---

# Demo App Provider

The Demo App provider allows Terraform to manage resources in [Demo App](https://github.com/billgrant/demo-app), a universal demo application designed for infrastructure, security, and platform demonstrations.

## Why Use This Provider?

Demo App is intentionally stateless — it uses an in-memory database that resets on restart. This provider turns Terraform into the persistence layer:

- Run `terraform apply` to populate your demo environment
- App crashes? Restart and run `terraform apply` again
- Your demo state is restored from Terraform state

**Terraform becomes the source of truth for a stateless app.**

## Example Usage

```terraform
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

resource "demoapp_item" "example" {
  name        = "Provisioned by Terraform"
  description = "Created during demo"
}

resource "demoapp_display" "status" {
  data = jsonencode({
    message = "Hello from Terraform!"
  })
}
```

## Authentication

Demo App does not require authentication. Simply provide the endpoint URL.

## Schema

### Optional

- `endpoint` (String) The base URL of the Demo App API (e.g., `http://localhost:8080`). Can also be set via the `DEMOAPP_ENDPOINT` environment variable.
