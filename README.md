# Temporal Terraform Provider

**This is experimental!**

A [Terraform](https://terraform.io) Provider for [Temporal](https://temporal.io/).

This project is not affiliated with nor supported by Temporal Technologies, Inc.

----

### Current Status

**It is a Work-In-Progress that currently only round-trips Schedules**

Supported Temporal Features:

 * [ ] Schedule
 * [ ] Workflow
 * [ ] Args
 * [ ] Memo
 * [ ] SearchAttributes

----

### Example

Examples are in the [`examples`](./examples) directory.

```
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

output "test-desc" {
  value = data.temporal_schedule.test.desc
}
```

----

### Development

```
# spin up a dev temporal server for testing
temporal server start-dev
```

Reminder put this in `~/.terraformrc`:

```
provider_installation {
  dev_overrides {
    "neomantra/temporal" = "/Users/<username>/go/bin"
  }

  # For all other providers, install them directly from their origin provider
  # registries as normal. If you omit this, Terraform will _only_ use
  # the dev_overrides block, and so no other providers will be available.
  direct {}
}
```

----

### License

Authored by [Evan Wies](https://github.com/neomantra).

Copyright (c) 2023 Neomantra BV.  All rights reserved.

Released under the MIT License, see [`LICENSE`](./LICENSE).
