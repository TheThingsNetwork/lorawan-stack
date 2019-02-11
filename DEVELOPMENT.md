# The Things Network Stack for LoRaWAN Development

The Things Network Stack components are primarily built in Go, while we use Node for web front-ends. It is assumed that you have decent knowledge and experience with these technologies. If you want to get more familiar with Go, we strongly recommend to take [A Tour of Go](https://tour.golang.org/).

## Development Environment

The Things Network's development environment heavily relies on [`make`](https://www.gnu.org/software/make/). Under the hood, `make` calls other tools such as `git`, `go`, `yarn` etc. Recent versions are supported; Node v10.x and Go v1.11.5. Let's first make sure you have `go`, `node` and `yarn`:

On macOS using [Homebrew](https://brew.sh):

```sh
brew install go node yarn
```

On Ubuntu (or on Windows [using the Windows Subsystem for Linux](https://www.microsoft.com/nl-NL/store/p/ubuntu/9nblggh4msv6?rtc=1)):

```sh
curl -sL https://deb.nodesource.com/setup_10.x | sudo -E bash -
sudo apt-get install -y build-essential nodejs

curl -sSL https://dl.google.com/go/go1.11.5.linux-amd64.tar.gz | sudo tar -xz -C /usr/local
sudo ln -s /usr/local/go/bin/* /usr/local/bin
```

### Cloning the repository

If you are unfamiliar with forking projects on GitHub or cloning them locally, please [see the GitHub documentation](https://help.github.com/articles/fork-a-repo/).

### Getting started

As most of the tasks will be managed by `make` we will first initialize the tooling. You might want to run this commands from time to time:

```sh
make init
```

For convenience, you can initialize the development databases with some defaults.

>Note: this requires [Docker Desktop](https://www.docker.com/products/docker-desktop).

```sh
make dev.stack.init
```

This starts a CockroachDB and Redis database in Docker containers, creates a database, migrates tables and creates a user `admin` with password `admin`.

### Managing the development databases

You can also use the following commands to start, stop and erase databases.

```bash
make dev.databases.start # Starts all databases in a Docker container
make dev.databases.stop  # Stops all databases

# The contents of the databases will be saved in .dev/data.

make dev.databases.erase # Stop all databases and erase storage.
```

#### CockroachDB

CockroachDB is a distributed SQL database that we use in the Identity Server.

You can use `make dev.databases.sql` to enter an SQL shell.

#### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use `make dev.databases.redis-cli` to enter a Redis-CLI shell.

### Testing

```sh
make test
```

### Building

There is a single binary for the server, `ttn-lw-stack`, as well as a binary for the command-line interface `ttn-lw-cli`. The single binary contains all components start one or multiple components. This allows you to run the stack with one command in simple deployment scenarios, as well as distributing micro-services for more advanced scenarios.

We provide binary releases for all supported platforms, including packages for various package managers at https://github.com/TheThingsNetwork/lorawan-stack/releases. We suggest you use the compiled packages we provide in production scenarios.

For development/testing purposes we suggest either running required binaries via `go run` (e.g. `go run ./cmd/ttn-lw-cli` from repository root for CLI), or using `go build` directly. Note, that frontend (if used) needs to be built manually via `make js.build` before `go build` or `go run` commands are run.

If you must, you can build all arifacts with the following command:

```
make clean build-all
```

>Note: You will at least need to have [`rpm`](http://rpm5.org/) and [`snapcraft`](https://snapcraft.io/) in your `PATH`.

This will compile the front-end in `public`, the binaries for all supported platforms, `deb`, `rpm` and Snapcraft packages, release archives in `dist`,  as well as Docker images.

>Note: The operating system and architecture represent the name of the directory in `dist` in which the binaries are placed.
>For example, the binaries for Darwin x64 (macOS) will be located at `dist/darwin_amd64`.

#### API

> Note: If you don't work on changes in the API you can skip this section.

Our APIs are defined in `.proto` files in the `api` folder. These files describe the messages and interfaces of the different components of the Stack. If this is the first time you hear the term "protocol buffers" you should probably read the [protocol buffers documentation](https://developers.google.com/protocol-buffers/docs/proto3) before you continue.

From the `.proto` files, we generate code using the `protoc` compiler. As we plan to compile to a number of different languages, we decided to put the compiler and its dependencies in a Docker image, so make sure you have [Docker](https://www.docker.com/) installed before you try to compile them.

The actual commands for compilation are handled by our Makefile, so the only thing you have to execute, is:

```sh
make protos.clean protos
```

#### Folder Structure

```
.
├── .editorconfig       configuration for your editor, see editorconfig.org
├── CODEOWNERS          maintainers of folders who are required to approve pull requests
├── CONTRIBUTING.md     guidelines for contributing: branching, commits, code style, etc.
├── DEVELOPMENT.md      guide for setting up your development environment
├── docker-compose.yml  deployment file (including databases) for Docker Compose
├── Dockerfile          formula for building Docker images
├── LICENSE             the license that explains what you're allowed to do with this code
├── Makefile            dev/test/build tooling
├── README.md           general information about this project
│   ...
├── api                 contains the protocol buffer definitions for our API
├── cmd                 contains the different binaries that form the TTN stack for LoRaWAN
│   ├── internal        contains internal files shared between the different binaries
│   │   ...
│   ├── ttn-lw-cli      the command-line-interface for the TTN stack for LoRaWAN
│   └── ttn-lw-stack    bundles the server binaries that form the TTN stack for LoRaWAN
├── config              configuration for our JavaScript SDK and frontend
├── doc                 detailed documentation on the workings of the TTN stack for LoRaWAN
├── pkg                 contains all libraries used in the TTN stack for LoRaWAN
│   ├── component       contains the base component; all other components extend this component
│   ├── config          package for configuration using config files, environment and CLI flags
│   ├── errors          package for rich errors that include metadata and cross API boundaries
│   ├── log             package for logging
│   ├── messages        contains non-proto messages (such as the messages that are sent over MQTT)
│   ├── metrics         package for metrics collection
│   ├── ttnpb           contains generated code from our protocol buffer definitions and some helper functions
│   ├── types           contains primitive types
│   └── ...
├── public              frontend code will be compiled to this folder - not added to git
├── release             binaries will be compiled to this folder - not added to git
└── sdk                 source code for our SDKs
    └── js              source code for our JavaScript SDK
```
