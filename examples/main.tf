terraform {
  required_providers {
    demoapp = {
      source = "billgrant/demoapp"
    }
  }
}

provider "demoapp" {
  endpoint = "http://localhost:8080"
}

# Create some items
resource "demoapp_item" "example" {
  name        = "Provisioned by Terraform"
  description = "This item was created by the Terraform provider"
}

resource "demoapp_item" "another" {
  name        = "Another Item"
  description = "Demonstrating multiple resources"
}

# Display provisioning info
resource "demoapp_display" "status" {
  data = jsonencode({
    provisioned_by = "terraform"
    provider       = "demoapp"
    items_created  = 2
    message        = "Hello from Terraform!"
  })
}

# Output the IDs
output "item_ids" {
  value = {
    example = demoapp_item.example.id
    another = demoapp_item.another.id
  }
}
