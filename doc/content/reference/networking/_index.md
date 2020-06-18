---
title: "Networking"
description: ""
---

{{% tts %}} uses a port per protocol, with a TLS counterpart when applicable. Ports can be shared by multiple services using the same protocol, i.e. gRPC services sharing management, data and events services.

<!--more-->

## Port Allocations

The following table lists the default ports used.

| **Purpose** | **Protocol** | **Authentication** | **Port** | **Port (TLS)** |
| --- | --- | --- | --- | --- | 
| Gateway data | [Semtech Packet Forwarder](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT) | None | 1700 (UDP) | N/A |
| Gateway data | MQTT (V2) | API key, token | 1881 | 8881 |
| Gateway data | MQTT | API key, token | 1882 | 8882 |
| Application data, events | MQTT | API key, token | 1883 | 8883 |
| Management, data, events | gRPC | API key, token | 1884 | 8884 |
| Management | HTTP | API key, token | 1885 | 8885 |
| Backend Interfaces | HTTP | Custom | N/A | 8886 |
| Basic Station LNS | HTTP | Auth Token, Custom | 1887 | 8887 |

## Service Discovery

{{% tts %}} supports discovering services using DNS SRV records. This is useful when dialing a cluster only by host name; the supported services and target host name and port are discovered using DNS.

To support service discovery for your {{% tts %}} cluster, configure DNS SRV records with the following services and protocols:

| **Protocol** | **SRV Service** | **SRV Protocol** | **SRV Target** |
| --- | --- | --- | --- |
| gRPC | `ttn-v3-is-grpc` | `tcp` | Identity Server |
| gRPC | `ttn-v3-gs-grpc` | `tcp` | Gateway Server |
| Semtech Packet Forwarder | `ttn-v3-gs-udp` | `udp` | Gateway Server |
| MQTT (V2) | `ttn-v3-gs-mqttv2` | `tcp` | Gateway Server |
| MQTT | `ttn-v3-gs-mqtt` | `tcp` | Gateway Server |
| Basic Station LNS | `ttn-v3-gs-basicstationlns` | `tcp` | Gateway Server |
| gRPC | `ttn-v3-ns-grpc` | `tcp` | Network Server |
| gRPC | `ttn-v3-as-grpc` | `tcp` | Application Server |
| MQTT | `ttn-v3-as-mqtt` | `tcp` | Application Server |
| gRPC | `ttn-v3-js-grpc` | `tcp` | Join Server |
| gRPC | `ttn-v3-dtc-grpc` | `tcp` | Device Template Converter |
| gRPC | `ttn-v3-dcs-grpc` | `tcp` | Device Claiming Server |
| gRPC | `ttn-v3-gcs-grpc` | `tcp` | Gateway Configuration Server |
| gRPC | `ttn-v3-qrg-grpc` | `tcp` | QR Code Generator |

For port, use the ports as defined above, or any custom port that you configured for your {{% tts %}} cluster.

### Example: Link External Application Server

If you want to link an Application Server outside your cluster on the Network Server, you can configure the DNS SRV record for the Network Server as follows:

```
_ttn-v3-ns-grpc._tcp.example.com. 86400 IN SRV 0 5 8884 ns.example.com.
```

When you configure the application link with `example.com` as Network Server address and enable TLS, the Application Server will discover the Network Server on `ns.example.com` and port `8884`.
