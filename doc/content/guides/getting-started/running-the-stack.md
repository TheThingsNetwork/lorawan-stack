---
title: "Running {{% tts %}}"
description: ""
weight: 4
---

Now that all configuration is done, we're ready to initialize {{% tts %}} and start it. Open a terminal prompt in the same directory as your `docker-compose.yml` file.

## Initialization

The first time {{% tts %}} is started, it requires some initialization. We'll start by pulling the Docker images:

```bash
$ docker-compose pull
```

Next, we need to initialize the database of the Identity Server:

```bash
$ docker-compose run --rm stack is-db init
```

We'll now create an initial `admin` user. Make sure to give it a good password.

```bash
$ docker-compose run --rm stack is-db create-admin-user \
  --id admin \
  --email your@email.com
```

Then we'll register the command-line interface as an OAuth client:

```bash
$ docker-compose run --rm stack is-db create-oauth-client \
  --id cli \
  --name "Command Line Interface" \
  --owner admin \
  --no-secret \
  --redirect-uri "local-callback" \
  --redirect-uri "code"
```

We do the same for the console. For `--secret`, make sure to enter the same value as you set for `TTN_LW_CONSOLE_OAUTH_CLIENT_SECRET` in the [Configuration]({{< relref "configuration" >}}) step.

```bash
$ docker-compose run --rm stack is-db create-oauth-client \
  --id console \
  --name "Console" \
  --owner admin \
  --secret the secret you generated before \
  --redirect-uri "https://thethings.example.com/console/oauth/callback" \
  --redirect-uri "/console/oauth/callback"
```

## Running {{% tts %}}

Now it's time to start {{% tts %}}:

```bash
$ docker-compose up
```

This will start the stack and print logs to your terminal. You can also start the stack in detached mode by adding `-d` to the command above. In that case you can get logs with [`docker-compose logs`](https://docs.docker.com/compose/reference/logs/).

With {{% tts %}} up and running, it's time to connect gateways, create devices and work with streaming data. See [Console]({{< relref "console" >}}) or [Command-line Interface]({{< relref "cli" >}}) to proceed.
