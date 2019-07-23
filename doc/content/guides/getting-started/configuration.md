---
title: "Configuration"
description: ""
weight: 2
---

The stack can be started for development without passing any configuration. However, there are a lot of things you can configure. In this guide we'll set some environment variables in `docker-compose.yml`. These environment variables will configure the stack as a development server on `localhost`. For setting up up a public server or for requesting certificates from an ACME provider such as Let's Encrypt, take a closer look at the comments in `docker-compose.yml`.

## Frequency plans

By default, frequency plans are fetched by the stack from a [public GitHub repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans).

## Command-line interface

The command-line interface has some built-in defaults, but you'll want to create a config file or set some environment variables to point it at your deployment.

The recommended way to configure the CLI is with a `.ttn-lw-cli.yml` in your `$XDG_CONFIG_HOME` or `$HOME` directory. You can also put the config file in a different location, and pass it to the CLI as `-c path/to/config.yml`. In this guide we will use the following configuration file:

```yml
oauth-server-address: https://localhost:8885/oauth

identity-server-grpc-address: localhost:8884
gateway-server-grpc-address: localhost:8884
network-server-grpc-address: localhost:8884
application-server-grpc-address: localhost:8884
join-server-grpc-address: localhost:8884

ca: /path/to/your/cert.pem

log:
  level: info
```
