---
title: "Configuration"
description: ""
weight: 2
---

The Things Stack binary can be configured with many different options. Those options can be provided as command-line flags, environment variables or using a configuration file.

<!--more-->

## Configuration Sources

In this reference we will refer to configuration options by name. On this page we will show how the `console.ui.canonical-url` option can be configured.

### Command-line flags

Command-line flags have the highest priority and, as such, override other means of configuration (environment variable or file). This looks as follows:

```bash
$ ttn-lw-stack start console --console.ui.canonical-url "https://thethings.example.com/console"
```

### Environment variables

Environment variables for configuration options are very similar to the command-line flags, except that they are in uppercase, and all separators (`.` or `-`) are replaced by underscores (`_`). Environment variables are also prefixed with `TTN_LW_`. 

> In many cases you'll want to use a `.env` file that is loaded using the [`dotenv` command of direnv](https://direnv.net/man/direnv-stdlib.1.html) or the [`env_file` option of Docker Compose](https://docs.docker.com/compose/compose-file/#env_file). You can also `export` each environment variable, or run `export $(grep -v '^#' .env | xargs)` to export all variables in the `.env` file.

The option from the command-line example from above would look as follows with environment variables:

```bash
TTN_LW_CONSOLE_UI_CANONICAL_URL="https://thethings.example.com/console"
```

### Configuration files

You can also configure The Things Stack with a YAML configuration file. This is again similar to the command-line flags, except that each `.` represents a YAML node. This allows you to group related options together:

```yaml
console:
  ui:
    canonical-url: 'https://thethings.example.com/console'
    # other console UI options
  # other console options
```

You can specify the location of the YAML configuration file with the command-line flag `-c` or `--config`. If this flag is not present, The Things Stack will look for config files in the following locations:

- The current directory
- The user's home directory (as [determined by Go](https://golang.org/pkg/os/#UserHomeDir))
- The user's config directory (as [determined by Go](https://golang.org/pkg/os/#UserConfigDir))

You can run The Things Stack with the `--help` flag, and check the description of the `--config` flag for the exact locations that are being checked.

### Defaults

The Things Stack can be used for local testing purposes without any custom configuration.

## Printing the Current Configuration

You can see the current configuration with the `config` command of `ttn-lw-stack` or `ttn-lw-cli`. By default this will print the configuration as CLI flags. Use the `--env` or `--yml` flags to print the configuration as environment variables or as YAML.
