# The Things Stack Networking

The Things Stack uses a port per protocol, with a TLS counterpart when applicable. Ports can be shared by multiple services using the same protocol, i.e. gRPC services sharing management, data and events services.

## Port Allocations

| Purpose | Protocol | Authentication | Port | Port (TLS) |
| --- | --- | --- | --- | --- | 
| Gateway data | [Semtech Packet Forwarder](https://github.com/Lora-net/packet_forwarder/blob/master/PROTOCOL.TXT) | None | 1700 (UDP) | N/A |
| Gateway data | MQTT (V2) | API key, token | 1881 | 8881 |
| Gateway data | MQTT | API key, token | 1882 | 8882 |
| Application data, events | MQTT | API key, token | 1883 | 8883 |
| Management, data, events | gRPC | API key, token | 1884 | 8884 |
| Management | HTTP | API key, token | 1885 | 8885 |
| Backend Interfaces | HTTP | Custom | N/A | 8886 |
| Basic Station LNS | HTTP | Auth Token, Custom | 1887 | 8887 |
