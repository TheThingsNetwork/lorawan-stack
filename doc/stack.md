# The Things Network Stack for LoRaWAN

## Setup

You can download the appropriate binary [here](../README.md#downloads).

### Dependencies

#### Certificates

By default, the Stack requires a `cert.pem` and `key.pem`, in order to to serve content over TLS.

+ You can disable use of TLS by passing `gs.mqtt.listen-tls=""`, `grpc.listen-tls=""` and `http.listen-tls=""` at execution.

+ To generate testing certificates, you can use the following command. This requires a [Go environment setup](../DEVELOPMENT.md#development-environment).

    ```bash
    $ go run $(go env GOROOT)/src/crypto/tls/generate_cert.go -ca -host localhost
    ```

    You can change the `localhost` value if you want to change the hostname the stack will be accessed at.

    + You can specify a custom path for the certificate and key using the `--tls.certificate` and `--tls.key` flags when executing the Stack. You can, for example, use the TLS certificate and key you requested using Let's Encrypt. To retrieve your certificates, we recommend using [Certbot](https://certbot.eff.org/) in [manual mode](https://certbot.eff.org/docs/using.html#getting-certificates-and-choosing-plugins) or [acmetool](https://hlandau.github.io/acme/userguide).

#### Databases

To run the Stack, you will also need to have started an instance of [CockroachDB](https://www.cockroachlabs.com/product/cockroachdb/) and of [Redis](https://redis.io/).

+ The Identity Server will try to connect by default to a Cockroach instance at `localhost:25267`, with the `root` username and without password. You can change the [connection URI](https://www.cockroachlabs.com/docs/v2.0/connection-parameters#connect-using-a-url) and parameters with the `--is.database-uri` flag.

    With Docker installed, you can start a container instance of Cockroach using the following command:

    ```bash
    docker run -d -p 127.0.0.1:26257:26257 -p 127.0.0.1:26256:26256 -v "./cockroach:/cockroach/cockroach-data" cockroachdb/cockroach:v2.0.3 start --http-port 26256 --insecure
    ```

    If you'd rather not use Docker, you can explore the [Quickstart section](https://github.com/cockroachdb/cockroach/#quickstart) of the Cockroach documentation to set it up.

+ The Stack will connect by default to a Redis instance at `localhost:6379`. You can change the connection parameters with the `--redis.address`, `--redis.database` and `--redis.namespace` flags.

    With Docker installed, you can start a container instance of Redis using the following command:

    ```bash
    docker run -d -p 127.0.0.1:6379:6379 -v "./redis:/data" redis:4.0-alpine redis-server --appendonly yes
    ```

    If you'd rather not use Docker, you can explore the [Redis documentation](https://redis.io/download) to set it up.

### Configuration

The Stack can be started without passing any [configuration](config.md). We however recommend paying attention to the following parameters:

+ `--http.cookie.hash-key` is a 32 or 64 bytes long parameter, and `--http.cookie.block-key` is a 32 bytes long parameter. Both are used for cookie secrets. They should be passed in a hexadecimal string form. If no value is passed, a random value will be generated at startup.

+ `--cluster.keys` is a set of 16, 24 or 32 bytes long hexadecimal keys, used to identify components within a cluster. If no value is passed, a random value will be generated at startup. The first one passed is used for outgoing RPC calls, and all of them can be used to accept incoming RPC calls.

You can refer to our [networking documentation](networking.md) for the default endpoints of the Stack.

#### Frequency plans

By default, frequency plans are fetched by the stack from the [`TheThingsNetwork/gateway-conf` repository](https://github.com/TheThingsNetwork/gateway-conf). To set a new source:

+ `--frequency-plans.url` allows you to serve frequency plans fetched from a HTTP server.

+ `--frequency-plans.directory` allows you to serve frequency plans from a local directory.

### Running the Stack

After having downloaded the Stack and having [prepared the dependencies](#dependencies), you can start it by executing the binary:

```bash
# On a Linux amd64 environment:
$ ./ttn-lw-linux-identity-server-amd64 init
$ ./ttn-lw-linux-stack-amd64 start
```

The `init` command runs database migrations and creates an admin account, with `admin`/`admin` as default logins.

#### Run with `docker-compose`

You can also run it using Docker, or container orchestration solutions. A testing [Docker Compose configuration](../docker-compose.yml) is available in the repository:

```bash
$ docker-compose run -e TTN_LW_IDENTITY_SERVER_IS_DATABASE_URI=postgres://root@cockroach:26257/is_development?sslmode=disable --entrypoint ttn-lw-identity-server --rm stack init
$ docker-compose up
```
