# The Things Network Stack for LoRaWAN

[![Build Status](https://travis-ci.com/TheThingsNetwork/lorawan-stack.svg?branch=master)](https://travis-ci.com/TheThingsNetwork/lorawan-stack) [![Coverage Status](https://coveralls.io/repos/github/TheThingsNetwork/lorawan-stack/badge.svg?branch=master)](https://coveralls.io/github/TheThingsNetwork/lorawan-stack?branch=master)

The Things Network Stack for LoRaWAN is an open-source LoRaWAN network stack suitable for large, global and geo-distributed public and private networks as well as smaller networks. The architecture follows the LoRaWAN Network Reference Model for standards compliancy and interoperability.

LoRaWAN is a protocol for low-power wide area networks. It allows for large scale Internet of Things deployments where low-powered devices efficiently communicate with Internet-connected applications over long range wireless connections.

## Features

- LoRaWAN Network Server
  - [x] Supports LoRaWAN 1.1
  - [x] Supports LoRaWAN 1.0, 1.0.1, 1.0.2 and 1.0.3
  - [x] Supports LoRaWAN Regional Parameters 1.0, 1.0.2 rev B, 1.0.3 rev A, 1.1 rev A and B
  - [x] Supports Class A devices
  - [ ] Supports Class B devices
  - [x] Supports Class C devices
  - [x] Supports OTAA devices
  - [x] Supports ABP devices
  - [x] Supports MAC Commands
  - [x] Supports Adaptive Data Rate
  - [ ] Implements LoRaWAN Back-end Interfaces 1.0
- LoRaWAN Application Server
  - [x] Payload conversion of well-known payload formats
  - [x] Payload conversion using custom JavaScript functions
  - [x] MQTT pub/sub API
  - [x] HTTP Webhooks API
  - [ ] Implements LoRaWAN Back-end Interfaces 1.0
- LoRaWAN Join Server
  - [x] Supports OTAA session key derivation
  - [x] Supports external crypto services
  - [ ] Implements LoRaWAN Back-end Interfaces 1.0
- OAuth 2.0 Identity Server
  - [x] User management
  - [x] Entity management
  - [x] ACLs
- GRPC APIs
- HTTP APIs
- Command-Line Interface
  - [x] Create account and login
  - [x] Application management and traffic
  - [x] End device management, status and traffic
  - [x] Gateway management and status
- Web Interface (Console)
  - [x] Create account and login
  - [ ] Application management and traffic
  - [ ] End device management, status and traffic
  - [ ] Gateway management, status and traffic

## Getting Started

You want to get started? Fantastic! Here's an [extensive Getting Started guide](./doc/gettingstarted.md).

The easiest way to set up a private network is with the provided [`docker-compose.yml`](docker-compose.yml).

0. Prerequisites
    - Make sure you have [Docker](https://docs.docker.com/install/#supported-platforms) and [Docker Compose](https://docs.docker.com/compose/install/) installed.
    - Make sure you have a TLS certificate and key ready. We'll expect a `cert.pem` and `key.pem`. For development, you can generate self-signed certificates with `make dev.certs`.
1. Pull the Docker images:
    ```sh
    docker-compose pull
    ```
2. Initialize the database:
    ```sh
    docker-compose run --rm stack is-db init
    ```
3. Create the `admin` user:
    ```sh
    docker-compose run --rm stack is-db create-admin-user \
      --id admin --email admin@example.com
    ```
    You can choose a different user ID if you want. If you do, make sure to use
    that for `owner` below.
4. Register the CLI as an OAuth client:
    ```sh
    docker-compose run --rm stack is-db create-oauth-client \
      --id cli --name "Command Line Interface" --no-secret \
      --owner admin \
      --redirect-uri 'http://localhost:11885/oauth/callback' \
      --redirect-uri 'code'
    ```
5. Register the Console as an OAuth client:
    ```sh
    docker-compose run --rm stack is-db create-oauth-client \
      --id console --name "Console" \
      --owner admin \
      --redirect-uri 'http://example.com:1885/console/oauth/callback' \
      --redirect-uri 'https://example.com:8885/console/oauth/callback' \
      --redirect-uri '/console/oauth/callback'
    ```
    Make sure to copy the value of the `secret` that is printed by this command.
6. Export the Console's OAuth client secret or replace it in your `docker-compose.yml`:
    ```sh
    export TTN_LW_CONSOLE_OAUTH_CLIENT_SECRET=<PASTE SECRET HERE>
    ```
7. Run the Stack:
    ```sh
    docker-compose up
    ```

[Read the full Getting Started guide here](./doc/gettingstarted.md)

## Documentation

- General documentation can be found at [thethingsnetwork.org/docs](https://www.thethingsnetwork.org/docs/)
- Learn how to contribute in [CONTRIBUTING.md](CONTRIBUTING.md)
- Setting up a development environment is documented in [DEVELOPMENT.md](DEVELOPMENT.md)
- Documentation for the Go code can be found at [godoc.org](https://godoc.org/go.thethings.network/lorawan-stack)

## Support

- The [forums](https://www.thethingsnetwork.org/forum) contain a massive amount of information and has great search
- You can chat on [Slack](http://thethingsnetwork.slack.com), an invite can be requested from your [account page](https://account.thethingsnetwork.org)
- Hosted solutions, as well as commercial support and consultancy are offered by [The Things Industries](https://www.thethingsindustries.com)
