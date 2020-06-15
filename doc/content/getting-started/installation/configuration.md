---
title: "Configuration"
description: ""
weight: 3
---

{{% tts %}} can be started for development without passing any configuration. However, there are a number of things you need to configure for a production deployment. In this guide we will configure {{% tts %}} as a private deployment on `thethings.example.com`. For that purpose, we are going to need two configuration files: `docker-compose.yml`, which defines the Docker services of {{% tts %}} and its dependencies, and `ttn-lw-stack-docker.yml`, which contains configuration specific to our {{% tts %}} deployment.

Create a new folder where your deployment files will be placed. This guide assumes the following directory hierarchy:

```bash
docker-compose.yml          # defines Docker services for running {{% tts %}}
config/
└── stack/
    └── ttn-lw-stack-docker.yml    # configuration file for {{% tts %}}
```

You can find the example [docker-compose.yml]({{% repo-file-url "raw" "docker-compose.yml" %}}) and [ttn-lw-stack-docker.yml]({{% repo-file-url "raw" "config/stack/ttn-lw-stack-docker.yml" %}}) files [in the Github repository of {{% tts %}}]({{% repo %}}).

We will configure these files for a production deployment on `thethings.example.com`.

## Databases

We need to configure an SQL database, in this guide we'll use a single instance of [CockroachDB](https://www.cockroachlabs.com/). Make sure to find a recent tag of the [cockroachdb/cockroach image on Docker Hub](https://hub.docker.com/r/cockroachdb/cockroach/tags) and update it in the `docker-compose.yml` file. Make sure that the `volumes` are set up correctly so that the database is persisted on your server's disk.

The simplest configuration for CockroachDB will look like this (remember to replace `<the tag>`):

```yaml
# file: docker-compose.yml
services:
  # ...
  cockroach:
    image: 'cockroachdb/cockroach:<the tag>'
    command: 'start --http-port 26256 --insecure'
    restart: 'unless-stopped'
    volumes:
      - './data/cockroach:/cockroach/cockroach-data'
  # ...
```

Read the CockroachDB documentation for setting up a highly available cluster.

