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

// Namespace round-trip
resource "temporal_namespace" "hello" {
  name = "hello"
  lifecycle {
    prevent_destroy = true // Terraform Provider can't delete it anyway
  }
}

data "temporal_namespace" "hello" {
  name = temporal_namespace.hello.name
}

output "namespace_owner_email" {
  value = data.temporal_namespace.hello.owner_email
}


// Schedule round-trip
# resource "temporal_schedule" "test" {
#   id = "test-schedule"
#   # schedule {
#   #   crons = ["CRON_TZ=America/New_York 20 16 * * * *"]
#   # }
#   action = {
#     start_workflow = {
#       workflow   = "my-workflow"
#       task_queue = "my-task-queue"
#     }
#   }
# }

# data "temporal_schedule" "test" {
#   id = temporal_schedule.test.id
# }

# output "test-id" {
#   value = data.temporal_schedule.test.id
# }

# output "test-desc" {
#   value = data.temporal_schedule.test.desc
# }
