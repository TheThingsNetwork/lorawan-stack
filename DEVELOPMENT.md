# The Things Network Stack for LoRaWAN Development

## Development Environment

The Things Network's development environment heavily relies on [`make`](https://www.gnu.org/software/make/). Under the hood, `make` calls other tools such as `git`, `go`, `yarn` etc. Let's first make sure you have `go`, `node` and `yarn`:

On macOS using [Homebrew](https://brew.sh):

```sh
brew install go node yarn
```

On Ubuntu (or Ubuntu [using the Windows Subsystem for Linux](https://www.microsoft.com/nl-NL/store/p/ubuntu/9nblggh4msv6?rtc=1)):

```sh
curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list

curl -sS https://deb.nodesource.com/gpgkey/nodesource.gpg.key | sudo apt-key add -
echo "deb https://deb.nodesource.com/node_8.x xenial main" | sudo tee /etc/apt/sources.list.d/nodesource.list
echo "deb-src https://deb.nodesource.com/node_8.x xenial main" | sudo tee -a /etc/apt/sources.list.d/nodesource.list

sudo apt-get update
sudo apt-get install build-essential nodejs yarn

curl -sS https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz -o go1.8.3.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.8.3.linux-amd64.tar.gz
sudo ln -s /usr/local/go/bin/* /usr/local/bin
```

### Getting started with Go Development

We will first need a Go workspace. The Go workspace is a folder that contains the following sub-folders:

- `src` which contains all source files
- `pkg` which contains compiled package objects
- `bin` which contains binary executables

From now on this folder is referred to as `$GOPATH`. By default, Go assumes that it's in `$HOME/go`, but to be sure that everything works correctly, you can add the following to your profile (in `$HOME/.profile`):

```sh
export GOPATH="$(go env GOPATH)"
export PATH="$PATH:$GOPATH/bin"
```

Now that your Go development environment is ready, we strongly recommend to get familiar with Go by following the [Tour of Go](https://tour.golang.org/).

### External dependencies

We rely on a number of external dependencies such as databases. You can either install these on your machine or run them in [Docker](https://www.docker.com).

We provide easy start-up methods if you have Docker installed:

```bash
make dev.databases.start # Starts all databases in a Docker container
make dev.databases.stop # Stops all databases

# The content of the databases will be by default saved in dev/databases.
# If you want to erase the databases' content:
make dev.databases.erase
```

#### CockroachDB

CockroachDB is a distributed SQL database that we use in the identity server.

You can use it using the `make dev.databases.{start,stop,erase}` method, or use `make dev.cockroach.{start,stop,erase}` to only manipulate Cockroach. You can also use `make dev.cockroach.sql` to enter a SQL shell, and `make dev.cockroach.drop NAME=<dbname>` to drop a database:

```
$ make dev.cockroach.drop NAME=test-db
ttn  Dropping test-db Cockroach database
DROP DATABASE
```

#### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

You can use it using the `make dev.databases.{start,stop,erase}` method, or use `make dev.redis.{start,stop,erase}` to only manipulate Redis.

### Getting started with development of The Things Network Stack for LoRaWAN

Since version 3 of our network stack, we use a single repository for our open source network components. The repository should be cloned inside your Go workspace:

```sh
git clone https://github.com/TheThingsNetwork/lorawan-stack.git $GOPATH/src/go.thethings.network/lorawan-stack
```

All development is done in this directory.

```sh
cd $GOPATH/src/go.thethings.network/lorawan-stack
```

As most of the tasks will be managed by `make` we will first initialize the tooling. You might want to run this command from time to time:

```sh
make init
```

#### Folder Structure

```
.
├── CONTRIBUTING.md     guidelines for contributing: branching, commits, code style, etc.
├── DEVELOPMENT.md      guide for setting up your development environment
├── Dockerfile          formula for building Docker images
├── Gopkg.lock          dependency lock file managed by golang/dep
├── Gopkg.toml          dependency file managed by golang/dep
├── LICENSE             the license that explains what you're allowed to do with this code
├── Makefile            dev/test/build tooling
├── README.md           general information about this project
│   ...
├── api                 contains the protocol buffer definitions for our API
├── cmd                 contains the different binaries that form the TTN stack for LoRaWAN
│   ├── shared          contains configuration that is shared between the different binaries
│   │   ...
│   └── ttn-lw-stack    bundles the binaries that form the TTN stack for LoRaWAN
├── config
├── doc
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
├── release             binaries will be compiled to this folder - not added to git
└── vendor              dependencies managed by golang/dep - not added to git
```

#### Testing

```sh
make go.test
```

#### Building

There's one binary to be built: the `ttn-example` binary, which holds an example component for the stack.

To build it run:

```
make ttn-example
```

This will result in two files, `release/ttn-example-darwin-amd64` (suffix can differ based on your architecture and operating system) and
`release/ttn-example` which is a symlink to `release/ttn-example-darwin-amd64`. This symlink will always be there and will point the correct 
binary for your architecture and operating system. This way builds on different machines will always result in one runnable binary: `./release/ttn-example`.

#### Development builds

When developing, it is not necessary to re-build all packages every time you compile. You should instead be using dev builds. The dev build rule for `ttn-example` is 
called `ttn-example.dev`. This builds the `ttn-example` binary, but makes some assumptions which can be made only in development mode, to speed up builds.

Use this together with artifact caching to speed up builds dramatically. The `cache` rule pre-builds all relevant build artifacts and caches them so they can be used for
dev builds.

When developing you should therefore run:

```
make cache
make ttn-example.dev
# make edits...
make ttn-example.dev
# make some more edits...
make ttn-example.dev
# ...
```

#### API

> Note: If you don't work on changes in the API you can skip this section.

Our APIs are defined in `.proto` files in the `api` folder. These files describe the messages and interfaces of the different components of the Stack. If this is the first time you hear the term "protocol buffers" you should probably read the [protocol buffers documentation](https://developers.google.com/protocol-buffers/docs/proto3) before you continue.

From the `.proto` files, we generate code using the `protoc` compiler. As we plan to compile to a number of different languages, we decided to put the compiler and its dependencies in a Docker image, so make sure you have [Docker](https://www.docker.com/) installed before you try to compile them.

The actual commands for compilation are handled by our Makefile, so the only thing you have to execute, is:

```sh
make go.protos
```
