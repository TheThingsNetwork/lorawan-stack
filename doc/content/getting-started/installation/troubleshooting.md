---
title: "Troubleshooting"
description: ""
---

This section contains help for common issues you may encounter while installing {{% tts %}}.

## Version in "docker-compose.yml" is unsupported

Our `docker-compose.yml` file uses [Compose file version 3.7](https://docs.docker.com/compose/compose-file/). If using a package manager to install Docker Compose, it is possible to install an old, unsupported version. See Docker's [installation instructions](https://docs.docker.com/compose/install/) to upgrade to a more recent version.

## Token Exchange Refused

This is likely an Oauth issue. Double check that you used the correct `client-secret` when you authorized the client in [Running the Stack]({{< relref "running-the-stack" >}}).

## Can't access the server

Ensure you have a DNS record pointing to your server's public IP address. See your domain registrar's help section for instructions, or [name.com's DNS guide](https://www.name.com/support/articles/205188538-Pointing-your-domain-to-hosting-with-A-records) for general information about pointing records to your IP address.
