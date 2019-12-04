---
title: "Generic Server Options"
description: ""
weight: 1
---

## Global Options

Under normal circumstances, only `info`, `warn` and `error` logs are printed to the console. For development, you may also want to see `debug` logs.

- `log.level`: The minimum level log messages must have to be shown (default "info")

## TLS Options

{{% tts %}} serves several endpoints using TLS. TLS certificates can come from different sources.

- `tls.source`: Source of the TLS certificate (file, acme, key-vault)

If `file` is specified as `tls.source`, the location of the certificate and key need to be configured.

- `tls.certificate`: Location of TLS certificate
- `tls.key`: Location of TLS private key

If `acme` is specified as `tls.source`, certificates will be requested from Let's Encrypt (the default `tls.acme.endpoint`) and stored in the given `tls.acme.dir`.

- `tls.acme.endpoint`: ACME endpoint
- `tls.acme.dir`: Location of ACME storage directory
- `tls.acme.email`: Email address to register with the ACME account
- `tls.acme.hosts`: Hosts to enable automatic certificates for
- `tls.acme.default-host`: Default host to assume for clients without SNI

If `key-vault` is specified as `tls.source`, the certificate with the given ID is loaded from the key vault.

- `tls.key-vault.id`: ID of the certificate

For client-side TLS, you may configure a Root CA and optionally disable verification of certificate chains.

- `tls.root-ca`: Location of TLS root CA certificate (optional)
- `tls.insecure-skip-verify`: Skip verification of certificate chains (insecure)

## gRPC Options

The `grpc` options configure how {{% tts %}} listens for gRPC connections. The format is `host:port`. When listening on TLS ports, it uses the global [TLS configuration]({{< ref "#tls-options" >}}).

- `grpc.listen`: Address for the TCP gRPC server to listen on
- `grpc.listen-tls`: Address for the TLS gRPC server to listen on

When running a cluster in a trusted network, you can allow sending credentials over insecure connections with the `allow-insecure-for-credentials` option:

- `grpc.allow-insecure-for-credentials`: Allow transmission of credentials over insecure transport

## HTTP Options

The `http` options configure how {{% tts %}} listens for HTTP connections. The format is `host:port`. When listening on TLS ports, it uses the global [TLS configuration]({{< ref "#tls-options" >}}).

- `http.listen`: Address for the HTTP server to listen on
- `http.listen-tls`: Address for the HTTPS server to listen on

{{% tts %}} uses secure cookies that are encrypted with a `block-key` and signed with a `hash-key`. In production deployments you'll want these to stay the same between restarts. The keys are encoded as hex.

- `http.cookie.block-key`: Key for cookie contents encryption (16, 24 or 32 bytes)
- `http.cookie.hash-key`: Key for cookie contents verification (32 or 64 bytes)

{{% tts %}} serves a number of internal endpoints for health, metrics and debugging. These will usually be disabled or password protected in production deployments.

- `http.health.enable`: Enable health check endpoint on HTTP server
- `http.health.password`: Password to protect health endpoint (username is health)
- `http.metrics.enable`: Enable metrics endpoint on HTTP server
- `http.metrics.password`: Password to protect metrics endpoint (username is metrics)
- `http.pprof.enable`: Enable pprof endpoint on HTTP server
- `http.pprof.password`: Password to protect pprof endpoint (username is pprof)

It is possible to redirect users to the canonical URL of a deployment. There are options to redirect to a given host, or redirect from HTTP to HTTPS.

- `http.redirect-to-host`: Redirect all requests to one host
- `http.redirect-to-tls`: Redirect HTTP requests to HTTPS

The HTTP server serves static files for the web UI. If these files are not in the standard location, you may need to change the search path.

- `http.static.mount`: Path on the server where static assets will be served
- `http.static.search-path`: List of paths for finding the directory to serve static assets from

## Interoperability Options

{{% tts %}} supports interoperability according to LoRaWAN Backend Interfaces specification. The following options are used to configure the server for this.

- `interop.listen-tls`: Address for the interop server to listen on

- `interop.sender-client-ca.source`: Source of the interop server sender client CAs configuration (static, directory, url, blob)

The `url` source loads interop server sender client CAs configuration from the given URL.

- `interop.sender-client-ca.url`

The `directory` source loads from the given directory.

- `interop.sender-client-ca.directory`

The `blob` source loads from the given path in a bucket. This requires the global [blob configuration]({{< ref "#blob-options" >}}).

- `interop.sender-client-ca.blob.bucket`: Bucket to use
- `interop.sender-client-ca.blob.path`: Path to use

