---
title: "Configure & Start"
description: ""
weight: 2
draft: false
--- 

## <a name="configure">Configuration</a>
The stack can be started without passing any configuration. 
However, there are a lot of things you can configure. See [configuration documentation](../../hosting/config) for more information.

Refer to the [networking documentation](../../hosting/networking.md) for the endpoints and ports that the stack uses by default.

### <a name="frequencyplans">Frequency plans</a>

By default, frequency plans are fetched by the stack from a [public GitHub repository](https://github.com/TheThingsNetwork/lorawan-frequency-plans).
To configure a local directory in offline environments, see the [configuration documentation](config.md) for more information.

### <a name="cli-config">Command-line interface</a>

The CLI have a built-in configuration but you will likely need to change it so the CLI point to your own deployment.

By default the CLI look for the  `.ttn-lw-cli.yml` in your `$XDG_CONFIG_HOME` or `$HOME` directory.
You can specify a different file by using `-c path/to/config.yml` command flag.

In this guide we will use the following configuration:

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
> `ttn-lw-cli --help` to see all the possible configuration.

> `ttn-lw-cli config` to see the current configuration.


## <a name="running">Run</a>

For simplicity, we will use docker-compose to orchestrate the local deployment. If you did not already, install [Docker](https://docs.docker.com/install/).

To ensure a smooth experience we provide you a basic [docker-compose.yml]({{< reffile "docker-compose.yml" >}}).

Now, in a terminal, go to the folder where the `docker-compose.yml` is located.
Then enter the following commands to:

1. Pull the necessary docker images.
2. Initialize the database.
3. Create the first user `admin`.
4. Create the CLI OAuth client.
5. Start the stack.

```bash
$ docker-compose pull
$ docker-compose run --rm stack is-db init
$ docker-compose run --rm stack is-db create-admin-user \
  --id admin \
  --email admin@localhost
$ docker-compose run --rm stack is-db create-oauth-client \
  --id cli \
  --name "Command Line Interface" \
  --owner admin \
  --no-secret \
  --redirect-uri 'local-callback' \
  --redirect-uri 'code'
$ docker-compose up
```

