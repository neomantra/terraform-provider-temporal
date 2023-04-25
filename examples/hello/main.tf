# examples/hello/main.tf
# Simple Exercise of the Temporal Provider

terraform {
  required_providers {
    temporal = {
      source = "neomantra/temporal"
    }
  }
}

provider "temporal" {
  hostport  = "127.0.0.1:7233"
  namespace = "default"
}

// Schedule round-trip
resource "temporal_schedule" "test" {
  id = "test-schedule"
}

data "temporal_schedule" "test" {
  id = temporal_schedule.test.id
}

output "test-id" {
  value = data.temporal_schedule.test.id
}

output "test-desc" {
  value = data.temporal_schedule.test.desc
}