## Redis Options

Redis is the main data store for the [Network Server]({{< relref "network-server.md" >}}), [Application Server]({{< relref "application-server.md" >}}) and [Join Server]({{< relref "join-server.md" >}}). Redis is also used by the [Identity Server]({{< relref "identity-server.md" >}}) for caching and can be used by the [events system]({{< ref "#events-options" >}}) for exchanging events between components.

Redis configuration options:

- `redis.password`: Password of the Redis server
- `redis.database`: Redis database to use
- `redis.namespace`: Namespace for Redis keys

If connecting to a single Redis instance:

- `redis.address`: Address of the Redis server

Or you can enable failover using [Redis Sentinel](https://redis.io/topics/sentinel):

- `redis.failover.enable`: Set to `true`
- `redis.failover.addresses`: List of addresses of the Redis Sentinel instances (required)
- `redis.failover.master-name`: Redis Sentinel master name (required)

## Blob Options

The `blob` options configure how {{% tts %}} reads or writes files such as pictures, the frequency plans repository or files required for Backend Interfaces interoperability. The `provider` field selects the provider that is used, and which other options are read.

- `blob.provider`: Blob store provider (local, aws, gcp) (default "local")

If the blob provider is `local`, you need to specify the directory to use.

- `blob.local.directory`: Local directory that holds the buckets (default "./public/blob")

If the blob provider is `aws`, you need to specify the S3 region, the access key ID and secret access key.

- `blob.aws.region`: S3 region
- `blob.aws.access-key-id`: Access key ID
- `blob.aws.secret-access-key`: Secret access key

If the blob provider is `gcp`, you can specify the credentials with either the credentials data, or with the path to the credentials file.

- `blob.gcp.credentials`: JSON data of the GCP credentials, if not using JSON file
- `blob.gcp.credentials-file`: Path to the GCP credentials JSON file

## Events Options

The `events` options configure how events are shared between components. When using a single instance of {{% tts %}}, the `internal` backend is the best option. If you need to communicate in a cluster, you can use the `redis` or `cloud` backend.

- `events.backend`: Backend to use for events (internal, redis, cloud) (default "internal")

When using the `redis` backend, the global [Redis configuration]({{< ref "#redis-options" >}}) is used. Alternatively, you may customize the Redis configuration that is used for events.

- `events.redis.address`: Address of the Redis server
- `events.redis.password`: Password of the Redis server
- `events.redis.database`: Redis database to use
- `events.redis.namespace`: Namespace for Redis keys

With the `cloud` backend, the configured publish and subscribe URLs are passed to [the Go CDK](https://gocloud.dev/howto/pubsub/).

- `events.cloud.publish-url`: URL for the topic to send events
- `events.cloud.subscribe-url`: URL for the subscription to receiving events

## Frequency Plans Options

The `frequency-plans` configuration is used by the [Gateway Server]({{< relref "gateway-server.md" >}}) and the [Network Server]({{< relref "network-server.md" >}}). It can load configuration from a number of sources.

- `frequency-plans.config-source`: Source of the frequency plans (static, directory, url, blob)

The `url` source loads frequency plans from the given URL. See the [lorawan-frequency-plans](https://github.com/TheThingsNetwork/lorawan-frequency-plans) repository for more information.

- `frequency-plans.url`

The `directory` source loads from the given directory.

- `frequency-plans.directory`

The `blob` source loads from the given path in a bucket. This requires the global [blob configuration]({{< ref "#blob-options" >}}).

- `frequency-plans.blob.bucket`: Bucket to use
- `frequency-plans.blob.path`: Path to use

## Cluster Options

The `cluster` options configure how {{% tts %}} communicates with other components in the cluster. These options do not need to be set when running a single instance of {{% tts %}}. The most important options are the ones to configure the addresses of the other components in the cluster.

- `cluster.identity-server`: Address for the Identity Server
- `cluster.gateway-server`: Address for the Gateway Server
- `cluster.network-server`: Address for the Network Server
- `cluster.application-server`: Address for the Application Server
- `cluster.join-server`: Address for the Join Server
- `cluster.crypto-server`: Address for the Crypto Server

The cluster keys are 128 bit, hex-encoded keys that cluster components use to authenticate to each other.

- `cluster.keys`: Keys used to communicate between components of the cluster. The first one will be used by the cluster to identify itself

It is possible to configure the cluster to use TLS or not. We recommend to enable TLS for production deployments.

- `cluster.tls`: Do cluster gRPC over TLS
