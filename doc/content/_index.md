---
title: "Home"
description: ""
weight: 1
draft: false
--- 

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
