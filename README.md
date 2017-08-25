# The Things Network Stack

[![Build Status](https://travis-ci.org/TheThingsNetwork/ttn.svg?branch=master)](https://travis-ci.org/TheThingsNetwork/ttn) [![Coverage Status](https://coveralls.io/repos/github/TheThingsNetwork/ttn/badge.svg?branch=master)](https://coveralls.io/github/TheThingsNetwork/ttn?branch=master) [![GoDoc](https://godoc.org/github.com/TheThingsNetwork/ttn?status.svg)](https://godoc.org/github.com/TheThingsNetwork/ttn)

The Things Network Stack is an open-source LoRaWAN network server suitable for large, global public networks as well as smaller private networks.

LoRaWAN is a MAC protocol for low-power wide area networks. It allows for large scale Internet of Things deployments where low-powered devices efficiently communicate with Internet-connected applications over long range wireless connections. 

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

## Installation

## Downloads

For **stable** versions, see the [Releases](https://github.com/TheThingsNetwork/ttn/releases) page on Github.

For the latest **master**, you can download pre-compiled binaries:

| **File Name** | **Operating System** | **Architecture** |
| [ttn-darwin-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-darwin-amd64.zip) | macOS | amd64 |
| [ttn-linux-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-linux-amd64.zip) | linux | amd64 |
| [ttn-linux-386.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-linux-386.zip) | linux | 386 |
| [ttn-linux-arm.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-linux-arm.zip) | linux | arm |
| [ttn-windows-amd64.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-windows-amd64.zip) | windows | amd64 |
| [ttn-windows-386.zip](https://ttnreleases.blob.core.windows.net/release/master/ttn-windows-386.zip) | windows | 386 |

## Getting Started

## Documentation

- General documentation can be found on [thethingsnetwork.org/docs](https://www.thethingsnetwork.org/docs/)
- Setting up a development environment is documented in [DEVELOPMENT.md](DEVELOPMENT.md)
- Our contributing guidelines are described in [CONTRIBUTING.md](CONTRIBUTING.md)
- Documentation for our Go code can be found on [godoc.org](https://godoc.org/github.com/TheThingsNetwork/ttn)
- Our [forums](https://www.thethingsnetwork.org/forum) contain a massive amount of information and has great search
- You can chat on [Slack](http://thethingsnetwork.slack.com), an invite can be requested from your [account page](https://account.thethingsnetwork.org)

## Support
