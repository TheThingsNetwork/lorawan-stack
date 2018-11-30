# The Things Network Stack for LoRaWAN

[![Build Status](https://travis-ci.org/TheThingsNetwork/ttn.svg?branch=master)](https://travis-ci.org/TheThingsNetwork/ttn) [![Coverage Status](https://coveralls.io/repos/github/TheThingsNetwork/ttn/badge.svg?branch=master)](https://coveralls.io/github/TheThingsNetwork/ttn?branch=master) [![GoDoc](https://godoc.org/go.thethings.network/lorawan-stack?status.svg)](https://godoc.org/go.thethings.network/lorawan-stack)

The Things Network Stack for LoRaWAN is an open-source LoRaWAN network server suitable for large, global public networks as well as smaller private networks.

LoRaWAN is a protocol for low-power wide area networks. It allows for large scale Internet of Things deployments where low-powered devices efficiently communicate with Internet-connected applications over long range wireless connections. 

## Features

- LoRaWAN 1.1 Network Server
  - [x] Support for Class A devices
  - [ ] Support for Class B devices
  - [x] Support for Class C devices
  - [x] Support for ABP devices
  - [x] Support for MAC Commands
  - [x] Support for Adaptive Data Rate
- LoRaWAN 1.1 Application Server
  - [x] Payload conversion of well-known payload formats
  - [x] Payload conversion using custom JavaScript functions
  - [x] MQTT Pub/Sub API
- LoRaWAN 1.1 Join Server
  - [x] Over The Air Activation
- LoRaWAN 1.0 compatibility
- OAuth 2.0 Identity Server
  - [x] User Management
  - [x] ACLs
- GRPC APIs
- HTTP APIs
- Web Interface
  - [x] Application management and traffic
  - [x] End device management, status and traffic
  - [x] Gateway management, status and traffic

## Installation

Version 3 of The Things Network Stack for LoRaWAN is still under heavy development. We currently recommend to use version 2 instead.

## Downloads

For **stable** versions, see the [Releases](https://github.com/TheThingsNetwork/lorawan-stack/releases) page on Github.

For the latest **master**, you can download pre-compiled binaries:

| **File Name** | **Operating System** | **Architecture** |
| ------------- | -------------------- | ---------------- |
| [ttn-lw-darwin-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-darwin-amd64.zip) | macOS | amd64 |
| [ttn-lw-linux-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-linux-amd64.zip) | linux | amd64 |
| [ttn-lw-linux-386.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-linux-386.zip) | linux | 386 |
| [ttn-lw-linux-arm.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-linux-arm.zip) | linux | arm |
| [ttn-lw-windows-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-windows-amd64.zip) | windows | amd64 |
| [ttn-lw-windows-386.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-lw-windows-386.zip) | windows | 386 |

## Private Network Setup

The simplest way to set up a private network is with our provided [`docker-compose.yml`](docker-compose.yml).

0. Prerequisites
    - Make sure you have [Docker](https://docs.docker.com/install/#supported-platforms) and [Docker Compose](https://docs.docker.com/compose/install/) installed.
    - Make sure you have a TLS certificate and key ready. We'll expect a `cert.pem` and `key.pem`. For development, you can generate self-signed certificates with `make dev.certs`.
1. Pull the Docker images:  
    ```sh
    docker-compose pull
    ```
2. Initialize the database:  
    ```sh
    docker-compose run --rm stack ttn-lw-identity-server db init
    ```
3. Create the `admin` user:
    ```sh
    docker-compose run --rm stack ttn-lw-identity-server create-admin-user \
      --id admin --email admin@example.com
    ```
    You can choose a different user ID if you want. If you do, make sure to use
    that for `owner` below.
4. Register the CLI as an OAuth client:  
    ```sh
    docker-compose run --rm stack ttn-lw-identity-server create-oauth-client \
      --id cli --name "Command Line Interface" --no-secret \
      --owner admin \
      --redirect-uri 'http://localhost:11885/oauth/callback'
    ```
5. Register the Console as an OAuth client:  
    ```sh
    docker-compose run --rm stack ttn-lw-identity-server create-oauth-client \ 
      --id console --name "Console" \
      --owner admin \
      --redirect-uri 'http://example.com:1885/console/oauth/callback' \
      --redirect-uri 'https://example.com:8885/console/oauth/callback'
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

## Documentation

- General documentation can be found on [thethingsnetwork.org/docs](https://www.thethingsnetwork.org/docs/)
- Learn how to contribute in [CONTRIBUTING.md](CONTRIBUTING.md)
- Setting up a development environment is documented in [DEVELOPMENT.md](DEVELOPMENT.md)
- Documentation for our Go code can be found on [godoc.org](https://godoc.org/go.thethings.network/lorawan-stack)

## Support

- Our [forums](https://www.thethingsnetwork.org/forum) contain a massive amount of information and has great search
- You can chat on [Slack](http://thethingsnetwork.slack.com), an invite can be requested from your [account page](https://account.thethingsnetwork.org)
- Hosted solutions, as well as commercial support and consultancy are offered by [The Things Industries](https://www.thethingsindustries.com)
