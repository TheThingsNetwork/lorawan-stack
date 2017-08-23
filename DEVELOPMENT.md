# The Things Network Stack Development

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

#### CockroachDB

CockroachDB is a distributed SQL database that we use in the identity server.

#### Redis

Redis is an in-memory data store that we use as a database for "hot" data.

### Getting started with development of The Things Network Stack

Since version 3 of our network stack, we use a single repository for our open source network components. The repository should be cloned inside your Go workspace:

```sh
git clone https://github.com/TheThingsNetwork/ttn.git $GOPATH/src/github.com/TheThingsNetwork/ttn
```

All development is done in this directory.

```sh
cd $GOPATH/src/github.com/TheThingsNetwork/ttn
```

As most of the tasks will be managed by `make` we will first initialize the tooling. You might want to run this command from time to time:

```sh
make init
```

#### Folder Structure

_TODO_

#### Testing

#### Building
