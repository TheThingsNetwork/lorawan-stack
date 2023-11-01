### Internal Events API

The Console internal events API is designed as an alternative to the `Events.Stream` gRPC API for event stream interactions. It allows multiple subscriptions to be multiplexed over a singular [WebSocket](https://en.wikipedia.org/wiki/WebSocket) connection.

### Reasoning

The `Events.Stream` gRPC API is available to HTTP clients via [grpc-gateway](https://github.com/grpc-ecosystem/grpc-gateway). While translated to HTTP, it is visible as a long-polling request whose response body will contain the events as a series of JSON objects.

This approach is efficient in the context of [HTTP/2](https://en.wikipedia.org/wiki/HTTP/2) which supports multiplexing multiple requests over a singular TCP connection.

Unfortunately the connection between a browser and The Things Stack is susceptible to proxies. Corporate environments are generally equipped with such proxies, and in their presence the connections are downgraded to HTTP/1.1 semantics.

In HTTP/1.1 connections can be used for a singular request at a time - it is not possible to multiplex the requests over a singular connection, and only [keep-alive](https://en.wikipedia.org/wiki/HTTP_persistent_connection) connections are available.

This is problematic as browsers have builtin limits for the number of concurrent connections that singular windows may use. This leads to hard to debug issues which are hardly reproducible.

But, there is one silver lining - the connection limit _does not apply to WebSocket connections_. The internal events API is designed to deal with this limitation while providing an experience similar to the original `Events.Stream` gRPC API.

### Endpoint

The endpoint for the internal events API is `/api/v3/console/internal/events/`. Note that the trailing slash is not optional.

### Semantics

The protocol is [full-duplex](https://en.wikipedia.org/wiki/Duplex_(telecommunications)#Full_duplex) - the client side and server side may transmit messages at any time without waiting for a response from the other party.

The protocol is centered around subscriptions. Subscriptions are identified by an unsigned numerical ID, which is selected by the client.

A subscription is initiated by the client via a subscription request, which the server confirms either with a subscription response or an error response.

Following a successful subscription, the server may send at any time publication responses containing the subscription identifier and an event. The subscription identifier can be used on the client side in order to route the event to the appropriate component or view.

A subscription can be terminated via an unsubscribe request, which the server confirms either with an unsubscribe response or an error response.

The client can expect that no publication responses will follow an unsubscribe response, but it is recommended that subscription identifiers are not recycled within the same session.

Error responses can be expected when the request contents are invalid (lack of identifiers, or invalid identifiers), or the caller is not authorized to subscribe to the provided identifiers. It is also invalid to request a subscription with the same identifier as an existing subscription, or to unsubscribe using an identifier which is not subscribed.

Error response are provided as a debugging facility, and the errors are generally not fixable by the Console user.

A special case exists for situations in which the caller is no longer authorized to receive any events associated with the provided identifiers _after_ the subscription response has been sent. This can happen if the caller token has expired or the rights have been revoked while the stream is ongoing. In such situations the server will terminate the connection explicitly.

### Authentication and Authorization

The authentication for the internal API is similar to other APIs available in The Things Stack. Given a `Bearer` token `t`, the `Authorization` header should contain the value `Bearer t`.

Upon connecting, no authorization will take place - the endpoint only will check that the provided token is valid (i.e. exists and it is not expired).

The [standard WebSocket API](https://developer.mozilla.org/en-US/docs/Web/API/WebSockets_API) [does not support custom request headers](https://github.com/whatwg/websockets/issues/16). As a result of this limitation, the backend allows the Console to provide the token as a [protocol](https://developer.mozilla.org/en-US/docs/Web/API/WebSocket/WebSocket#parameters). Specifically, given a `Bearer` token `t`, the following protocols should be provided to the `WebSocket` constructor:

- `ttn.lorawan.v3.console.internal.events.v1`
- `ttn.lorawan.v3.header.authorization.bearer.t`

### Message Format

Both requests and responses sent over the WebSocket connection are JSON encoded. All messages are JSON objects and are required to contain at least the following two fields:

- `type`: a string whose value must be either `subscribe`, `unsubscribe`, `publish` or `error`.
- `id`: an unsigned integer which identifies the underlying subscription being served.

Each of the following subsections describes an individual message and the message direction (client to server or server to client).

#### `SubscribeRequest` [C -> S]

- `type`: `subscribe`
- `id`: the subscription identifier
- `identifiers`, `tail`, `after`, `names`: semantically the same fields as those of the `StreamEventsRequest` Protobuf message.

Example:

```json
{
  "type": "subscribe",
  "id": 1,
  "tail": 10,
  "identifiers": [
    {
      "application_ids": {
        "application_id": "app1"
      }
    }
  ]
}
```

#### `SubscribeResponse` [S -> C]

- `type`: `subscribe`
- `id`: the subscription identifier

Example:

```json
{
  "type": "subscribe",
  "id": 1
}
```

#### `UnsubscribeRequest` [C -> S]

- `type`: `unsubscribe`
- `id`: the subscription identifier

Example:

```json
{
  "type": "unsubscribe",
  "id": 1
}
```

#### `UnsubscribeResponse` [S -> C]

- `type`: `unsubscribe`
- `id`: the subscription identifier

Example:

```json
{
  "type": "unsubscribe",
  "id": 1
}
```

#### `PublishResponse` [S -> C]

- `type`: `publish`
- `id`: the subscription identifier
- `event`: an `Event` Protobuf message encoded as a JSON object

Example:

```json
{
  "type": "publish",
  "id": 1,
  "event": {
    "name": "as.up.data.forward",
    "time": "2023-10-26T16:27:14.103854Z",
    "identifiers": [
      {
        "device_ids": {
          "device_id": "eui-0000000000000003",
          "application_ids": {
            "application_id": "app1"
          }
        }
      }
    ],
    "context": {
      "tenant-id": "Cgl0aGV0aGluZ3M="
    },
    "visibility": {
      "rights": [
        "RIGHT_APPLICATION_TRAFFIC_READ"
      ]
    },
    "unique_id": "01HDPCZDSQ358JMHD4SC2BQAB8"
  }
}
```

#### ErrorResponse [S -> C]

- `type`: `error`
- `id`: the subscription identifier
- `error`: a `Status` Protobuf message encoded as a JSON object

Example:

```json
{
  "type": "error",
  "id": 1,
  "error": {
    "code": 6,
    "message": "error:pkg/console/internal/events/subscriptions:already_subscribed (already subscribed with ID `1`)",
    "details": [
      {
        "@type": "type.googleapis.com/ttn.lorawan.v3.ErrorDetails",
        "namespace": "pkg/console/internal/events/subscriptions",
        "name": "already_subscribed",
        "message_format": "already subscribed with ID `{id}`",
        "attributes": {
          "id": "1"
        },
        "correlation_id": "5da004b9f61f479aafe5bbcae4551e63",
        "code": 6
      }
    ]
  }
}
```
