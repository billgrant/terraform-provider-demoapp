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

# Create multiple items to test concurrent writes
resource "demoapp_item" "web_server" {
  name        = "Web Server"
  description = "Frontend application server"
}

resource "demoapp_item" "api_gateway" {
  name        = "API Gateway"
  description = "Kong API gateway"
}

resource "demoapp_item" "database" {
  name        = "Database"
  description = "PostgreSQL primary"
}

resource "demoapp_item" "cache" {
  name        = "Cache"
  description = "Redis cluster"
}

resource "demoapp_item" "queue" {
  name        = "Message Queue"
  description = "RabbitMQ broker"
}

resource "demoapp_item" "storage" {
  name        = "Object Storage"
  description = "S3-compatible storage"
}

resource "demoapp_item" "monitoring" {
  name        = "Monitoring"
  description = "Prometheus + Grafana stack"
}

resource "demoapp_item" "logging" {
  name        = "Logging"
  description = "ELK stack"
}

resource "demoapp_item" "vault" {
  name        = "Secrets Manager"
  description = "HashiCorp Vault"
}

resource "demoapp_item" "consul" {
  name        = "Service Mesh"
  description = "HashiCorp Consul"
}

# Display provisioning info
resource "demoapp_display" "status" {
  data = jsonencode({
    provisioned_by = "terraform"
    provider       = "demoapp"
    items_created  = 10
    message        = "Infrastructure provisioned!"
  })
}

# Output the IDs
output "item_ids" {
  value = {
    web_server  = demoapp_item.web_server.id
    api_gateway = demoapp_item.api_gateway.id
    database    = demoapp_item.database.id
    cache       = demoapp_item.cache.id
    queue       = demoapp_item.queue.id
    storage     = demoapp_item.storage.id
    monitoring  = demoapp_item.monitoring.id
    logging     = demoapp_item.logging.id
    vault       = demoapp_item.vault.id
    consul      = demoapp_item.consul.id
  }
}
