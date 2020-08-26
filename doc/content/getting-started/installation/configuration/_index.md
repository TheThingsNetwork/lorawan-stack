---
title: "Configuration"
description: ""
weight: 2
---

{{% tts %}} can be configured using command-line flags, environment variables, or configuration files. See the [Configuration Reference]({{< ref src="/reference/configuration" >}}) for more information about the configuration options.

In this guide, we will configure {{% tts %}} using a configuration file, with an example domain `thethings.example.com` and TLS certificates from Let's Encrypt.

To configure {{% tts %}}, we will use the configuration file `ttn-lw-stack-docker.yml`, which contains configuration specific to our {{% tts %}} deployment. When {{% tts %}} starts, it looks for `ttn-lw-stack-docker.yml` for a license key, hostname, and other configuration parameters.

Download the example `ttn-lw-stack-docker.yml` used in this guide [here](ttn-lw-stack-docker.yml).

To configure Docker, we also need a `docker-compose.yml`, which defines the Docker services of {{% tts %}} and its dependencies.

Download the example `docker-compose.yml` used in this guide [here](docker-compose.yml).

Create a new folder where your deployment files will be placed. This guide assumes the following directory hierarchy:

```bash
docker-compose.yml          # defines Docker services for running {{% tts %}}
config/
└── stack/
    └── ttn-lw-stack-docker.yml    # configuration file for {{% tts %}}
```

## Configure Docker

Docker runs an instance of {{% tts %}}, as well as an SQL database and a Redis database which {{% tts %}} depend on to store data.

We will configure Docker to run three services:

- {{% tts %}}
- An SQL database (CockroachDB and PostgreSQL are supported)
- Redis
 
### SQL Database

We need to configure an SQL database, so in this guide we'll use a single instance of [CockroachDB](https://www.cockroachlabs.com/). Make sure to find a recent tag of the [cockroachdb/cockroach image on Docker Hub](https://hub.docker.com/r/cockroachdb/cockroach/tags) and update it in the `docker-compose.yml` file. Make sure that the `volumes` are set up correctly so that the database is persisted on your server's disk.

The simplest configuration for CockroachDB will look like this (remember to replace `latest` with a version tag in production):

{{< highlight yaml "linenos=table,linenostart=5" >}}
{{< readfile path="/content/getting-started/installation/configuration/docker-compose.yml" from=5 to=13 >}}
{{< /highlight >}}

> NOTE: It also possible (and even preferred) to use a managed SQL database. In this case, you will need to update the [`is.database-uri` configuration option]({{< ref src="/reference/configuration/identity-server/#database-options" >}}) to point to the address of the managed database.

### Redis

We also need to configure [Redis](https://redis.io/). In this guide we'll use a single instance of Redis. Just as with the SQL database, find a recent tag of the [redis image on Docker Hub](https://hub.docker.com/_/redis?tab=tags) and update it in the `docker-compose.yml` file. Again, make sure that the `volumes` are set up correctly so that the datastore is persisted on your server's disk. Note that {{% tts %}} requires Redis version 5.0 or newer.

The simplest configuration for Redis will look like this (remember to replace `latest` with a version tag in production):

{{< highlight yaml "linenos=table,linenostart=28" >}}
{{< readfile path="/content/getting-started/installation/configuration/docker-compose.yml" from=28 to=35 >}}
{{< /highlight >}}

> NOTE: It also possible (and even preferred) to use a managed Redis database. In this case, you will need to update the [`redis.address` configuration option]({{< ref src="/reference/configuration/the-things-stack/#redis-options" >}}) to point to the address of the managed database.

### {{% tts %}}

We need to configure Docker to pull and run {{% tts %}}. Below you see part the configuration of the `stack` service in the `docker-compose.yml` file. As with the databases, you need to find a recent tag of the [thethingsindustries/lorawan-stack image on Docker Hub](https://hub.docker.com/r/thethingsnetwork/lorawan-stack/tags) and update the `docker-compose.yml` file with that.

#### Entrypoint and dependencies

We tell Docker Compose to use `ttn-lw-stack -c /config/ttn-lw-stack-docker.yml`, as the container entry point so that our configuration file `ttn-lw-stack-docker.yml` is always loaded (more on the config file below). The default command is `start`, which starts {{% tts %}}.

With the `depends_on` field we tell Docker Compose that {{% tts %}} depends on CockroachDB and Redis. With this, Docker Compose will wait for CockroachDB and Redis to come online before starting {{% tts %}}.

> NOTE: If using a managed SQL or Redis database, these can be removed from `depends_on` and the services do not need to be started in Docker.

#### Volumes

Under the `volumes` section, we define volumes for files that need to be persisted on disk. There are stored blob files (such as profile pictures) and certificate files retrieved with ACME (if required). We also mount the local `./config/stack/` directory on the container under `/config`, so that {{% tts %}} can find our configuration file at `/config/ttn-lw-stack-docker.yml`.

> NOTE: If your `ttn-lw-stack-docker.yml` is in a directory other than `./config/stack`, you will need to change this volume accordingly.

#### Ports

The `ports` section exposes {{% tts %}}'s ports to the world. Port `80` and `443` are mapped to the internal HTTP and HTTPS ports. The other ports have a direct mapping. If you don't need support for gateways and applications that don't support TLS, you can remove ports starting with `188`.

In the `environment` section, we configure the databases used by {{% tts %}}. We will set these to the CockroachDB and Redis instances that are defined in the `docker-compose.yml` above.

{{< highlight yaml "linenos=table,linenostart=37" >}}
{{< readfile path="/content/getting-started/installation/configuration/docker-compose.yml" from=37 to=83 >}}
{{< /highlight >}}

> NOTE: If using managed databased, the `environment` ports need to be changed to the ports of the managed databases.

## Configure {{% tts %}}

Once Docker starts {{% tts %}}, we need to specify configuration options for running {{% tts %}} in the `ttn-lw-stack-docker.yml` file. Let's have a look at the configuration options which are required.

### TLS

{{% tts %}} supports TLS with Let's Encrypt. Since we're deploying {{% tts %}} on
`thethings.example.com`, we configure it to only request certificates for that
host, and also to use it as the default host (see the [`tls` configuration reference]({{< ref src="/reference/configuration/the-things-stack" >}}) section).

> NOTE: Make sure that you use the correct `tls` depending on whether you will be using Let's Encrypt or your own certificate files.

### HTTP

We also configure HTTP server keys for encrypting and verifying cookies, as well
as passwords for endpoints that you may want to keep for internal use (see the `http` section).

### Email

{{% tts %}} sends emails to users, so we need to configure how these are sent.
You can use Sendgrid or an SMTP server. If you skip setting up an email provider,
{{% tts %}} will print emails to the stack logs (see the `email` section).

### Component URLs

Finally, we also need to configure the URLs for the Web UI and the secret used
by the console client (see the `console` section). These tell {{% tts %}} where all its components are accessible.

>NOTE: Failure to correctly configure component URLs is a common problem that will prevent the stack from starting. Be sure to replace all instances of `thethings.example.com` with your domain name!

Below is an example `ttn-lw-stack-docker.yml` file:

{{< highlight yaml "linenos=table" >}}
{{< readfile path="/content/getting-started/installation/configuration/ttn-lw-stack-docker.yml" >}}
{{< /highlight >}}

> NOTE: Make note of the client secret, as it will be needed again when initializing {{% tts %}}.
