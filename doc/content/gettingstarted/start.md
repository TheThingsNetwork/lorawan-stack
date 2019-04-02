---
title: "Configure & Start"
description: ""
weight: 2
draft: false
--- 

## <a name="running">Running the stack</a>

Download our [Docker Compose configuration]({{ .Site.Params.github_repository }}) example. The stack can be run without any configuration, but for the purpose of this guide
we provided you a basic one.

With the `docker-compose.yml` file in the directory of your terminal prompt, enter the following commands to initialize the database, create the first user `admin`, create the CLI OAuth client and start the stack:

* Download the necessary `images`:
```bash
$ docker-compose pull
```
{{< asciinema 6eF6gN7BMwETjwY3ABYEpuHWx >}}

* Initialize the stack:
```bash
$ docker-compose run --rm stack is-db init
```
{{< asciinema aoPjC2Xyt3d9LXN9oNyaQ2G1N >}}

* Create the admin account, use the password you want:
```bash
$ docker-compose run --rm stack is-db create-admin-user --id admin --email admin@localhost
```
{{< asciinema aDA2v7lEQAn1N5BqCP357vizB >}}

* Create the first oauth client:
```bash
$ docker-compose run --rm stack is-db create-oauth-client --id cli --name "Command Line Interface" --owner admin --no-secret --redirect-uri 'http://localhost:11885/oauth/callback' --redirect-uri '/oauth/code'
```
{{< asciinema FKL9KFBiauFSWjl2M22P8V4UE >}}

* Start the stack:
```bash
$ docker-compose up -d
```
{{< asciinema PxIst4QuC9dRhXFz0DoO1AtLF >}}

