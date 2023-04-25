# Temporal Terraform Provider

**This is experimental!**

A [Terraform](https://terraform.io) Provider for [Temporal](https://temporal.io/).

This project is not affiliated with nor supported by Temporal Technologies, Inc.


###

```
# spin up a dev temporal server for testing
temporal server start-dev


```

### Development

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

### License

Authored by [Evan Wies](https://github.com/neomantra).

Copyright (c) 2023 Neomantra BV.

Released under the MIT License, see `LICENSE`.
