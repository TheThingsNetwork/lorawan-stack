---
title: "Start"
description: ""
weight: 3
--- 

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

