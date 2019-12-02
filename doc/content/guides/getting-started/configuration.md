---
title: "Configuration"
description: ""
weight: 3
---

The Things Stack can be started for development without passing any configuration. However, there are a number of things you need to configure for a production deployment. In this guide we'll set environment variables in the `docker-compose.yml` file to configure The Things Stack as a private deployment on `thethings.example.com`.

## Databases

We need to configure an SQL database, in this guide we'll use a single instance of [CockroachDB](https://www.cockroachlabs.com/). Make sure to find a recent tag of the [cockroachdb/cockroach image on Docker Hub](https://hub.docker.com/r/cockroachdb/cockroach/tags) and update it in the `docker-compose.yml` file. Make sure that the `volumes` are set up correctly so that the database is persisted on your server's disk.

The simplest configuration for CockroachDB will look like this:

```yaml
cockroach:
  image: 'cockroachdb/cockroach:<the tag>'
  command: 'start --http-port 26256 --insecure'
  restart: 'unless-stopped'
  volumes:
    - './data/cockroach:/cockroach/cockroach-data'
```

Read the CockroachDB documentation for setting up a highly available cluster.

We also need to configure [Redis](https://redis.io/), in this guide we'll use a single instance of Redis. Just as with the SQL database, find a recent tag of the [redis image on Docker Hub](https://hub.docker.com/_/redis?tab=tags) and update it in the `docker-compose.yml` file. Again, make sure that the `volumes` are set up correctly so that the datastore is persisted on your server's disk.

The simplest configuration for Redis will look like this:

```yaml
redis:
  image: 'redis:<the tag>'
  command: 'redis-server --appendonly yes'
  restart: 'unless-stopped'
  volumes:
    - './data/redis:/data'
```

Read the Redis documentation for setting up a highly available cluster.

## The Things Stack

Before we go into the details of The Things Stack's configuration, we'll take a look at the basics. Below you see part the configuration of the `stack` service in the `docker-compose.yml` file. As with the databases, you need to find a recent tag of the [thethingsnetwork/lorawan-stack image on Docker Hub](https://hub.docker.com/r/thethingsnetwork/lorawan-stack/tags) and update the `docker-compose.yml` file with that. We tell Docker Compose to start the container with `ttn-lw-stack start`, we indicate that it depends on CockroachDB and Redis, and we configure a volume for storing blobs (such as profile pictures).

The `ports` section exposes The Things Stack's ports to the world. Port `80` and `443` are mapped to the internal HTTP and HTTPS ports. The other ports have a direct mapping. If you don't need support for gateways and applications that don't support TLS, you can remove ports starting with `188`.

```yaml
stack:
  image: 'thethingsnetwork/lorawan-stack:<the tag>'
  entrypoint: 'ttn-lw-stack'
  command: 'start'
  restart: 'unless-stopped'
  depends_on:
    - 'cockroach'
    - 'redis'
  volumes:
    - './acme:/var/lib/acme'
    - './data/blob:/srv/ttn-lorawan/public/blob'
  ports:
    - '80:1885'
    - '443:8885'
    - '1882:1882'
    - '8882:8882'
    - '1883:1883'
    - '8883:8883'
    - '1884:1884'
    - '8884:8884'
    - '1887:1887'
    - '8887:8887'
    - '1700:1700/udp'
  env_file: '.env'
```

Next, we'll have a look at the configuration options for your private deployment. We'll set these options in the `.env` file that is referenced by the `env_file` option of the `stack` service in `docker-compose.yml`.

First we'll make sure that The Things Stack uses the correct databases.

```bash
TTN_LW_IS_DATABASE_URI="postgres://root@cockroach:26257/ttn_lorawan?sslmode=disable"
TTN_LW_REDIS_ADDRESS="redis:6379"
```

Then we'll configure TLS with Let's Encrypt. Since we're deploying The Things Stack on `thethings.example.com`, we configure it to only request certificates for that host, and to also use it as the default host.

```bash
TTN_LW_TLS_SOURCE="acme"
TTN_LW_TLS_ACME_DIR="/var/lib/acme"
TTN_LW_TLS_ACME_EMAIL="your@email.com"
TTN_LW_TLS_ACME_HOSTS="thethings.example.com"
TTN_LW_TLS_ACME_DEFAULT_HOST="thethings.example.com"
```

Next, we'll configure the HTTP server with keys for encrypting and verifying cookies, and with passwords for endpoints that you may want to keep for internal use.

```bash
TTN_LW_HTTP_COOKIE_HASH_KEY=...  # generate 64 bytes (openssl rand -hex 64)
TTN_LW_HTTP_COOKIE_BLOCK_KEY=... # generate 32 bytes (openssl rand -hex 32)
TTN_LW_HTTP_METRICS_PASSWORD=... # choose a password
TTN_LW_HTTP_PPROF_PASSWORD=...   # choose a password
```

The Things Stack sends emails to users, so we need to configure how those are sent. 

```bash
TTN_LW_IS_EMAIL_SENDER_NAME="The Things Stack"
TTN_LW_IS_EMAIL_SENDER_ADDRESS="noreply@thethings.example.com"
TTN_LW_IS_EMAIL_NETWORK_CONSOLE_URL="https://thethings.example.com/console"
TTN_LW_IS_EMAIL_NETWORK_IDENTITY_SERVER_URL="https://thethings.example.com/oauth"
```

You can either use Sendgrid or an SMTP server. If you don't set up an email provider, The Things Stack will print emails to the server log.

```bash
TTN_LW_IS_EMAIL_PROVIDER="sendgrid"
TTN_LW_IS_EMAIL_SENDGRID_API_KEY=... # enter Sendgrid API key
```

or

```bash
TTN_LW_IS_EMAIL_PROVIDER="smtp"
TTN_LW_IS_EMAIL_SMTP_ADDRESS=...  # enter mail server address
TTN_LW_IS_EMAIL_SMTP_USERNAME=... # enter username
TTN_LW_IS_EMAIL_SMTP_PASSWORD=... # enter password
```

Finally, we need to configure the Web UI to use `thethings.example.com`.

```bash
TTN_LW_IS_OAUTH_UI_CANONICAL_URL="https://thethings.example.com/oauth"
TTN_LW_IS_OAUTH_UI_IS_BASE_URL="https://thethings.example.com/api/v3"

TTN_LW_CONSOLE_OAUTH_AUTHORIZE_URL="https://thethings.example.com/oauth/authorize"
TTN_LW_CONSOLE_OAUTH_TOKEN_URL="https://thethings.example.com/oauth/token"

TTN_LW_CONSOLE_UI_CANONICAL_URL="https://thethings.example.com/console"
TTN_LW_CONSOLE_UI_IS_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_GS_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_NS_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_AS_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_JS_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_EDTC_BASE_URL="https://thethings.example.com/api/v3"
TTN_LW_CONSOLE_UI_QRG_BASE_URL="https://thethings.example.com/api/v3"

TTN_LW_CONSOLE_OAUTH_CLIENT_ID="console"
TTN_LW_CONSOLE_OAUTH_CLIENT_SECRET=... # choose or generate a secret
```

You will need the `TTN_LW_CONSOLE_OAUTH_CLIENT_SECRET` again in a later step.