We also need to configure [Redis](https://redis.io/). In this guide we'll use a single instance of Redis. Just as with the SQL database, find a recent tag of the [redis image on Docker Hub](https://hub.docker.com/_/redis?tab=tags) and update it in the `docker-compose.yml` file. Again, make sure that the `volumes` are set up correctly so that the datastore is persisted on your server's disk. Note that {{% tts %}} requires Redis version 5.0 or newer.

The simplest configuration for Redis will look like this (remember to replace `<the tag>`):

```yaml
# file: docker-compose.yml
services:
  # ...
  redis:
    image: 'redis:<the tag>'
    command: 'redis-server --appendonly yes'
    restart: 'unless-stopped'
    volumes:
      - './data/redis:/data'
  # ...
```

Read the Redis documentation for setting up a highly available cluster.

## {{% tts %}}

### Docker

Before we go into the details of {{% tts %}}'s configuration, we'll take a look at the basics. Below you see part the configuration of the `stack` service in the `docker-compose.yml` file. As with the databases, you need to find a recent tag of the [thethingsnetwork/lorawan-stack image on Docker Hub](https://hub.docker.com/r/thethingsnetwork/lorawan-stack/tags) and update the `docker-compose.yml` file with that.

### Entrypoint and dependencies.

We tell Docker Compose to use `ttn-lw-stack -c /config/ttn-lw-stack-docker.yml`, as the container entry point so that our configuration file is always loaded (more on the config file below). The default command is `start`, which starts {{% tts %}}.

With the `depends_on` section we tell Docker Compose that {{% tts %}} depends on CockroachDB and Redis. With this, Docker Compose will wait for CockroachDB and Redis to come online before starting {{% tts %}}.

### Volumes

Under the `volumes` section, we define volumes for files that need to be persisted on disk. There are stored blob files (such as profile pictures) and certificate files retrieved with ACME (if required). We also mount the local `./config/stack/` directory on the container under `/config`, so that {{% tts %}} can find our configuration file at `/config/ttn-lw-stack-docker.yml`.

> NOTE: If your `ttn-lw-stack-docker.yml` is in a directory other than `./config/stack`, you will need to change this volume accordingly.

### Ports

The `ports` section exposes {{% tts %}}'s ports to the world. Port `80` and `443` are mapped to the internal HTTP and HTTPS ports. The other ports have a direct mapping. If you don't need support for gateways and applications that don't support TLS, you can remove ports starting with `188`.

In the `environment` section, we configure the databases used by {{% tts %}}. We will set these to the CockroachDB and Redis instances that are defined in the `docker-compose.yml` above.

If you plan on using existing certificate files, then you will need to uncomment the two `secrets` sections below, by removing the `# ` prefixes.

```yaml
# file: docker-compose.yml
services:
  # ...
  stack:
    image: 'thethingsnetwork/lorawan-stack:<the tag>'
    entrypoint: 'ttn-lw-stack -c /config/ttn-lw-stack-docker.yml'
    command: 'start'
    restart: 'unless-stopped'
    depends_on:
      - 'cockroach'
      - 'redis'
    volumes:
      - './data/blob:/srv/ttn-lorawan/public/blob'
      - './config/stack:/config:ro'
      # If using Let's Encrypt
      - './acme:/var/lib/acme'
    ports:
      - '80:1885'
      - '443:8885'
      - '1881:1881'
      - '8881:8881'
      - '1882:1882'
      - '8882:8882'
      - '1883:1883'
      - '8883:8883'
      - '1884:1884'
      - '8884:8884'
      - '1887:1887'
      - '8887:8887'
      - '1700:1700/udp'
    environment:
      TTN_LW_BLOB_LOCAL_DIRECTORY: '/srv/ttn-lorawan/public/blob'
      TTN_LW_REDIS_ADDRESS: 'redis:6379'
      TTN_LW_IS_DATABASE_URI: 'postgres://root@cockroach:26257/ttn_lorawan?sslmode=disable'

    # If using (self) signed certificates
    # secrets:
    #   - cert.pem
    #   - key.pem

# If using (self) signed certificates
# secrets:
#   cert.pem:
#     file: ./cert.pem
#   key.pem:
#     file: ./key.pem

```

Next, we'll have a look at the configuration options for your private deployment. We'll set these options in the `ttn-lw-stack-docker.yml` file.

First is TLS with Let's Encrypt. Since we're deploying {{% tts %}} on
`thethings.example.com`, we configure it to only request certificates for that
host, and also to use it as the default host (see the `tls` section).

> NOTE: Make sure that you use the correct `tls` depending on whether you will be using Let's Encrypt or your own certificate files.

We also configure HTTP server keys for encrypting and verifying cookies, as well
as passwords for endpoints that you may want to keep for internal use (see the `http` section).

{{% tts %}} sends emails to users, so we need to configure how these are sent.
You can use Sendgrid or an SMTP server. If you skip setting up an email provider,
{{% tts %}} will print emails to the stack logs (see the `email` section).

Finally, we also need to configure the URLs for the Web UI and the secret used
by the console client (see the `console` section).

For more detailed configuration, refer to [the relevant documentation]({{< ref "/reference/configuration" >}}).

Below is an example `ttn-lw-stack-docker.yml` file:

```yaml
# file: config/stack/ttn-lw-stack.yaml

# Example ttn-lw-stack configuration file

# Identity Server configuration
is:
  # Email configuration for "thethings.example.com"
  email:
    sender-name: 'The Things Stack'
    sender-address: 'noreply@thethings.example.com'
    network:
      name: 'The Things Stack'
      console-url: 'https://thethings.example.com/console'
      identity-server-url: 'https://thethings.example.com/oauth'

  # Web UI configuration for "thethings.example.com":
  oauth:
    ui:
      canonical-url: 'https://thethings.example.com/oauth'
      is:
        base-url: 'https://thethings.example.com/api/v3'

# HTTP server configuration
http:
  cookie:
    # generate 32 bytes (openssl rand -hex 32)
    block-key: '0011223344556677001122334455667700112233445566770011223344556677'
    # generate 64 bytes (openssl rand -hex 64)
    hash-key: '00112233445566770011223344556677001122334455667700112233445566770011223344556677001122334455667700112233445566770011223344556677'
  metrics:
    password: 'metrics'               # choose a password
  pprof:
    password: 'pprof'                 # choose a password

# If using (self) signed certificates:
# tls:
#   source: file
#   root-ca: /run/secrets/cert.pem
#   certificate: /run/secrets/cert.pem
#   key: /run/secrets/key.pem

# If using Let's encrypt for "thethings.example.com"
tls:
  source: 'acme'
  acme:
    dir: '/var/lib/acme'
    email: 'you@thethings.example.com'
    hosts: ['thethings.example.com']
    default-host: 'thethings.example.com'

# If Gateway Server enabled, defaults for "thethings.example.com":
gs:
  mqtt:
    public-address: 'thethings.example.com:1882'
    public-tls-address: 'thethings.example.com:8882'
  mqtt-v2:
    public-address: 'thethings.example.com:1881'
    public-tls-address: 'thethings.example.com:8881'

# If Gateway Configuration Server enabled, defaults for "thethings.example.com":
gcs:
  basic-station:
    default:
      lns-uri: 'wss://thethings.example.com:8887'
  the-things-gateway:
    default:
      mqtt-server: 'mqtts://thethings.example.com:8881'

# Web UI configuration for "thethings.example.com":
console:
  ui:
    canonical-url: 'https://thethings.example.com/console'
    is:
      base-url: 'https://thethings.example.com/api/v3'
    gs:
      base-url: 'https://thethings.example.com/api/v3'
    ns:
      base-url: 'https://thethings.example.com/api/v3'
    as:
      base-url: 'https://thethings.example.com/api/v3'
    js:
      base-url: 'https://thethings.example.com/api/v3'
    qrg:
      base-url: 'https://thethings.example.com/api/v3'
    edtc:
      base-url: 'https://thethings.example.com/api/v3'

  oauth:
    authorize-url: 'https://thethings.example.com/oauth/authorize'
    logout-url: 'https://thethings.example.com/oauth/logout'
    token-url: 'https://thethings.example.com/oauth/token'
    client-id: 'console'
    client-secret: 'console'          # choose or generate a secret (*)
```

> NOTE: Make note of the client secret, as it will be needed again when initializing {{% tts %}}.
